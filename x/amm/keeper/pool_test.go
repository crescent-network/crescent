package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestCreatePool_WithoutMarket() {
	creatorAddr := s.FundedAccount(1, enoughCoins)
	_, err := s.keeper.CreatePool(s.Ctx, creatorAddr, 1, utils.ParseDec("1.2"))
	s.Require().EqualError(err, "market not found: not found")
}

func (s *KeeperTestSuite) TestCreatePool_MultiplePoolsPerMarket() {
	creatorAddr := s.FundedAccount(1, enoughCoins)
	market, _ := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1.2"))
	_, err := s.keeper.CreatePool(s.Ctx, creatorAddr, market.Id, utils.ParseDec("2.5"))
	s.Require().EqualError(err, "cannot create more than one pool per market: invalid request")
}

func (s *KeeperTestSuite) TestCreatePool_InsufficientFee() {
	s.keeper.SetPoolCreationFee(s.Ctx, utils.ParseCoins("100_000000ucre"))
	market := s.CreateMarket("ucre", "uusd")
	creatorAddr := utils.TestAddress(1)
	_, err := s.keeper.CreatePool(s.Ctx, creatorAddr, market.Id, utils.ParseDec("5"))
	s.Require().EqualError(err, "insufficient pool creation fee: 0ucre is smaller than 100000000ucre: insufficient funds")
}

func (s *KeeperTestSuite) TestCreatePool() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	pool2, found := s.keeper.GetPoolByMarket(s.Ctx, market.Id)
	s.Require().True(found)
	s.Require().Equal(pool, pool2)
	pool2, found = s.keeper.GetPoolByReserveAddress(s.Ctx, pool.MustGetReserveAddress())
	s.Require().True(found)
	s.Require().Equal(pool, pool2)
	// Check pool state
	poolState, found := s.keeper.GetPoolState(s.Ctx, pool.Id)
	s.Require().True(found)
	s.AssertEqual(utils.ParseBigDec("2.236067977499789696409173668731276236"), poolState.CurrentSqrtPrice)
	s.AssertEqual(sdk.ZeroInt(), poolState.CurrentLiquidity)
	s.Require().EqualValues(40000, poolState.CurrentTick)
	s.AssertEqual(sdk.DecCoins{}, poolState.FeeGrowthGlobal)
	s.AssertEqual(sdk.DecCoins{}, poolState.FarmingRewardsGrowthGlobal)
}

