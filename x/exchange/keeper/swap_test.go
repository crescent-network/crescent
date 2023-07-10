package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestSwapExactAmountIn() {
	creatorAddr := utils.TestAddress(0)
	s.FundAccount(creatorAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "ucre")

	pool1 := s.CreatePool(market1.Id, utils.ParseDec("5"))
	s.AddLiquidity(
		creatorAddr, creatorAddr, pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	pool2 := s.CreatePool(market2.Id, utils.ParseDec("2"))
	s.AddLiquidity(
		creatorAddr, creatorAddr, pool2.Id, utils.ParseDec("1.5"), utils.ParseDec("3"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))

	ordererAddr := utils.TestAddress(1)
	s.FundAccount(ordererAddr, utils.ParseCoins("10000_000000ucre,10000_000000uatom,10000_000000uusd"))

	routes := []uint64{market1.Id, market2.Id}
	input := sdk.NewInt64Coin("uusd", 100_000000)
	minOutput := sdk.NewInt64Coin("uatom", 9_000000)
	output, _ := s.SwapExactAmountIn(ordererAddr, routes, input, minOutput, false)
	s.Require().Equal("9874876uatom", output.String())
}

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
