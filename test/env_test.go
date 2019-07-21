package test

import (
	"os"

	"github.com/go-pg/pg"
)

func (suite *ParserTestSuite) TestParseEnv() {
	// TODO: usefull testing for this
	envpw := os.Getenv("PGPASS")
	if envpw == "" {
		return
	}
	suite.testValid(map[string]*pg.Options{
		"": {
			Addr:     "localhost:5432",
			Network:  "tcp",
			User:     suite.username,
			Password: "mypassword",
			Database: suite.username,
		},
	})
}