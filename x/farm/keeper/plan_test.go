package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestCreatePrivatePlan_PastEndTime() {
	s.ctx = s.ctx.WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	creatorAddr := utils.TestAddress(0)
	s.fundAddr(creatorAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, creatorAddr, "Farming Plan",
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(1, utils.ParseCoins("100_00000stake")),
		},
		utils.ParseTime("2020-01-01T00:00:00Z"),
		utils.ParseTime("2021-01-01T00:00:00Z"))
	s.Require().EqualError(err, "end time is past: invalid request")
}

func (s *KeeperTestSuite) TestCreatePrivatePlan_TooManyPrivatePlans() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPairWithLastPrice("denom2", "denom3", sdk.NewDec(1))

	s.keeper.SetMaxNumPrivatePlans(s.ctx, 1)

	s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	creatorAddr := utils.TestAddress(0)
	s.fundAddr(creatorAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, creatorAddr, "Farming Plan",
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(2, utils.ParseCoins("100_000000stake")),
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"))
	s.Require().EqualError(
		err, "maximum number of active private plans reached: invalid request")
}

func (s *KeeperTestSuite) TestCreatePrivatePlan_PairNotFound() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))

	creatorAddr := utils.TestAddress(0)
	s.fundAddr(creatorAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, creatorAddr, "Farming Plan",
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(2, utils.ParseCoins("100_000000stake")),
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"))
	s.Require().EqualError(
		err, "pair 2 not found: not found")
}

func (s *KeeperTestSuite) TestAllocateRewards_NoFarmer() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	plan := s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
		types.NewDenomRewardAllocation("pool1", utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	s.nextBlock()

	s.assertEq(utils.ParseCoins("10000_000000stake"), s.getBalances(plan.GetFarmingPoolAddress()))
	farm, _ := s.keeper.GetFarm(s.ctx, "pool1")
	s.assertEq(sdk.DecCoins{}, farm.CurrentRewards)
	s.assertEq(sdk.DecCoins{}, farm.OutstandingRewards)
}

func (s *KeeperTestSuite) TestAllocateRewards_PairWithNoLastPrice() {
	s.createPair("denom1", "denom2")
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	plan := s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool1"))

	s.nextBlock()

	s.assertEq(utils.ParseCoins("10000_000000stake"), s.getBalances(plan.GetFarmingPoolAddress()))
	farm, _ := s.keeper.GetFarm(s.ctx, "pool1")
	s.assertEq(sdk.DecCoins{}, farm.CurrentRewards)
	s.assertEq(sdk.DecCoins{}, farm.OutstandingRewards)
}

func (s *KeeperTestSuite) TestAllocateRewards() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	// Ratio between two pools' liquidity ~= 1:6.83
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createRangedPool(
		1, utils.ParseCoins("200_000000denom1,200_000000denom2"),
		utils.ParseDec("0.5"), utils.ParseDec("2"), utils.ParseDec("1.0"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool1"))
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool2"))

	s.nextBlock()

	farm, _ := s.keeper.GetFarm(s.ctx, "pool1")
	// Block rewards = 100_000000(stake) * 5(secs) / 86400(secs) ~= 5787(stake)
	// Rewards for pool1 = 5787(stake) * (1 / 7.83) ~= 739(stake)
	s.assertEq(utils.ParseDecCoins("739.228954652576344845stake"), farm.CurrentRewards)

	farm, _ = s.keeper.GetFarm(s.ctx, "pool2")
	// Rewards for pool2 = 5787(stake) * (1 / 7.83) ~= 5047(stake)
	s.assertEq(utils.ParseDecCoins("5047.771045347423649368stake"), farm.CurrentRewards)
}

func (s *KeeperTestSuite) TestAllocateRewards_MultiplePlansToOnePair() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("200_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool1"))

	s.nextBlock()

	s.assertEq(utils.ParseDecCoins("17361stake"), s.rewards(farmerAddr, "pool1"))
}

