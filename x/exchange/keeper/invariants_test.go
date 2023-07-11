package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
)

func (s *KeeperTestSuite) TestCanCancelOrderInvariant() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)

	// Cancelling order within the same block height is impossible.
	_, broken := keeper.CanCancelOrderInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	// Now the order is cancellable.
	s.NextBlock()
	_, broken = keeper.CanCancelOrderInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	// Move funds from escrow to another address.
	s.Require().NoError(
		s.App.BankKeeper.SendCoins(
			s.Ctx, market.MustGetEscrowAddress(), utils.TestAddress(2), utils.ParseCoins("10uusd")))

	_, broken = keeper.CanCancelOrderInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestOrderStateInvariant() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)

	_, order, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("5.1"), sdk.NewInt(10_000000), time.Hour)

	_, broken := keeper.OrderStateInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	origDeadline := order.Deadline
	order.Deadline = s.Ctx.BlockTime()
	s.keeper.SetOrder(s.Ctx, order)
	_, broken = keeper.OrderStateInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
	order.Deadline = origDeadline
	s.keeper.SetOrder(s.Ctx, order)

	order.RemainingDeposit = sdk.ZeroInt()
	s.keeper.SetOrder(s.Ctx, order)
	_, broken = keeper.OrderStateInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestOrderBookInvariant() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	_, order, _ := s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("4.99"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.01"), sdk.NewInt(5_000000), time.Hour)

	_, broken := keeper.OrderBookInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	s.keeper.DeleteOrderBookOrderIndex(s.Ctx, order)
	order.Price = utils.ParseDec("5.02")
	s.keeper.SetOrder(s.Ctx, order)
	s.keeper.SetOrderBookOrderIndex(s.Ctx, order)

	_, broken = keeper.OrderBookInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestOrderBookOrderInvariant() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	_, order, _ := s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("4.99"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.01"), sdk.NewInt(5_000000), time.Hour)

	_, broken := keeper.OrderBookOrderInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	s.keeper.DeleteOrder(s.Ctx, order)
	_, broken = keeper.OrderBookOrderInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
	s.keeper.SetOrder(s.Ctx, order)

	s.keeper.DeleteOrderBookOrderIndex(s.Ctx, order)
	_, broken = keeper.OrderBookOrderInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestNumMMOrdersInvariant() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)

	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.999"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.998"), sdk.NewInt(10_000000), time.Hour)

	_, broken := keeper.NumMMOrdersInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	numMMOrders, _ := s.keeper.GetNumMMOrders(s.Ctx, ordererAddr, market.Id)
	s.Require().EqualValues(3, numMMOrders)

	s.keeper.SetNumMMOrders(s.Ctx, ordererAddr, market.Id, numMMOrders-1)
	_, broken = keeper.NumMMOrdersInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}
