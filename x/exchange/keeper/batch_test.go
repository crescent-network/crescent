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
		market.Id, ordererAddr1, false, utils.ParseDec("1"), sdk.NewDec(5_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.001"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.002"), sdk.NewDec(10_000000), time.Hour)

	order := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.01"), sdk.NewDec(1_000000), time.Hour)

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

	s.NextBlock()

	s.AssertEqual(utils.ParseDec("1"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)
	order = s.keeper.MustGetOrder(s.Ctx, order.Id)
	s.AssertEqual(sdk.NewDec(6_000000), order.OpenQuantity)
}
