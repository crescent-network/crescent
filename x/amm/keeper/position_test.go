package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestAddLiquidity() {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, sdk.NewDec(1))

	senderAddr := s.FundedAccount(1, enoughCoins)
	position, liquidity, amt := s.AddLiquidity(
		senderAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1000000ucre,1000000uusd"))
	fmt.Println(position, liquidity, amt)

	_, amt = s.RemoveLiquidity(senderAddr, position.Id, sdk.NewInt(9472135))
	fmt.Println(amt)
}

func (s *KeeperTestSuite) TestReinitializePosition() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	ownerAddr := s.FundedAccount(1, enoughCoins)
	lowerPrice, upperPrice := utils.ParseDec("4.5"), utils.ParseDec("5.5")
	desiredAmt := utils.ParseCoins("100_000000ucre,500_000000uusd")
	position, liquidity, _ := s.AddLiquidity(
		ownerAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewDec(1000000))
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewDec(1000000))

	s.RemoveLiquidity(ownerAddr, position.Id, liquidity)
	position, _ = s.keeper.GetPosition(s.Ctx, position.Id)
	fmt.Println(position.Liquidity)
	s.AddLiquidity(
		ownerAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)
}

func (s *KeeperTestSuite) TestRemoveAllAndCollect() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	// Accrue fees.
	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewDec(10_000000))
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewDec(10_000000))

	s.RemoveLiquidity(lpAddr, position.Id, position.Liquidity)

	fee, farmingRewards, err := s.keeper.CollectibleCoins(s.Ctx, position.Id)
	s.Require().NoError(err)
	s.Collect(lpAddr, position.Id, fee.Add(farmingRewards...))
}

func (s *KeeperTestSuite) TestNegativeFarmingRewardsGrowthInside() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1.1366"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("1.1"), utils.ParseDec("1.2"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	creatorAddr := s.FundedAccount(2, enoughCoins)
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.NextBlock()
	s.NextBlock()
	_, farmingRewards := s.CollectibleCoins(1)
	s.Require().Equal("11573ucre", farmingRewards.String())
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.9"), utils.ParseDec("1.1"),
		utils.ParseCoins("1000_000000uusd"))
	_, farmingRewards = s.CollectibleCoins(1)
	s.Require().Equal("11573ucre", farmingRewards.String())
	_, farmingRewards = s.CollectibleCoins(2)
	s.Require().Equal("", farmingRewards.String())
}

func (s *KeeperTestSuite) TestNarrowPosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("20"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("19.999"), utils.ParseDec("20.001"),
		utils.ParseCoins("500ucre,10000uusd"))
	s.AssertEqual(sdk.NewInt(89441601), liquidity)

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("19.998"), utils.ParseDec("20.001"),
		utils.ParseCoins("250ucre,10000uusd"))
	s.AssertEqual(sdk.NewInt(44720241), liquidity)

	s.FundAccount(lpAddr, utils.ParseCoins("1_000_000_000_000000000000000000ufoo,1_000_000_000_000000000000000000ubar"))
	_, pool2 := s.CreateMarketAndPool("ufoo", "ubar", utils.ParseDec("1"))

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("0.99999"), utils.ParseDec("1.0001"),
		utils.ParseCoins("1_000_000_000_000000000000000000ufoo,1_000_000_000_000000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("20001499987500662572187476524684"), liquidity)
}

