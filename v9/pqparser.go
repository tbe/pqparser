package pqparser

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
)

// TODO: this is only a quick hack to support go-pg/v9
// The struct Options struct itself is compatible, but not casteble, so we have to return a v9 struct for users
// of the v9 version of go-pg
// The idea is to move the parser to a seperate internal package, and just do the version specific things in version specific
// packages

const (
	// TODO: we should find something better
	slashReplacement = "HereWasASlash"
)

// Parse a connection string or URI.
// The connection string is parsed as defined in the libpq documentation:
// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
//
// The described environment variables are used exactly the same. If a value does not
// exist, than the corresponding environment variable is taken if it exists.
func Parse(connstr string) (*pg.Options, error) {

	// check if this is a url
	if strings.HasPrefix(connstr, "postgresql://") || strings.HasPrefix(connstr, "postgres://") {
		return parseURI(connstr)
	}

	// not a valid uri, we handle it as key value
	return parseConnstr(connstr)
}

func parseURI(dsn string) (*pg.Options, error) {
	options := &pg.Options{}

	// remove the prefix, we will add a dummy later for the net/url parser
	connuri := strings.TrimPrefix(strings.TrimPrefix(dsn, "postgresql://"), "postgres://")
	// split at the first slash
	uriparts := strings.SplitN(connuri, "/", 2)

	var uripath string
	if len(uriparts) > 1 {
		uripath = uriparts[1]
	}

	var userpass string
	var netloc string
	// check if we have a user delimiter
	if strings.ContainsRune(uriparts[0], '@') {
		// split the user part away
		userhost := strings.SplitN(uriparts[0], "@", 2)
		userpass = userhost[0]
		netloc = userhost[1]
	} else {
		netloc = uriparts[0]
	}

	// there may be cases where there is no slash at all, or the ? separator before the slash
	if strings.ContainsRune(netloc, '?') {
		parts := strings.SplitN(netloc, "?", 2)
		netloc = parts[0]
		if len(uriparts) > 1 {
			uripath = "/" + uripath
		}
		uripath = "?" + parts[1] + uripath
	}

	// remove empty port declaration at the end of the Host
	netloc = strings.TrimSuffix(netloc, ":")

	// libpq states, that is uses RFC 3986 URIs, but this RFC does not allow %-escapes inside the host part
	// expect for ipv6 strings and the % itself. So, we replace %2f sequences in the host part with a placeholder
	// and move back to slashes later. The whole splitting above is just because if this ...
	netloc = strings.ReplaceAll(strings.ReplaceAll(netloc, "%f2", slashReplacement), "%2F", slashReplacement)
	// and now, rebuild the uri
	connuri = "postgresql://"
	if userpass != "" {
		connuri += userpass + "@"
	}
	connuri += netloc + "/" + uripath

	// now, parse the uri
	uri, err := url.Parse(connuri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %v", err)
	}

	// now, convert the URI to the final object

	options.User = uri.User.Username()
	options.Password, _ = uri.User.Password()
	options.Database = strings.Trim(uri.Path, "/")

	// flatten the query parameters
	parameters := make(map[string]string)
	queryparam, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI query: %v", err)
	}
	for key, values := range queryparam {
		if len(values) != 1 {
			return nil, fmt.Errorf("failed to parse URI: parameter %v has not exactly one value", key)
		}
		parameters[key] = values[0]
	}

	urihost := strings.ReplaceAll(uri.Host, slashReplacement, "/")
	if _, exists := parameters["port"]; !exists {
		if portstr := uri.Port(); portstr != "" {
			urihost = strings.TrimSuffix(urihost, ":"+portstr)
			parameters["port"] = portstr
		}
	}

	// set host and port if they not exists in the parameter list
	if _, exists := parameters["host"]; !exists && urihost != "" {
		parameters["host"] = urihost
	}

	if err := parseParameters(options, parameters); err != nil {
		return nil, err
	}

	return options, nil
}

func parseConnstr(connstr string) (*pg.Options, error) {
	// we use the CSV parser, at it does all the quote handling for us
	cr := csv.NewReader(strings.NewReader(connstr))
	cr.Comma = ' '
	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %v", err)
	}

	parameters := make(map[string]string)
	// we have only one row
	if len(records) > 0 {
		for _, keyvalue := range records[0] {
			// split the key value and
			parts := strings.SplitN(keyvalue, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("failed to parse connection string: missing value after %s", keyvalue)
			}
			parameters[parts[0]] = parts[1]
		}
	}

	options := &pg.Options{}
	if err := parseParameters(options, parameters); err != nil {
		return nil, err
	}
	return options, nil
}

