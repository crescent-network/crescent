package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestBatchOneUserOrderAndPool() {
	// Test the batch matching when there's a pool and only one user order.

	s.keeper.SetDefaultTickSpacing(s.Ctx, 10) // Choose small enough tick spacing
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	ordererAddr := s.FundedAccount(1, enoughCoins)
	// At this moment, the limit order will not be executed since there's no
	// liquidity in the pool thus the order will rest in the order book.
	// The order will be matched against the pool in the next block's batch matching.
	_, order, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("5.005"), sdk.NewInt(1_000000), time.Hour)
	s.AssertEqual(sdk.ZeroInt(), res.ExecutedQuantity)

	lpAddr := s.FundedAccount(2, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	s.NextBlock() // Run batch matching

	// Check there was a match.
	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	s.AssertEqual(utils.ParseDec("5.005"), *marketState.LastPrice)

	// Check that the limit order has been executed.
	order = s.App.ExchangeKeeper.MustGetOrder(s.Ctx, order.Id)
	s.AssertEqual(sdk.NewInt(526748), order.OpenQuantity)
}
