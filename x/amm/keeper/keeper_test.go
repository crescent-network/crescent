package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v5/app/testutil"
)

type KeeperTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
