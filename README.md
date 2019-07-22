# pqparser

A helper library to support libpq style connection strings for go-pg

## Why?

[go-pg][1] does have a very limited support for URI style DSN's. The goal of this library is,
to support the full feature set of [libpq][2]'s connection strings.

## Usage

```go
package main
import (
	"gitlab.com/tbe/pqparser"
	"github.com/go-pg/pg"
)	
func main() {
	options,err := pqparser.Parse("postgres://%2Fvar%2Flib%2Fpostgresql/dbname")
	if err != nil {
		panic(err)
	}
	
	db := pg.Connect(options)
	...
}
```

For full documentation about the behavior of `Parse`, see the [`libpq` documenation][2]

## Limitations

In difference to `libpq`, we use a TCP connection to `localhost` if no hostname is given.
`libpq` has a default socket path, but we don't know where the socket could be found.

Currently, there are a few features that are not implemented.

### `keepalive*` settings

As go-pq does not have support for keepalive settings, we ignore these parameters.

### `sslcompression`

I'm not sure if it is possible to implement this without dark hacks.

### `requirepeer`

This may be implemented later. We could implement this by overriding the `Dialer` function.

### `krbsrvname` and `gsslib`

There is no support for this authentication methods in go-pg.

### `service`

Could be implemented later. PR's welcome

### `sslmode`

Expect `verify-full`, all sslmodes skip the host verification at the moment.

### `sslrootcert` and `sslcrl`

Should be simple to implement, but the work is not done yet

[1]: https://github.com/go-pg/pg
[2]: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
