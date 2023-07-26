package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var enoughCoins = utils.ParseCoins(
	"10000_000000000000000000ucre,10000_000000000000000000uatom,10000_000000000000000000uusd")

type KeeperTestSuite struct {
	testutil.TestSuite
	keeper  keeper.Keeper
	querier keeper.Querier
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.keeper = s.App.ExchangeKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supplies
}

func (s *KeeperTestSuite) TestSetOrderSources() {
	// Same source name
	s.Require().PanicsWithValue("duplicate order source name: a", func() {
		k := keeper.Keeper{}
		k.SetOrderSources(types.NewMockOrderSource("a"), types.NewMockOrderSource("a"))
	})
	k := keeper.Keeper{}
	k.SetOrderSources(types.NewMockOrderSource("a"), types.NewMockOrderSource("b"))
	// Already set
	s.Require().PanicsWithValue("cannot set order sources twice", func() {
		s.keeper.SetOrderSources(types.NewMockOrderSource("b"), types.NewMockOrderSource("c"))
	})
}