func (s *KeeperTestSuite) TestPoolOrders() {
	type order struct {
		price sdk.Dec
		qty   sdk.Int
	}
	for _, tc := range []struct {
		name         string
		addLiquidity func(pool types.Pool, lpAddr sdk.AccAddress)
		buyOrders    []order
		sellOrders   []order
	}{
		{
			"simple liquidity",
			func(pool types.Pool, lpAddr sdk.AccAddress) {
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("4.98"), utils.ParseDec("5.02"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
			},
			[]order{
				{utils.ParseDec("4.9950"), sdk.NewInt(24993721)},
				{utils.ParseDec("4.9900"), sdk.NewInt(25031278)},
				{utils.ParseDec("4.9850"), sdk.NewInt(25068928)},
				{utils.ParseDec("4.9800"), sdk.NewInt(25106673)},
			},
			[]order{
				{utils.ParseDec("5.0050"), sdk.NewInt(24956259)},
				{utils.ParseDec("5.0100"), sdk.NewInt(24918890)},
				{utils.ParseDec("5.0150"), sdk.NewInt(24881614)},
				{utils.ParseDec("5.0200"), sdk.NewInt(24844431)},
			},
		},
		{
			"valley",
			func(pool types.Pool, lpAddr sdk.AccAddress) {
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("4.96"), utils.ParseDec("4.98"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("5.02"), utils.ParseDec("5.04"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
			},
			[]order{
				{utils.ParseDec("4.9750"), sdk.NewInt(25094072)},
				{utils.ParseDec("4.9700"), sdk.NewInt(25131931)},
				{utils.ParseDec("4.9650"), sdk.NewInt(25169885)},
				{utils.ParseDec("4.9600"), sdk.NewInt(25207935)},
			},
			[]order{
				{utils.ParseDec("5.0250"), sdk.NewInt(25055960)},
				{utils.ParseDec("5.0300"), sdk.NewInt(25018591)},
				{utils.ParseDec("5.0350"), sdk.NewInt(24981315)},
				{utils.ParseDec("5.0400"), sdk.NewInt(24944131)},
			},
		},
		{
			"high valley",
			func(pool types.Pool, lpAddr sdk.AccAddress) {
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("4.97"), utils.ParseDec("5.03"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("4.98"), utils.ParseDec("4.99"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
				s.AddLiquidity(
					lpAddr, pool.Id, utils.ParseDec("5.01"), utils.ParseDec("5.02"),
					utils.ParseCoins("100_000000ucre,500_000000uusd"))
			},
			[]order{
				{utils.ParseDec("4.9950"), sdk.NewInt(16654120)},
				{utils.ParseDec("4.9900"), sdk.NewInt(16679145)},
				{utils.ParseDec("4.9850"), sdk.NewInt(66816983)},
				{utils.ParseDec("4.9800"), sdk.NewInt(66917586)},
				{utils.ParseDec("4.9750"), sdk.NewInt(16754597)},
				{utils.ParseDec("4.9700"), sdk.NewInt(16779875)},
			},
			[]order{
				{utils.ParseDec("5.0050"), sdk.NewInt(16629158)},
				{utils.ParseDec("5.0100"), sdk.NewInt(16604258)},
				{utils.ParseDec("5.0150"), sdk.NewInt(66616807)},
				{utils.ParseDec("5.0200"), sdk.NewInt(66517255)},
				{utils.ParseDec("5.0250"), sdk.NewInt(16529929)},
				{utils.ParseDec("5.0300"), sdk.NewInt(16505276)},
			},
		},
	} {
		s.Run(tc.name, func() {
			s.SetupTest()
			market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
			lpAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
			tc.addLiquidity(pool, lpAddr)
			var buyOrders, sellOrders []order
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, market, pool, true, func(price sdk.Dec, qty sdk.Int) (stop bool) {
				buyOrders = append(buyOrders, order{price, qty})
				return false
			})
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, market, pool, false, func(price sdk.Dec, qty sdk.Int) (stop bool) {
				sellOrders = append(sellOrders, order{price, qty})
				return false
			})
			s.Require().Len(buyOrders, len(tc.buyOrders))
			for i := range tc.buyOrders {
				s.AssertEqual(tc.buyOrders[i].price, buyOrders[i].price)
				s.AssertEqual(tc.buyOrders[i].qty, buyOrders[i].qty)
			}
			s.Require().Len(sellOrders, len(tc.sellOrders))
			for i := range tc.sellOrders {
				s.AssertEqual(tc.sellOrders[i].price, sellOrders[i].price)
				s.AssertEqual(tc.sellOrders[i].qty, sellOrders[i].qty)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPoolMinOrderQuantity() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	market.OrderQuantityLimits.Min = sdk.NewInt(100)
	s.App.ExchangeKeeper.SetMarket(s.Ctx, market)

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.0001"), utils.ParseDec("10000"),
		utils.ParseCoins("10000ucre,50000uusd"))

	buyObs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:      true,
		PriceLimit: utils.ParseDecP("4.995"),
	})
	s.Require().Empty(buyObs.Levels)
	sellObs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:      false,
		PriceLimit: utils.ParseDecP("5.005"),
	})
	s.Require().Empty(sellObs.Levels)
}

func (s *KeeperTestSuite) TestSwapEdgecase1() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("45.821"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, lpAddr, utils.ParseDec("45.821"))
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("1"), utils.ParseDec("100"),
		utils.ParseCoins("100000000ucre,1000000000uusd"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("45.821"), sdk.NewInt(39636169), time.Hour)

	s.SwapExactAmountIn(ordererAddr, []uint64{market.Id}, utils.ParseCoin("35987097uusd"), utils.ParseCoin("0ucre"), false)
}

func (s *KeeperTestSuite) TestPoolOrdersEdgecase() {
	// Check if there's no infinite loop in IteratePoolOrders.
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.000000089916180444"))
	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	marketState.LastPrice = utils.ParseDecP("0.000000089795000000")
	s.App.ExchangeKeeper.SetMarketState(s.Ctx, market.Id, marketState)

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, types.MinPrice, types.MaxPrice,
		utils.ParseCoins("13010176813853779ucre,1169825406uusd"))

	obs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:             false,
		MaxNumPriceLevels: 1,
	})
	s.Require().Len(obs.Levels, 1)
	obs = s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:             true,
		MaxNumPriceLevels: 1,
	})
	s.Require().Len(obs.Levels, 1)
}

func (s *KeeperTestSuite) TestPoolOrderMaxOrderPriceRatio() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("5"))

	// last price != pool price
	pool := s.CreatePool(market.Id, utils.ParseDec("100"))

	s.AddLiquidityByLiquidity(
		mmAddr, pool.Id, utils.ParseDec("50"), utils.ParseDec("200"),
		sdk.NewInt(100000000))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("5.05"), sdk.NewInt(100_000000), 0)

	s.AssertEqual(
		utils.ParseBigDec("9.486854161057564143941381866561324639"),
		s.keeper.MustGetPoolState(s.Ctx, pool.Id).CurrentSqrtPrice)
	s.AssertEqual(
		utils.ParseDec("90"),
		*s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)
}

func (s *KeeperTestSuite) TestPoolOrdersMatchingInterval() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("5.01"), utils.ParseDec("5.02"),
		utils.ParseCoins("10_000000ucre"))

	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("5.04"), utils.ParseDec("5.05"),
		utils.ParseCoins("10_000000ucre"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, _, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("5.1"), sdk.NewInt(20_000000), 0)

	s.Require().True(res.IsMatched())
	s.AssertEqual(sdk.NewInt(19_999998), res.ExecutedQuantity)
	s.AssertEqual(
		utils.ParseBigDec("2.247220504978800695084768613573562822"),
		s.keeper.MustGetPoolState(s.Ctx, pool.Id).CurrentSqrtPrice)
}
