package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestAddLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)

	// This is one of the cases which types.ValidatePriceRange fails;
	// lower price is not a valid tick price
	_, _, _, err := s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("0.9999999"), utils.ParseDec("1"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.Require().EqualError(
		err, "invalid lower tick price: 0.999999900000000000: invalid request")

	// Pool not found
	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, 2, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.Require().EqualError(err, "pool not found: not found")

	// lowerTick % tickSpacing != 0
	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("0.9997"), utils.ParseDec("1.1"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.Require().EqualError(
		err, "lower tick -30 must be multiple of tick spacing 50: invalid request")

	// upperTick % tickSpacing != 0
	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("0.9"), utils.ParseDec("1.0003"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.Require().EqualError(
		err, "upper tick 3 must be multiple of tick spacing 50: invalid request")

	// Denom not in the pool
	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1_000000uatom,1_000000uusd"))
	s.Require().EqualError(err, "pool doesn't have denom uatom: invalid request")

	// Added liquidity is zero
	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("0.9"),
		utils.ParseCoins("1_000000ucre"))
	s.Require().EqualError(err, "added liquidity is zero: invalid request")

	// Happy case
	position, liquidity, amt := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.AssertEqual(sdk.NewInt(9472135), liquidity)
	s.AssertEqual(liquidity, position.Liquidity)
	s.AssertEqual(utils.ParseCoins("1_000000ucre,1_000000uusd"), amt)
	s.AssertEqual(amt, s.GetAllBalances(pool.MustGetReserveAddress()))
}

func (s *KeeperTestSuite) TestRemoveLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1_000000ucre,1_000000uusd"))
	s.AssertEqual(sdk.NewInt(9472135), position.Liquidity)

	// Position not found
	_, _, err := s.keeper.RemoveLiquidity(s.Ctx, lpAddr, lpAddr, 2, position.Liquidity)
	s.Require().EqualError(err, "position not found: not found")

	// Position is not owed by the account
	anotherAddr := s.FundedAccount(2, enoughCoins)
	_, _, err = s.keeper.RemoveLiquidity(
		s.Ctx, anotherAddr, anotherAddr, position.Id, position.Liquidity)
	s.Require().EqualError(err, "position is not owned by the address: unauthorized")

	// Removing liquidity > position liquidity
	_, _, err = s.keeper.RemoveLiquidity(
		s.Ctx, lpAddr, lpAddr, position.Id, position.Liquidity.Add(sdk.NewInt(10000)))
	s.Require().EqualError(
		err, "liquidity in position is smaller than liquidity specified: 9472135 < 9482135: invalid request")

	// Happy case
	lpBalancesBefore := s.GetAllBalances(lpAddr)
	position, amt := s.RemoveLiquidity(lpAddr, position.Id, sdk.NewInt(1000000))
	s.AssertEqual(sdk.NewInt(8472135), position.Liquidity)
	s.AssertEqual(utils.ParseCoins("105572ucre,105572uusd"), amt)
	lpBalancesAfter := s.GetAllBalances(lpAddr)
	s.AssertEqual(amt, lpBalancesAfter.Sub(lpBalancesBefore))

	position, amt = s.RemoveLiquidity(lpAddr, position.Id, position.Liquidity)
	s.AssertEqual(sdk.ZeroInt(), position.Liquidity)
	s.AssertEqual(utils.ParseCoins("894428ucre,894428uusd"), amt)

	s.AssertEqual(sdk.Coins{}, s.GetAllBalances(pool.MustGetReserveAddress()))
}

