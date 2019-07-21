package test

import (
	"os/user"
	"testing"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/suite"

	"gitlab.com/tbe/pqparser"
)

type ParserTestSuite struct {
	suite.Suite
	username string
}

func (suite *ParserTestSuite) SetupTest() {
	if sysuser, err := user.Current(); err == nil {
		suite.username = sysuser.Username
	} else {
		suite.username = "postgres"
	}

}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (suite *ParserTestSuite) testValid(testcases map[string]*pg.Options) {
	for uri, expected := range testcases {
		result, err := pqparser.Parse(uri)
		suite.Nil(err, uri)
		suite.Equal(expected, result, uri)
	}
}

func (suite *ParserTestSuite) testInvalid(testcases []string) {
	for _, uri := range testcases {
		_, err := pqparser.Parse(uri)
		suite.NotNil(err, uri)
	}
}
