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
	s.AssertEqual(utils.ParseDec("2.236067977499789697"), poolState.CurrentSqrtPrice)
	s.AssertEqual(sdk.ZeroInt(), poolState.CurrentLiquidity)
	s.Require().EqualValues(40000, poolState.CurrentTick)
	s.AssertEqual(sdk.DecCoins{}, poolState.FeeGrowthGlobal)
	s.AssertEqual(sdk.DecCoins{}, poolState.FarmingRewardsGrowthGlobal)
}

func (s *KeeperTestSuite) TestPoolOrders() {
	type order struct {
		price sdk.Dec
		qty   sdk.Dec
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
				{utils.ParseDec("4.9950"), utils.ParseDec("25006228.045359273458597011")},
				{utils.ParseDec("4.9900"), utils.ParseDec("25043815.694766946591052267")},
				{utils.ParseDec("4.9850"), utils.ParseDec("25081497.603931454430609096")},
				{utils.ParseDec("4.9800"), utils.ParseDec("25119274.104114621728236342")},
			},
			[]order{
				{utils.ParseDec("5.0050"), utils.ParseDec("24956259.314253166236473605")},
				{utils.ParseDec("5.0100"), utils.ParseDec("24918890.317138035518922385")},
				{utils.ParseDec("5.0150"), utils.ParseDec("24881614.486321922356604332")},
				{utils.ParseDec("5.0200"), utils.ParseDec("24844431.496940975997647047")},
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
				{utils.ParseDec("4.9750"), utils.ParseDec("25106679.687535992611509571")},
				{utils.ParseDec("4.9700"), utils.ParseDec("25144570.207463681941517428")},
				{utils.ParseDec("4.9650"), utils.ParseDec("25182556.129454879252944134")},
				{utils.ParseDec("4.9600"), utils.ParseDec("25220637.790137883245677765")},
			},
			[]order{
				{utils.ParseDec("5.0250"), utils.ParseDec("25055960.893621086799461016")},
				{utils.ParseDec("5.0300"), utils.ParseDec("25018591.820577420876682247")},
				{utils.ParseDec("5.0350"), utils.ParseDec("24981315.543854639280576018")},
				{utils.ParseDec("5.0400"), utils.ParseDec("24944131.741163331026323419")},
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
				{utils.ParseDec("4.9950"), utils.ParseDec("16662453.996505449171218157")},
				{utils.ParseDec("4.9900"), utils.ParseDec("16687499.856199124957359553")},
				{utils.ParseDec("4.9850"), utils.ParseDec("66850484.536911765050696598")},
				{utils.ParseDec("4.9800"), utils.ParseDec("66951171.401038954332729693")},
				{utils.ParseDec("4.9750"), utils.ParseDec("16763015.169045484918964269")},
				{utils.ParseDec("4.9700"), utils.ParseDec("16788313.590350720122141501")},
			},
			[]order{
				{utils.ParseDec("5.0050"), utils.ParseDec("16629158.223875967012562200")},
				{utils.ParseDec("5.0100"), utils.ParseDec("16604258.059237103564291374")},
				{utils.ParseDec("5.0150"), utils.ParseDec("66616807.814489223777976688")},
				{utils.ParseDec("5.0200"), utils.ParseDec("66517255.912062619753483793")},
				{utils.ParseDec("5.0250"), utils.ParseDec("16529929.178629755600051203")},
				{utils.ParseDec("5.0300"), utils.ParseDec("16505276.037865950359873963")},
			},
		},
	} {
		s.Run(tc.name, func() {
			s.SetupTest()
			_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
			lpAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
			tc.addLiquidity(pool, lpAddr)
			var buyOrders, sellOrders []order
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, pool, true, func(price, qty, openQty sdk.Dec) (stop bool) {
				buyOrders = append(buyOrders, order{price, openQty})
				return false
			})
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, pool, false, func(price, qty, openQty sdk.Dec) (stop bool) {
				sellOrders = append(sellOrders, order{price, openQty})
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
	pool.MinOrderQuantity = sdk.NewDec(100)
	s.keeper.SetPool(s.Ctx, pool)

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.0001"), utils.ParseDec("10000"),
		utils.ParseCoins("10000ucre,50000uusd"))

	buyObs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:      true,
		PriceLimit: utils.ParseDecP("4.995"),
	}, nil)
	s.Require().Empty(buyObs.Levels())
	sellObs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:      false,
		PriceLimit: utils.ParseDecP("5.005"),
	}, nil)
	s.Require().Empty(sellObs.Levels())
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
		market.Id, ordererAddr, false, utils.ParseDec("45.821"), utils.ParseDec("39636169.911478604318885683"), time.Hour)

	s.SwapExactAmountIn(ordererAddr, []uint64{market.Id}, utils.ParseDecCoin("35987097uusd"), utils.ParseDecCoin("0ucre"), false)
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
	}, nil)
	s.Require().Len(obs.Levels(), 1)
	obs = s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
		IsBuy:             true,
		MaxNumPriceLevels: 1,
	}, nil)
	s.Require().Len(obs.Levels(), 1)
}

func (s *KeeperTestSuite) TestPoolOrderMaxOrderPriceRatio() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("5"))

	// last price != pool price
	pool := s.CreatePool(market.Id, utils.ParseDec("100"))

	s.AddLiquidityByLiquidity(
		mmAddr, pool.Id, utils.ParseDec("50"), utils.ParseDec("200"),
		sdk.NewInt(1000000))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("5.05"), sdk.NewDec(100_000000), 0)

	s.AssertEqual(utils.ParseDec("9.486832980505137996"), s.keeper.MustGetPoolState(s.Ctx, pool.Id).CurrentSqrtPrice)
	s.AssertEqual(utils.ParseDec("90"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)
}