func (s *KeeperTestSuite) TestAllocateRewards_InsufficientFunds() {
	// Create two pairs and pools.
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPairWithLastPrice("denom2", "denom3", sdk.NewDec(1))
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createPool(2, utils.ParseCoins("100_000000denom2,100_000000denom3"))

	farmingPoolAddr := utils.TestAddress(100)
	// Create two public plans sharing the same farming pool address.
	s.createPublicPlan(farmingPoolAddr, farmingPoolAddr, []types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	})
	s.createPublicPlan(farmingPoolAddr, farmingPoolAddr, []types.RewardAllocation{
		types.NewPairRewardAllocation(2, utils.ParseCoins("100_000000stake")),
	})
	s.fundAddr(farmingPoolAddr, utils.ParseCoins("17361stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1000000pool1"))
	s.farm(farmerAddr, utils.ParseCoin("1000000pool2"))

	s.nextBlock()

	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr, "pool1"))
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr, "pool2"))
	s.assertEq(utils.ParseCoins("11574stake"), s.getBalances(types.RewardsPoolAddress))

	s.nextBlock()

	s.assertEq(utils.ParseCoins("5787stake"), s.getBalances(farmingPoolAddr))
	s.assertEq(utils.ParseCoins("11574stake"), s.getBalances(types.RewardsPoolAddress))

	// Rewards allocation has been skipped.
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr, "pool1"))
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr, "pool2"))
}

func (s *KeeperTestSuite) TestAllocatedRewards_Complicated() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createRangedPool(
		1, utils.ParseCoins("100_000000denom1,100_000000denom2"),
		utils.ParseDec("0.8"), utils.ParseDec("1.25"), sdk.NewDec(1))
	s.createRangedPool(
		1, utils.ParseCoins("100_000000denom1,100_000000denom2"),
		utils.ParseDec("0.4"), utils.ParseDec("2.5"), sdk.NewDec(1))

	s.createPairWithLastPrice("denom2", "denom3", sdk.NewDec(1))
	s.createRangedPool(
		2, utils.ParseCoins("100_000000denom2,100_000000denom3"),
		utils.ParseDec("0.8"), utils.ParseDec("1.25"), sdk.NewDec(1))

	s.createPairWithLastPrice("denom3", "denom4", sdk.NewDec(1))
	s.createRangedPool(
		3, utils.ParseCoins("100_000000denom3,100_000000denom4"),
		utils.ParseDec("0.8"), utils.ParseDec("1.25"), sdk.NewDec(1))
	s.createRangedPool(
		3, utils.ParseCoins("100_000000denom3,100_000000denom4"),
		utils.ParseDec("0.4"), utils.ParseDec("2.5"), sdk.NewDec(1))

	s.createPairWithLastPrice("denom4", "denom5", sdk.NewDec(1))
	s.createRangedPool(
		4, utils.ParseCoins("100_000000denom4,100_000000denom5"),
		utils.ParseDec("0.8"), utils.ParseDec("1.25"), sdk.NewDec(1))

	farmingPoolAddr1 := utils.TestAddress(100)
	s.createPublicPlan(farmingPoolAddr1, farmingPoolAddr1, []types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
		types.NewPairRewardAllocation(2, utils.ParseCoins("100_000000stake")),
	})
	s.createPublicPlan(farmingPoolAddr1, farmingPoolAddr1, []types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
		types.NewPairRewardAllocation(3, utils.ParseCoins("100_000000stake")),
	})
	s.fundAddr(farmingPoolAddr1, utils.ParseCoins("10000_000000stake"))

	farmingPoolAddr2 := utils.TestAddress(101)
	s.createPublicPlan(farmingPoolAddr2, farmingPoolAddr2, []types.RewardAllocation{
		types.NewPairRewardAllocation(2, utils.ParseCoins("100_000000stake")),
		types.NewPairRewardAllocation(4, utils.ParseCoins("100_000000stake")),
	})
	s.fundAddr(farmingPoolAddr2, utils.ParseCoins("10000_000000stake"))

	farmerAddr1 := utils.TestAddress(0)
	s.farm(farmerAddr1, utils.ParseCoin("2000000pool1"))
	s.farm(farmerAddr1, utils.ParseCoin("1000000pool3"))
	s.farm(farmerAddr1, utils.ParseCoin("1000000pool5"))

	farmerAddr2 := utils.TestAddress(1)
	s.farm(farmerAddr2, utils.ParseCoin("1000000pool1"))
	s.farm(farmerAddr2, utils.ParseCoin("1000000pool2"))
	s.farm(farmerAddr2, utils.ParseCoin("1000000pool5"))

	farmerAddr3 := utils.TestAddress(2)
	s.farm(farmerAddr3, utils.ParseCoin("1000000pool3"))
	s.farm(farmerAddr3, utils.ParseCoin("1000000pool4"))
	s.farm(farmerAddr3, utils.ParseCoin("1000000pool5"))
	s.farm(farmerAddr3, utils.ParseCoin("1000000pool6"))

	s.nextBlock()

	// 11,574stake(for pair 1, from two plans)
	// -> 8,991stake(for pool 1, has 77.69% shares)
	// -> 5,994stake(for farmer 1, has 66.66% shares)
	s.assertEq(utils.ParseDecCoins("5994.228604400582stake"), s.rewards(farmerAddr1, "pool1"))
	// 11,574stake(for pair 2, from two plans)
	// -> 11,574stake(for pool 3, has 100% shares)
	// -> 5,787stake(for farmer 1, has 50% shares)
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr1, "pool3"))
	// 5,787stake(for pair 3, from one plan)
	// -> 1,291stake(for pool 5, has 22.31% shares)
	// -> 430stake(for farmer 1, has 33.33% shares)
	s.assertEq(utils.ParseDecCoins("430.442848899854stake"), s.rewards(farmerAddr1, "pool5"))

	// 11,574stake(for pair 1, from two plans)
	// -> 8,991stake(for pool 1, has 77.69% shares)
	// -> 2,997stake(for farmer 2, has 33.33% shares)
	s.assertEq(utils.ParseDecCoins("2997.114302200291stake"), s.rewards(farmerAddr2, "pool1"))
	// 11,574stake(for pair 1, from two plans)
	// -> 2,582stake(for pool 2, has 22.31% shares)
	// -> 2,582stake(for farmer 2, has 100% shares)
	s.assertEq(utils.ParseDecCoins("2582.657093399126stake"), s.rewards(farmerAddr2, "pool2"))
	// 5,787stake(for pair 3, from one plan)
	// -> 1,291stake(for pool 5, has 22.31% shares)
	// -> 430stake(for farmer 2, has 33.33% shares)
	s.assertEq(utils.ParseDecCoins("430.442848899854stake"), s.rewards(farmerAddr2, "pool5"))

	// 11,574stake(for pair 2, from two plans)
	// -> 11,574stake(for pool 3, has 100% shares)
	// -> 5,787stake(for farmer 3, has 50% shares)
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr3, "pool3"))
	// 5,787stake(for pair 3, from one plan)
	// -> 4,495(for pool 4, has 77.69% shares)
	// -> 4,495stake(for farmer 3, has 100% shares)
	s.assertEq(utils.ParseDecCoins("4495.671453300436stake"), s.rewards(farmerAddr3, "pool4"))
	// 5,787stake(for pair 3, from one plan)
	// -> 1,291stake(for pool 5, has 22.31% shares)
	// -> 430stake(for farmer 3, has 33.33% shares)
	s.assertEq(utils.ParseDecCoins("430.442848899854stake"), s.rewards(farmerAddr3, "pool5"))
	// 5,787stake(for pair 4, from one plan)
	// -> 5,787stake(for pool 6, has 100% shares)
	// -> 5,787stake(for farmer 3, has 100% shares)
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr3, "pool6"))
}

