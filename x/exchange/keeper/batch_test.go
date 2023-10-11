package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestBatchMatchingEdgecase() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("1"))

	ordererAddr1 := s.FundedAccount(2, enoughCoins)
	ordererAddr2 := s.FundedAccount(3, enoughCoins)

	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1"), sdk.NewInt(5_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.001"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.002"), sdk.NewInt(10_000000), time.Hour)

	order := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.01"), sdk.NewInt(1_000000), time.Hour)

	// Order book looks like:
	//            | 1.010 | #
	// ########## | 1.002 |
	// ########## | 1.001 |
	//      ##### | 1.000 | ########## <-- last price

	// After batch matching, it should look like(phase 1 only):
	//            | 1.010 |
	// ########## | 1.002 |
	// ########## | 1.001 |
	//            | 1.000 | ######     <-- last price

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)
	order = s.keeper.MustGetOrder(s.Ctx, order.Id)
	s.AssertEqual(sdk.NewInt(6_000000), order.OpenQuantity)

	ev := s.getEventOrderFilled(3)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000uusd"), ev.Received)

	ev = s.getEventOrderFilled(6)
	s.AssertEqual(sdk.NewInt(4_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("4_000000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("3_988000ucre"), ev.Received)

	ev = s.getEventOrderFilled(7)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_000000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("9_97000ucre"), ev.Received)
}
