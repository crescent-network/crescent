package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestOrderExpiration() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.ctx = s.ctx.WithBlockTime(utils.ParseTime("2022-03-01T12:00:00Z"))
	order := s.limitOrder(s.addr(1), pair.Id, types.OrderDirectionSell, utils.ParseDec("1.0"), sdk.NewInt(10000), 10*time.Second, true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	s.ctx = s.ctx.WithBlockTime(utils.ParseTime("2022-03-01T12:00:06Z"))
	liquidity.BeginBlocker(s.ctx, s.keeper)
	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found) // The order is not yet deleted.
	// A buy order comes in.
	s.limitOrder(s.addr(2), pair.Id, types.OrderDirectionBuy, utils.ParseDec("1.0"), sdk.NewInt(5000), 0, true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	s.ctx = s.ctx.WithBlockTime(utils.ParseTime("2022-03-01T12:00:12Z"))
	liquidity.BeginBlocker(s.ctx, s.keeper)
	order, found = s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusPartiallyMatched, order.Status)
	// Another buy order comes in, but this time the first order has been expired,
	// so there is no match.
	s.limitOrder(s.addr(3), pair.Id, types.OrderDirectionBuy, utils.ParseDec("1.0"), sdk.NewInt(5000), 0, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	order, _ = s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().Equal(types.OrderStatusExpired, order.Status)
	s.Require().True(intEq(sdk.NewInt(5000), order.OpenAmount))

	liquidity.BeginBlocker(s.ctx, s.keeper)
	_, found = s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().False(found) // The order is gone.
}
