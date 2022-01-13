//go:build norace
// +build norace

package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func TestQueryCmdTestSuite(t *testing.T) {
	suite.Run(t, new(QueryCmdTestSuite))
}