func parseParameters(options *pg.Options, parameters map[string]string) error {
	var host string
	var port int
	var fallbackAppname string

	var clientCert string
	var clientKey string

	sslNeedsHostname := false
	hasSSLConfig := false
	tlsconfig := &tls.Config{}
	sslEnabled := false

	parseSSLMode := func(mode string) error {
		hasSSLConfig = true

		switch mode {
		case "allow", "prefer", "require", "verify-ca":
			// use InsecureSkipVerify for require, as the behavior of libpq is not implementable
			// verify-ca can not be implemented, so we skip verification here
			tlsconfig.InsecureSkipVerify = true
			sslEnabled = true
		case "verify-full":
			tlsconfig.InsecureSkipVerify = false
			// remember we have to add the hostname later
			sslNeedsHostname = true
			sslEnabled = true
		case "disable":
			// disable TLS config
			sslEnabled = false
		default:
			return fmt.Errorf("failed to parse sslmode: unsupported mode %s", mode)
		}
		return nil
	}

	for key, value := range parameters {
		if len(value) == 0 {
			return fmt.Errorf("missing value for parameter %v", key)
		}
		switch key {
		case "host", "hostaddr":
			host = value
		case "port":
			var err error
			port, err = strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("failed to parse URI port: %v", err)
			}
		case "dbname":
			options.Database = value
		case "user":
			options.User = value
		case "password":
			options.Password = value
		case "connect_timeout":
			timeout, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("failed to parse URI connect_timeout: %v", err)
			}
			options.DialTimeout = time.Duration(timeout) * time.Second
		case "client_encoding":
			// not supported
		case "options":
			// not supported yet, could use onConnect handler
		case "application_name":
			options.ApplicationName = value
		case "fallback_application_name":
			fallbackAppname = value
		case "keepalives", "keepalives_idle", "keepalives_interval", "keepalives_count":
			// not supported, not required
		case "tty":
			// legacy
		case "sslmode":
			if err := parseSSLMode(value); err != nil {
				return err
			}
		case "requiressl":
			// deprecated
			tlsconfig.InsecureSkipVerify = true
			hasSSLConfig = true
			sslEnabled = true
		case "sslcompression":
			// not supported
		case "sslcert":
			clientCert = value
		case "sslkey":
			clientKey = value
		case "sslrootcert":
			// TODO
		case "sslcrl":
			// TODO
		case "requirepeer":
			// not supported
		case "krbsrvname":
			// not supported yet
		case "gsslib":
			// not supported yet
		case "service":
			// not supported yet
		case "target_session_attrs":
			// not supported yet
		default:
			return fmt.Errorf("invalid parameters %s", key)
		}
	}

	// now, check the environment variables for all missing parameters
	if host == "" {
		host = os.Getenv("PGHOST")
	}

	if port == 0 {
		portstr := os.Getenv("PGPORT")
		if portstr != "" {
			var err error
			port, err = strconv.Atoi(portstr)
			if err != nil {
				return fmt.Errorf("failed to parse PGPORT: %v", err)
			}
		}
	}

	if options.Database == "" {
		options.Database = os.Getenv("PGDATABASE")
	}

	if options.User == "" {
		options.User = os.Getenv("PGUSER")
	}

	if options.Password == "" {
		options.Password = os.Getenv("PGPASSWORD")
	}

	if options.ApplicationName == "" {
		options.ApplicationName = os.Getenv("PGAPPNAME")
	}

	if !hasSSLConfig {

		if sslmode := os.Getenv("PGSSLMODE"); sslmode != "" {
			if err := parseSSLMode(sslmode); err != nil {
				return err
			}
		} else if sslrequire := os.Getenv("PGREQUIRESSL"); sslrequire != "" {
			// deprecated
			tlsconfig.InsecureSkipVerify = true
			hasSSLConfig = true
			sslEnabled = true
		}
	}

	if clientCert == "" {
		clientCert = os.Getenv("PGSSLCERT")
	}

	if clientKey == "" {
		clientKey = os.Getenv("PGSSLKEY")
	}
	/* TODO:
	/ PGSSLROOTCERT behaves the same as the sslrootcert connection parameter.
	/ PGSSLCRL behaves the same as the sslcrl connection parameter.
	*/

	if options.DialTimeout == 0 {
		if timeoutstr := os.Getenv("PGCONNECT_TIMEOUT"); timeoutstr != "" {
			timeout, err := strconv.Atoi(timeoutstr)
			if err != nil {
				return fmt.Errorf("failed to parse PGCONNECT_TIMEOUT: %v", err)
			}
			options.DialTimeout = time.Duration(timeout) * time.Second
		}
	}

	// build the addr
	isSocket := false
	options.Network = "tcp"
	if host == "" {
		// default would be the unix socket, but we don't know
		// where the socket is on the system, so we fall back to TCP
		host = "localhost"
	} else if strings.HasPrefix(host, "/") {
		options.Network = "unix"
		isSocket = true
	}
	if port == 0 {
		port = 5432
	}

	if isSocket {
		options.Addr = fmt.Sprintf("%s/.s.PGSQL.%d", host, port)
	} else {
		options.Addr = fmt.Sprintf("%s:%d", host, port)
	}

	// set the default user
	if options.User == "" {
		if sysuser, err := user.Current(); err == nil {
			options.User = sysuser.Username
		} else {
			options.User = "postgres"
		}
	}

	if options.Database == "" {
		options.Database = options.User
	}

	// check the appname and possible fallback
	if options.ApplicationName == "" {
		options.ApplicationName = fallbackAppname
	}

	// set the TLS context
	if sslEnabled {
		options.TLSConfig = tlsconfig
		if sslNeedsHostname {
			// if we have a socket connection, we can not validate the hostname
			// so we would have to set SSL to non-validating
			if isSocket {
				options.TLSConfig.InsecureSkipVerify = true
			} else {
				options.TLSConfig.ServerName = host
			}
		}

		// check if we have ssl private key and cert
		if clientKey != "" && clientCert != "" {
			if cert, err := tls.LoadX509KeyPair(clientCert, clientKey); err != nil {
				return fmt.Errorf("failed to load SSL Keypair: %v", err)
			} else {
				options.TLSConfig.Certificates = append(options.TLSConfig.Certificates, cert)
			}
		}
	}
	return nil
}