func (s *KeeperTestSuite) TestAllocateRewards_ToDenom() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createRangedPool(
		1, utils.ParseCoins("100_000000denom1,100_000000denom2"),
		utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"))

	s.createPrivatePlan([]types.RewardAllocation{
		types.NewDenomRewardAllocation("pool1", utils.ParseCoins("100_000000stake")),
		types.NewDenomRewardAllocation("pool2", utils.ParseCoins("200_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr1 := utils.TestAddress(0)
	s.farm(farmerAddr1, utils.ParseCoin("1000000pool1"))
	s.farm(farmerAddr1, utils.ParseCoin("1000000pool2"))

	farmerAddr2 := utils.TestAddress(1)
	s.farm(farmerAddr2, utils.ParseCoin("1000000pool2"))

	s.nextBlock()

	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr1, "pool1"))
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr1, "pool2"))
	s.assertEq(utils.ParseDecCoins("5787stake"), s.rewards(farmerAddr2, "pool2"))
}

func (s *KeeperTestSuite) TestAllocateRewards_ToPairAndDenom() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createRangedPool(
		1, utils.ParseCoins("100_000000denom1,100_000000denom2"),
		utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"))

	s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewDenomRewardAllocation("pool1", utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1000000pool1"))
	s.farm(farmerAddr, utils.ParseCoin("1000000pool2"))

	s.nextBlock()

	// 1310stake(from plan 1, pool 1 has 22.65% shares)
	// + 5787stake(for pool 1 from plan 2)
	// ~= 7097stake
	s.assertEq(utils.ParseDecCoins("7097.992302078128stake"), s.rewards(farmerAddr, "pool1"))
	// 4476stake(from plan 1, pool 2 has 77.35% shares)
	s.assertEq(utils.ParseDecCoins("4476.007697921871stake"), s.rewards(farmerAddr, "pool2"))
}
