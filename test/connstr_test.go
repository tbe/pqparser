package test

import (
	"github.com/go-pg/pg"
)

func (suite *ParserTestSuite) TestParseValidConnstr() {
	suite.testValid(map[string]*pg.Options{
		"host=host port=12345 user=uri-user password=secret dbname=db": {
			Addr:     "host:12345",
			Network:  "tcp",
			User:     "uri-user",
			Password: "secret",
			Database: "db",
		},
		"": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Database: suite.username,
		},
	})
}