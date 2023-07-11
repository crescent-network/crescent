package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
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
	s.Require().EqualError(err, "0ucre is smaller than 100000000ucre: insufficient funds")
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
	s.Require().Equal("5.000000000000000000", poolState.CurrentPrice.String())
	s.Require().Equal(sdk.ZeroInt(), poolState.CurrentLiquidity)
	s.Require().EqualValues(40000, poolState.CurrentTick)
	s.Require().Equal("", poolState.FeeGrowthGlobal.String())
	s.Require().Equal("", poolState.FarmingRewardsGrowthGlobal.String())
}

func (s *KeeperTestSuite) TestPoolOrders() {
	type order struct {
		Price sdk.Dec
		Qty   sdk.Int
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
				{utils.ParseDec("4.9950"), sdk.NewInt(25006228)},
				{utils.ParseDec("4.9900"), sdk.NewInt(25043815)},
				{utils.ParseDec("4.9850"), sdk.NewInt(25081497)},
				{utils.ParseDec("4.9800"), sdk.NewInt(25119274)},
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
				{utils.ParseDec("4.9750"), sdk.NewInt(25106679)},
				{utils.ParseDec("4.9700"), sdk.NewInt(25144570)},
				{utils.ParseDec("4.9650"), sdk.NewInt(25182556)},
				{utils.ParseDec("4.9600"), sdk.NewInt(25220637)},
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
				{utils.ParseDec("4.9950"), sdk.NewInt(16662453)},
				{utils.ParseDec("4.9900"), sdk.NewInt(16687499)},
				{utils.ParseDec("4.9850"), sdk.NewInt(66850484)},
				{utils.ParseDec("4.9800"), sdk.NewInt(66951171)},
				{utils.ParseDec("4.9750"), sdk.NewInt(16763015)},
				{utils.ParseDec("4.9700"), sdk.NewInt(16788313)},
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
			_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
			lpAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
			tc.addLiquidity(pool, lpAddr)
			var buyOrders, sellOrders []order
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, pool, true, func(price sdk.Dec, qty sdk.Int) (stop bool) {
				buyOrders = append(buyOrders, order{price, qty})
				return false
			})
			s.App.AMMKeeper.IteratePoolOrders(s.Ctx, pool, false, func(price sdk.Dec, qty sdk.Int) (stop bool) {
				sellOrders = append(sellOrders, order{price, qty})
				return false
			})
			s.Require().EqualValues(tc.buyOrders, buyOrders)
			s.Require().EqualValues(tc.sellOrders, sellOrders)
		})
	}
}