func (s *KeeperTestSuite) TestCollect() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("100_000000ucre,100_000000uusd"))
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("0.9"), utils.ParseDec("1.1"),
		utils.ParseCoins("100_000000ucre,100_000000uusd"))
	// Position 2 is narrower, so position 2's liquidity > position 1's liquidity.
	s.AssertEqual(sdk.NewInt(947213595), position1.Liquidity)
	s.AssertEqual(sdk.NewInt(1948683298), position2.Liquidity)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("1.01"), sdk.NewInt(5_000000), 0)

	// Position 2's fee > position 1's fee, and there's no farming rewards.
	fee, farmingRewards := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("2453ucre,5348uusd"), fee)
	s.AssertEqual(utils.ParseCoins(""), farmingRewards)
	fee, farmingRewards = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("5046ucre,11003uusd"), fee)
	s.AssertEqual(utils.ParseCoins(""), farmingRewards)

	// Let's start farming.
	farmingPoolAddr := s.FundedAccount(10000, enoughCoins)
	s.CreatePublicFarmingPlan(
		"Farming plan", farmingPoolAddr, farmingPoolAddr,
		[]types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre,50_000000uatom")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.NextBlock()
	s.NextBlock()

	// Position 2's farming rewards > position 1's farming rewards,
	// and fees are not touched.
	fee, farmingRewards = s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("2453ucre,5348uusd"), fee)
	s.AssertEqual(utils.ParseCoins("1892uatom,3785ucre"), farmingRewards)
	fee, farmingRewards = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("5046ucre,11003uusd"), fee)
	s.AssertEqual(utils.ParseCoins("3893uatom,7788ucre"), farmingRewards)

	// And here are some failing cases...

	// Position not found
	err := s.keeper.Collect(s.Ctx, lpAddr1, lpAddr1, 3, utils.ParseCoins("1000ucre"))
	s.Require().EqualError(err, "position not found: not found")

	// Position not owned
	err = s.keeper.Collect(s.Ctx, lpAddr1, lpAddr1, position2.Id, utils.ParseCoins("1000ucre"))
	s.Require().EqualError(err, "position is not owned by the address: unauthorized")

	// amt > collectible
	// Use cached context here, since RemoveLiquidity is called inside and
	// alters the state.
	cacheCtx, _ := s.Ctx.CacheContext()
	err = s.keeper.Collect(cacheCtx, lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("20000ucre"))
	s.Require().EqualError(
		err, "cannot collect 20000ucre from collectible coins 1892uatom,6238ucre,5348uusd: invalid request")

	// This fails of course.
	cacheCtx, _ = s.Ctx.CacheContext()
	err = s.keeper.Collect(cacheCtx, lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("1000stake"))
	s.Require().EqualError(
		err, "cannot collect 1000stake from collectible coins 1892uatom,6238ucre,5348uusd: invalid request")

	// Collect collects rewards from fee first, and then from faring rewards.
	lp1BalancesBefore := s.GetAllBalances(lpAddr1)
	rewardsPoolAddr := pool.MustGetRewardsPoolAddress()
	rewardsPoolBalancesBefore := s.GetAllBalances(rewardsPoolAddr)
	farmingRewardsPoolBalancesBefore := s.GetAllBalances(types.FarmingRewardsPoolAddress)

	s.Collect(lpAddr1, position1.Id, utils.ParseCoins("2000ucre,1000uatom"))
	fee, farmingRewards = s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("453ucre,5348uusd"), fee)
	s.AssertEqual(utils.ParseCoins("892uatom,3785ucre"), farmingRewards)

	lp1BalancesAfter := s.GetAllBalances(lpAddr1)
	rewardsPoolBalancesAfter := s.GetAllBalances(rewardsPoolAddr)
	farmingRewardsPoolBalancesAfter := s.GetAllBalances(types.FarmingRewardsPoolAddress)
	rewardsPoolBalancesDiff, _ := rewardsPoolBalancesAfter.SafeSub(rewardsPoolBalancesBefore)
	farmingRewardsPoolBalancesDiff, _ := farmingRewardsPoolBalancesAfter.
		SafeSub(farmingRewardsPoolBalancesBefore)

	s.AssertEqual(utils.ParseCoins("2000ucre,1000uatom"), lp1BalancesAfter.Sub(lp1BalancesBefore))
	s.AssertEqual(
		sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-2000)}},
		rewardsPoolBalancesDiff)
	s.AssertEqual(
		sdk.Coins{sdk.Coin{Denom: "uatom", Amount: sdk.NewInt(-1000)}},
		farmingRewardsPoolBalancesDiff)

	// Collect more
	s.Collect(lpAddr1, position1.Id, utils.ParseCoins("4000ucre,5000uusd"))
	fee, farmingRewards = s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("348uusd"), fee)
	s.AssertEqual(utils.ParseCoins("892uatom,238ucre"), farmingRewards)
}

