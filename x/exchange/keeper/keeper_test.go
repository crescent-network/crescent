package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

type mockOrderSource struct {
	name string
}

func (m mockOrderSource) Name() string {
	return m.name
}

func (mockOrderSource) GenerateOrders(ctx sdk.Context, market types.Market, createOrder types.CreateOrderFunc, opts types.GenerateOrdersOptions) {
}

func (mockOrderSource) AfterOrdersExecuted(ctx sdk.Context, market types.Market, results []types.TempOrder) {
}

func (s *KeeperTestSuite) TestSetOrderSources() {
	// Same source name
	s.Require().PanicsWithValue("duplicate order source name: a", func() {
		k := keeper.Keeper{}
		k.SetOrderSources(&mockOrderSource{"a"}, &mockOrderSource{"a"})
	})
	k := keeper.Keeper{}
	k.SetOrderSources(&mockOrderSource{"a"}, &mockOrderSource{"b"})
	// Already set
	s.Require().PanicsWithValue("cannot set order sources twice", func() {
		s.keeper.SetOrderSources(&mockOrderSource{"b"}, &mockOrderSource{"c"})
	})
}