func (s *KeeperTestSuite) TestExtremePrice() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1000000000"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("900000000"), utils.ParseDec("1100000000"),
		utils.ParseCoins("46ucre,50000000000uusd"))
	s.AssertEqual(sdk.NewInt(30811388), liquidity)

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("899900000"), utils.ParseDec("1100000000"),
		utils.ParseCoins("46ucre,50000000000uusd"))
	s.AssertEqual(sdk.NewInt(30779775), liquidity)

	_, pool2 := s.CreateMarketAndPool("uusd", "ucre", utils.ParseDec("0.00000001"))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("0.000000009"), utils.ParseDec("0.000000011"),
		utils.ParseCoins("90686675071631uusd,1000000ucre"))
	s.AssertEqual(sdk.NewInt(194868329792), liquidity)

	s.FundAccount(
		lpAddr, utils.ParseCoins("1000000000000000000000000000000000000000000ufoo,1000000000000000000000000000000000000000000ubar"))
	_, pool3 := s.CreateMarketAndPool("ufoo", "ubar", exchangetypes.PriceAtTick(types.MaxTick-1))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool3.Id, exchangetypes.PriceAtTick(types.MaxTick-2), exchangetypes.PriceAtTick(types.MaxTick),
		utils.ParseCoins("10ufoo,100000000000000000000000000000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("199998499993749943749335928"), liquidity)

	_, pool4 := s.CreateMarketAndPool("ubar", "ufoo", exchangetypes.PriceAtTick(types.MinTick+1))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool4.Id, exchangetypes.PriceAtTick(types.MinTick), exchangetypes.PriceAtTick(types.MinTick+2),
		utils.ParseCoins("10ufoo,1000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("2000050001250"), liquidity)
}

func (s *KeeperTestSuite) TestSmallPosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("0.5"), utils.ParseDec("1.5"),
		utils.ParseCoins("7ucre,10uusd"))
	s.AssertEqual(sdk.NewInt(34), liquidity)
	// TODO: check pool orders

	_, pool2 := s.CreateMarketAndPool("uusd", "ucre", utils.ParseDec("1"))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("0.99999"), utils.ParseDec("1.0001"),
		utils.ParseCoins("100uusd,10ucre"))
	s.AssertEqual(sdk.NewInt(1999994), liquidity)
	// TODO: check pool orders

	s.FundAccount(lpAddr, utils.ParseCoins("1ufoo,1ubar"))
	_, pool3 := s.CreateMarketAndPool("ufoo", "ubar", utils.ParseDec("1000"))
	_, _, _, err := s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool3.Id, utils.ParseDec("900"), utils.ParseDec("1100"),
		utils.ParseCoins("1ufoo,1ubar"))
	s.Require().EqualError(err, "minted liquidity is zero: insufficient funds")
}

func (s *KeeperTestSuite) TestBigPosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.99999"), utils.ParseDec("1.0001"),
		utils.ParseCoins("9999225064335ucre,1000000000000uusd"))
	s.AssertEqual(sdk.NewInt(199999499998770009), liquidity)
}

func (s *KeeperTestSuite) TestSingleSidePosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("10"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("10"), utils.ParseDec("20"),
		utils.ParseCoins("1000000000ucre"))
	s.AssertEqual(sdk.NewInt(10796691275), liquidity)

	_, pool2 := s.CreateMarketAndPool("uusd", "ucre", utils.ParseDec("20"))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("10"), utils.ParseDec("20"),
		utils.ParseCoins("1000000000ucre"))
	s.AssertEqual(sdk.NewInt(763441361), liquidity)
}

