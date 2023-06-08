package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
)

var enoughCoins = utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd")

type KeeperTestSuite struct {
	testutil.TestSuite
	keeper keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.keeper = s.App.ExchangeKeeper
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supplies
}
