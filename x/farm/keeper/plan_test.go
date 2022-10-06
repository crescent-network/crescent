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
			types.NewRewardAllocation(1, utils.ParseCoins("100_00000stake")),
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
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	creatorAddr := utils.TestAddress(0)
	s.fundAddr(creatorAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, creatorAddr, "Farming Plan",
		[]types.RewardAllocation{
			{
				PairId:        2,
				RewardsPerDay: utils.ParseCoins("100_000000stake"),
			},
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
			types.NewRewardAllocation(2, utils.ParseCoins("100_000000stake")),
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"))
	s.Require().EqualError(
		err, "pair 2 not found: not found")
}

func (s *KeeperTestSuite) TestAllocateRewards_NoFarmer() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	plan := s.createPrivatePlan([]types.RewardAllocation{
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
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
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
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
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
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
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewRewardAllocation(1, utils.ParseCoins("200_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool1"))

	s.nextBlock()

	s.assertEq(utils.ParseDecCoins("17361stake"), s.rewards(farmerAddr, "pool1"))
}
