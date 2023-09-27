package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestSwapExactAmountIn_SingleMarket() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("5"))

	s.PlaceLimitOrder(
		market.Id, mmAddr, false, utils.ParseDec("5.05"), sdk.NewInt(200_000000), 24*time.Hour)
	s.PlaceLimitOrder(
		market.Id, mmAddr, false, utils.ParseDec("5.03"), sdk.NewInt(100_000000), 24*time.Hour)
	s.PlaceLimitOrder(
		market.Id, mmAddr, true, utils.ParseDec("4.98"), sdk.NewInt(100_000000), 24*time.Hour)
	s.PlaceLimitOrder(
		market.Id, mmAddr, true, utils.ParseDec("4.95"), sdk.NewInt(200_000000), 24*time.Hour)

	ordererAddr := s.FundedAccount(2, utils.ParseCoins("750_000000uusd,150_000000ucre"))
	ctx := s.Ctx

	s.Ctx, _ = ctx.CacheContext()
	output, _ := s.SwapExactAmountIn(
		ordererAddr, []uint64{market.Id},
		utils.ParseCoin("750_000000uusd"), utils.ParseCoin("140_000000ucre"), false)
	s.AssertEqual(utils.ParseCoin("148_464158ucre"), output)

	s.Ctx, _ = ctx.CacheContext()
	output, _ = s.SwapExactAmountIn(
		ordererAddr, []uint64{market.Id},
		utils.ParseCoin("150_000000ucre"), utils.ParseCoin("700_000000uusd"), false)
	s.AssertEqual(utils.ParseCoin("743263500uusd"), output)
}

func (s *KeeperTestSuite) TestSwapInsufficientLiquidity() {
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "ucre")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market1.Id, mmAddr, utils.ParseDec("5"))
	s.MakeLastPrice(market2.Id, mmAddr, utils.ParseDec("2"))

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

func (s *KeeperTestSuite) TestSwapExactAmountIn_MaxOrderPriceRatio() {
	mmAddr := s.FundedAccount(
		1, utils.ParseCoins("1000000_000000ucre,1000000_000000uatom,1000000_000000uusd,1000000_000000ufoo,1000000_000000stake"))

	// Assume that the price for imaginary token FOO is $2.
	creUsdMarket := s.CreateMarket("ucre", "uusd")
	creFooMarket := s.CreateMarket("ucre", "ufoo")
	atomUsdMarket := s.CreateMarket("uatom", "uusd")
	atomFooMarket := s.CreateMarket("uatom", "ufoo")

	// Set last price for first three markets.
	s.MakeLastPrice(creUsdMarket.Id, mmAddr, utils.ParseDec("5"))
	s.MakeLastPrice(creFooMarket.Id, mmAddr, utils.ParseDec("2.5"))
	s.MakeLastPrice(atomUsdMarket.Id, mmAddr, utils.ParseDec("10"))
	// atomFooMarket has no last price

	s.createLiquidity2(creUsdMarket.Id, mmAddr, utils.ParseDec("4.8"), utils.ParseDec("0.2"), sdk.NewInt(1_000000))
	s.createLiquidity2(creFooMarket.Id, mmAddr, utils.ParseDec("2.5"), utils.ParseDec("0.2"), sdk.NewInt(1_000000))
	s.createLiquidity2(atomUsdMarket.Id, mmAddr, utils.ParseDec("10"), utils.ParseDec("0.2"), sdk.NewInt(3_000000))
	s.createLiquidity2(atomFooMarket.Id, mmAddr, utils.ParseDec("5.2"), utils.ParseDec("0.2"), sdk.NewInt(3_000000))

	resp, err := s.querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
		Input:       "3000000ucre",
		OutputDenom: "uatom",
	})
	s.Require().NoError(err)
	s.Require().Equal([]uint64{creUsdMarket.Id, atomUsdMarket.Id}, resp.Routes)

	ordererAddr := s.FundedAccount(2, utils.ParseCoins("1000_000000ucre"))
	cacheCtx, _ := s.Ctx.CacheContext()
	_, _, err = s.keeper.SwapExactAmountIn(
		cacheCtx, ordererAddr, []uint64{creFooMarket.Id, atomFooMarket.Id},
		utils.ParseCoin("3_000000ucre"), utils.ParseCoin("0uatom"), false)
	s.Require().EqualError(err, "market 4 has no last price: invalid request")

	cacheCtx, _ = s.Ctx.CacheContext()
	_, _, err = s.keeper.SwapExactAmountIn(
		cacheCtx, ordererAddr, []uint64{creUsdMarket.Id, atomUsdMarket.Id},
		utils.ParseCoin("50_000000ucre"), utils.ParseCoin("0uatom"), false)
	// Since the price impact is limited to MaxOrderPriceRatio(10% by default),
	// cannot sell CRE fully.
	s.Require().EqualError(err, "in market 1; paid 30000000ucre < input 50000000ucre: not enough liquidity in the market")
}