func (s *KeeperTestSuite) TestReinitializePosition() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	ownerAddr := s.FundedAccount(1, enoughCoins)
	lowerPrice, upperPrice := utils.ParseDec("4.5"), utils.ParseDec("5.5")
	desiredAmt := utils.ParseCoins("100_000000ucre,500_000000uusd")
	position, liquidity, _ := s.AddLiquidity(
		ownerAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("6"), sdk.NewInt(1000000), 0)
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(1000000))

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
	s.PlaceLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("6"), sdk.NewInt(10_000000), 0)
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(10_000000))

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

func (s *KeeperTestSuite) TestRewardsPool() {
	market1, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	market2, pool2 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseCoins("100_000000ucre,500_000000uusd"))
	s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("9"), utils.ParseDec("12"), utils.ParseCoins("100_000000uatom,1000_000000uusd"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market1.Id, ordererAddr, true, utils.ParseDec("6"), sdk.NewInt(1_000000), 0)
	s.PlaceLimitOrder(market2.Id, ordererAddr, false, utils.ParseDec("9"), sdk.NewInt(1_000000), 0)

	s.AssertEqual(utils.ParseCoins("1498ucre,2619uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("17660uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))

	fee, _ := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("1497ucre,2618uusd"), fee)
	s.Collect(lpAddr, position1.Id, utils.ParseCoins("1497ucre,2618uusd"))
	s.AssertEqual(utils.ParseCoins("1ucre,1uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("17660uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))
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

func (s *KeeperTestSuite) TestPositionAssets_ZeroLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)

	position, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	coin0, coin1, err := s.keeper.PositionAssets(s.Ctx, position.Id)
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseCoin("82529840ucre"), coin0)
	s.AssertEqual(utils.ParseCoin("499999999uusd"), coin1)

	// Remove all liquidity from the position.
	s.RemoveLiquidity(lpAddr1, position.Id, position.Liquidity)

	coin0, coin1, err = s.keeper.PositionAssets(s.Ctx, position.Id)
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseCoin("0ucre"), coin0)
	s.AssertEqual(utils.ParseCoin("0uusd"), coin1)
}

func (s *KeeperTestSuite) TestRemoveSmallLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)

	position, _, amt := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("10000ucre,50000uusd"))
	s.AssertEqual(sdk.NewInt(211803), position.Liquidity)
	s.AssertEqual(utils.ParseCoins("8253ucre,50000uusd"), amt)

	// This will prevent removing the last liquidity from position1 to withdraw
	// all remaining reserves.
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("3"), utils.ParseDec("7"),
		utils.ParseCoins("10000ucre,50000uusd"))

	// Removing very small amount of liquidity may cause withdrawing no assets
	// at all.
	position, amt = s.RemoveLiquidity(lpAddr1, position.Id, sdk.NewInt(1))
	s.AssertEqual(sdk.NewInt(211802), position.Liquidity)
	s.AssertEqual(sdk.Coins{}, amt)

	// Thus, removing all liquidity by removing small amount many times
	// may cause a loss in assets.
	for {
		position, amt = s.RemoveLiquidity(
			lpAddr1, position.Id, utils.MinInt(sdk.NewInt(4), position.Liquidity))
		s.AssertEqual(sdk.Coins{}, amt)
		if position.Liquidity.IsZero() {
			break
		}
	}
}

func (s *KeeperTestSuite) TestAddLiquidity_MinMaxPrice() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, _, _, err := s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, sdk.NewDecWithPrec(1, sdk.Precision), utils.ParseDec("10"),
		utils.ParseCoins("1000_000000ucre,5000_000000uusd"))
	s.Require().EqualError(
		err, "lower price must not be lower than the minimum: "+
			"0.000000000000000001 < 0.000000000000010000: invalid request")

	_, _, _, err = s.keeper.AddLiquidity(
		s.Ctx, lpAddr, lpAddr, pool.Id, utils.ParseDec("4"), sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 40)),
		utils.ParseCoins("1000_000000ucre,5000_000000uusd"))
	s.Require().EqualError(
		err, "upper price must not be higher than the maximum: "+
			"10000000000000000000000000000000000000000.000000000000000000 > "+
			"1000000000000000000000000.000000000000000000: invalid request")
}