func (s *KeeperTestSuite) TestSamePositions() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("2"), utils.ParseDec("7"),
		utils.ParseCoins("84259598ucre,1000000000uusd"))
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("2"), utils.ParseDec("7"),
		utils.ParseCoins("84259598ucre,1000000000uusd"))
	lpAddr3 := s.FundedAccount(3, enoughCoins)
	position3, _, _ := s.AddLiquidity(
		lpAddr3, pool.Id, utils.ParseDec("2"), utils.ParseDec("7"),
		utils.ParseCoins("84259598ucre,1000000000uusd"))
	s.AssertEqual(sdk.NewInt(1216760513), position1.Liquidity)
	s.AssertEqual(sdk.NewInt(1216760513), position2.Liquidity)
	s.AssertEqual(sdk.NewInt(1216760513), position3.Liquidity)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewDec(100000000))

	fee, _ := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("353ucre,235566uusd"), fee)
	fee, _ = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("353ucre,235566uusd"), fee)
	fee, _ = s.CollectibleCoins(position3.Id)
	s.AssertEqual(utils.ParseCoins("353ucre,235566uusd"), fee)

	_, amt := s.RemoveLiquidity(lpAddr1, position1.Id, position1.Liquidity)
	s.AssertEqual(utils.ParseCoins("117592576ucre,842955162uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr2, position2.Id, position2.Liquidity)
	s.AssertEqual(utils.ParseCoins("117592576ucre,842955162uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr3, position3.Id, position3.Liquidity)
	s.AssertEqual(utils.ParseCoins("117592580ucre,842955163uusd"), amt)
}

func (s *KeeperTestSuite) TestSamePositions2() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("500"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("400"), utils.ParseDec("625"),
		utils.ParseCoins("2000000000ucre,1000000000000uusd"))
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("400"), utils.ParseDec("625"),
		utils.ParseCoins("200000000ucre,100000000000uusd"))
	s.AssertEqual(sdk.NewInt(423606797749), position1.Liquidity)
	s.AssertEqual(sdk.NewInt(42360679774), position2.Liquidity)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewDec(200000000))

	fee, _ := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("272726ucre,46039478uusd"), fee)
	fee, _ = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("27272ucre,4603947uusd"), fee)

	// Use cache context to discard the operation results.
	cacheCtx, _ := s.Ctx.CacheContext()
	_, amt, err := s.keeper.RemoveLiquidity(cacheCtx, lpAddr1, lpAddr1, position1.Id, position1.Liquidity)
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseCoins("1818181818ucre,1091790048475uusd"), amt)
	_, amt, err = s.keeper.RemoveLiquidity(cacheCtx, lpAddr2, lpAddr2, position2.Id, position2.Liquidity)
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseCoins("181818182ucre,109179004846uusd"), amt)

	_, amt = s.RemoveLiquidity(lpAddr2, position2.Id, position2.Liquidity)
	s.AssertEqual(utils.ParseCoins("181818181ucre,109179004845uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr1, position1.Id, position1.Liquidity)
	s.AssertEqual(utils.ParseCoins("1818181819ucre,1091790048476uusd"), amt)
}

func (s *KeeperTestSuite) TestSymmetricPositions() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("100"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	_, liquidity, amt := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("50"), utils.ParseDec("200"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	s.AssertEqual(sdk.NewInt(34142135623), liquidity)
	s.AssertEqual(utils.ParseCoins("1000000000ucre,99999999998uusd"), amt)

	lpAddr2 := s.FundedAccount(2, enoughCoins)
	_, liquidity, amt = s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("20"), utils.ParseDec("500"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	s.AssertEqual(sdk.NewInt(18090169943), liquidity)
	s.AssertEqual(utils.ParseCoins("1000000000ucre,99999999996uusd"), amt)
}

func (s *KeeperTestSuite) TestIndependentRemoveLiquidity() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("100"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("50"), utils.ParseDec("200"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("50"), utils.ParseDec("200"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	lpAddr3 := s.FundedAccount(3, enoughCoins)
	position3, _, _ := s.AddLiquidity(
		lpAddr3, pool.Id, utils.ParseDec("20"), utils.ParseDec("500"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	lpAddr4 := s.FundedAccount(4, enoughCoins)
	position4, _, _ := s.AddLiquidity(
		lpAddr4, pool.Id, utils.ParseDec("20"), utils.ParseDec("500"),
		utils.ParseCoins("1000000000ucre,100000000000uusd"))
	s.AssertEqual(sdk.NewInt(34142135623), position1.Liquidity)
	s.AssertEqual(sdk.NewInt(34142135623), position2.Liquidity)
	s.AssertEqual(sdk.NewInt(18090169943), position3.Liquidity)
	s.AssertEqual(sdk.NewInt(18090169943), position4.Liquidity)

	_, amt := s.RemoveLiquidity(lpAddr3, position3.Id, position3.Liquidity)
	s.AssertEqual(utils.ParseCoins("999999999ucre,99999999995uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr1, position1.Id, position1.Liquidity)
	s.AssertEqual(utils.ParseCoins("999999999ucre,99999999997uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr2, position2.Id, position2.Liquidity)
	s.AssertEqual(utils.ParseCoins("999999999ucre,99999999997uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr4, position4.Id, position4.Liquidity)
	s.AssertEqual(utils.ParseCoins("1000000003ucre,99999999999uusd"), amt)
}

func (s *KeeperTestSuite) TestInitialPoolPriceDifference() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr1, false, utils.ParseDec("10"), sdk.NewDec(10000), time.Hour)
	_, buyOrder, _ := s.PlaceLimitOrder(market.Id, ordererAddr2, true, utils.ParseDec("10"), sdk.NewDec(1000000), time.Hour)

	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.AssertEqual(utils.ParseDec("10"), *marketState.LastPrice)

	buyOrder = s.App.ExchangeKeeper.MustGetOrder(s.Ctx, buyOrder.Id)
	s.AssertEqual(sdk.NewDec(990000), buyOrder.OpenQuantity)

	pool := s.CreatePool(market.Id, utils.ParseDec("1"))
	lpAddr := s.FundedAccount(3, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.1"), utils.ParseDec("10"),
		utils.ParseCoins("1000ucre,1000uusd"))
	s.AssertEqual(sdk.NewInt(1462), position.Liquidity)

	s.NextBlock() // Run matching at batch

	marketState = s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	s.AssertEqual(utils.ParseDec("10"), *marketState.LastPrice)
	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	s.AssertEqual(utils.ParseDec("10"), poolState.CurrentPrice)

	_, amt := s.RemoveLiquidity(lpAddr, position.Id, position.Liquidity)
	s.AssertEqual(utils.ParseCoins("4489uusd"), amt)
}

func (s *KeeperTestSuite) TestInitialPoolPriceDifference2() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	_, sellOrder, _ := s.PlaceLimitOrder(market.Id, ordererAddr1, false, utils.ParseDec("1"), sdk.NewDec(1000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, true, utils.ParseDec("1"), sdk.NewDec(10000), time.Hour)

	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.AssertEqual(utils.ParseDec("1"), *marketState.LastPrice)

	sellOrder = s.App.ExchangeKeeper.MustGetOrder(s.Ctx, sellOrder.Id)
	s.AssertEqual(sdk.NewDec(990000), sellOrder.OpenQuantity)

	pool := s.CreatePool(market.Id, utils.ParseDec("10"))
	lpAddr := s.FundedAccount(3, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("1"), utils.ParseDec("100"),
		utils.ParseCoins("100ucre,1000uusd"))
	s.AssertEqual(sdk.NewInt(462), position.Liquidity)

	s.NextBlock() // Run matching at batch

	marketState = s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	s.AssertEqual(utils.ParseDec("1"), *marketState.LastPrice)
	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	s.AssertEqual(utils.ParseDec("1"), poolState.CurrentPrice)

	_, amt := s.RemoveLiquidity(lpAddr, position.Id, position.Liquidity)
	s.AssertEqual(utils.ParseCoins("417ucre,681uusd"), amt) // All remaining reserve
}

func (s *KeeperTestSuite) TestRewardsPool() {
	market1, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	market2, pool2 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseCoins("100_000000ucre,500_000000uusd"))
	s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("9"), utils.ParseDec("12"), utils.ParseCoins("100_000000uatom,1000_000000uusd"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market1.Id, ordererAddr, true, sdk.NewDec(1_000000))
	s.PlaceMarketOrder(market2.Id, ordererAddr, false, sdk.NewDec(1_000000))

	s.AssertEqual(utils.ParseCoins("1499ucre,2620uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("268uatom,14982uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))

	fee, _ := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("1498ucre,2619uusd"), fee)
	s.Collect(lpAddr, position1.Id, utils.ParseCoins("1497ucre,2618uusd"))
	s.AssertEqual(utils.ParseCoins("2ucre,2uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("268uatom,14982uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))
}

func (s *KeeperTestSuite) TestLastRemoveLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	lpAddr3 := s.FundedAccount(3, enoughCoins)

	// Three position parameters are same.
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position3, _, _ := s.AddLiquidity(
		lpAddr3, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	_, amt := s.RemoveLiquidity(lpAddr1, position1.Id, position1.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529840ucre,499999999uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr2, position2.Id, position2.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529840ucre,499999999uusd"), amt)
	// The last liquidity remover takes all remaining reserve balances in the pool.
	_, amt = s.RemoveLiquidity(lpAddr3, position3.Id, position3.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529843ucre,500000002uusd"), amt)

	// No balances left in the pool.
	s.AssertEqual(sdk.Coins{}, s.GetAllBalances(pool.MustGetReserveAddress()))
}
