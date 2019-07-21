package test

import (
	"github.com/go-pg/pg"
)

func (suite *ParserTestSuite) TestParseValidURI() {
	suite.testValid(map[string]*pg.Options{
		"postgresql://uri-user:secret@host:12345/db": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Password: "secret",
			Database: "db",
		},
		"postgresql://uri-user@host:12345/db": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Database: "db",
		},
		"postgresql://uri-user@host/db": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "db",
		},
		"postgresql://host:12345/db": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     suite.username,
			Database: "db",
		},
		"postgresql://host/db": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: "db",
		},
		"postgresql://uri-user@host:12345/": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Database: "uri-user",
		},
		"postgresql://uri-user@host/": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "uri-user",
		},
		"postgresql://uri-user@": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "uri-user",
		},
		"postgresql://host:12345/": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://host:12345": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://host/": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://host": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://?hostaddr=127.0.0.1": {
			Addr:     "127.0.0.1:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},

		// TODO		"postgresql://example.com?hostaddr=63.1.2.4",
		// TODO		"postgresql://%68ost/",
		"postgresql://host/db?user=uri-user": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "db",
		},
		"postgresql://host/db?user=uri-user&port=12345": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Database: "db",
		},
		"postgresql://host/db?u%73er=someotheruser&port=12345": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "someotheruser",
			Database: "db",
		},

		"postgresql://host:12345?user=uri-user": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Database: "uri-user",
		},
		"postgresql://host?user=uri-user": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "uri-user",
		},
		"postgresql://host?": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://[::1]:12345/db": {
			Addr:     "[::1]:12345",
			Network:  "tcp",
			User:     suite.username,
			Database: "db",
		},
		"postgresql://[::1]/db": {
			Addr:     "[::1]:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: "db",
		},
		"postgresql://[2001:db8::1234]/": {
			Addr:     "[2001:db8::1234]:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://[200z:db8::1234]/": {
			Addr:     "[200z:db8::1234]:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgresql://[::1]": {
			Addr:     "[::1]:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres:///": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres:///db": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: "db",
		},
		"postgres://uri-user@/db": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     "uri-user",
			Database: "db",
		},
		"postgres://?host=/path/to/socket/dir": {
			Addr:     "/path/to/socket/dir/.s.PGSQL.5432",
			Network:  "unix",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://@host": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://host:/": {
			Addr:     "host:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://:12345/": {
			Addr:     "localhost:12345",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://otheruser@?host=/no/such/directory": {
			Addr:     "/no/such/directory/.s.PGSQL.5432",
			Network:  "unix",
			User:     "otheruser",
			Database: "otheruser",
		},
		"postgres://otheruser@/?host=/no/such/directory": {
			Addr:     "/no/such/directory/.s.PGSQL.5432",
			Network:  "unix",
			User:     "otheruser",
			Database: "otheruser",
		},
		"postgres://otheruser@:12345?host=/no/such/socket/path": {
			Addr:     "/no/such/socket/path/.s.PGSQL.12345",
			Network:  "unix",
			User:     "otheruser",
			Database: "otheruser",
		},
		"postgres://otheruser@:12345/db?host=/path/to/socket": {
			Addr:     "/path/to/socket/.s.PGSQL.12345",
			Network:  "unix",
			User:     "otheruser",
			Database: "db",
		},
		"postgres://:12345/db?host=/path/to/socket": {
			Addr:     "/path/to/socket/.s.PGSQL.12345",
			Network:  "unix",
			User:     suite.username,
			Database: "db",
		},
		"postgres://:12345?host=/path/to/socket": {
			Addr:     "/path/to/socket/.s.PGSQL.12345",
			Network:  "unix",
			User:     suite.username,
			Database: suite.username,
		},
		"postgres://%2Fvar%2Flib%2Fpostgresql/dbname": {
			Addr:     "/var/lib/postgresql/.s.PGSQL.5432",
			Network:  "unix",
			User:     suite.username,
			Database: "dbname",
		},
	})
}

func (suite *ParserTestSuite) TestParseInvalidURI() {
	suite.testInvalid([]string{
		"postgresql://host/db?u%7aer=someotheruser&port=12345",
		"postgresql://host?uzer=",
		"postgre://",
		"postgres://[::1",
		// TODO: "postgres://[]", should fail because of empty IPv6 addr
		"postgres://[::1]z",
		"postgresql://host?zzz",
		"postgresql://host?value1&value2",
		"postgresql://host?key=key=value",
		"postgres://host?dbname=%XXfoo",
		"postgresql://a%00b",
		"postgresql://%zz",
		"postgresql://%1",
		"postgresql://%",
	})
}
