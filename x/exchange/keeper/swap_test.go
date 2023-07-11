package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestSwapInsufficientLiquidity() {
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "ucre")

	mmAddr := s.FundedAccount(1, enoughCoins)
	// Place 10 buy orders in market1 starting from price 4.999
	for i := 0; i < 10; i++ {
		price := types.PriceAtTick(types.TickAtPrice(utils.ParseDec("4.999")) - int32(i*10))
		s.PlaceLimitOrder(market1.Id, mmAddr, true, price, sdk.NewInt(10_000000), time.Minute)
	}
	// Place 10 sell orders in market1 starting from price 5.001
	for i := 0; i < 10; i++ {
		price := types.PriceAtTick(types.TickAtPrice(utils.ParseDec("5.001")) + int32(i*10))
		s.PlaceLimitOrder(market1.Id, mmAddr, false, price, sdk.NewInt(10_000000), time.Minute)
	}
	// Place 10 buy orders in market2 starting from price 1.999
	for i := 0; i < 10; i++ {
		price := types.PriceAtTick(types.TickAtPrice(utils.ParseDec("1.999")) - int32(i*10))
		s.PlaceLimitOrder(market2.Id, mmAddr, true, price, sdk.NewInt(5_000000), time.Minute)
	}
	// Place 10 sell orders in market2 starting from price 2.001
	for i := 0; i < 10; i++ {
		price := types.PriceAtTick(types.TickAtPrice(utils.ParseDec("2.001")) + int32(i*10))
		s.PlaceLimitOrder(market2.Id, mmAddr, false, price, sdk.NewInt(5_000000), time.Minute)
	}

	ordererAddr := s.FundedAccount(2, enoughCoins)
	cacheCtx, _ := s.Ctx.CacheContext()
	_, _, err := s.keeper.SwapExactAmountIn(
		cacheCtx, ordererAddr, []uint64{market1.Id, market2.Id},
		utils.ParseCoin("600_000000uusd"), utils.ParseCoin("58_000000uatom"), false)
	s.Require().ErrorIs(err, types.ErrSwapNotEnoughLiquidity)
	cacheCtx, _ = s.Ctx.CacheContext()
	_, _, err = s.keeper.SwapExactAmountIn(
		cacheCtx, ordererAddr, []uint64{market1.Id, market2.Id},
		utils.ParseCoin("300_000000uusd"), utils.ParseCoin("28_000000uatom"), false)
	s.Require().NoError(err)
}
