package keeper_test

import (
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestCreatePrivatePlan_PastEndTime() {
	s.ctx = s.ctx.WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, utils.TestAddress(0), "Farming Plan",
		[]types.RewardAllocation{
			{
				PairId:        1,
				RewardsPerDay: utils.ParseCoins("100_000000stake"),
			},
		},
		utils.ParseTime("2020-01-01T00:00:00Z"),
		utils.ParseTime("2021-01-01T00:00:00Z"))
	s.Require().EqualError(err, "end time is past: invalid request")
}

func (s *KeeperTestSuite) TestCreatePrivatePlan_TooManyPrivatePlans() {
	s.createPair("denom1", "denom2")
	s.createPair("denom2", "denom3")

	s.keeper.SetMaxNumPrivatePlans(s.ctx, 1)

	s.createPrivatePlan([]types.RewardAllocation{
		{
			PairId:        1,
			RewardsPerDay: utils.ParseCoins("100_000000stake"),
		},
	})

	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, utils.TestAddress(0), "Farming Plan",
		[]types.RewardAllocation{
			{
				PairId:        2,
				RewardsPerDay: utils.ParseCoins("100_000000stake"),
			},
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"))
	s.Require().EqualError(
		err, "maximum number of active private plans reached: 1: invalid request")
}

func (s *KeeperTestSuite) TestCreatePrivatePlan_PairNotFound() {
	s.createPair("denom1", "denom2")

	_, err := s.keeper.CreatePrivatePlan(
		s.ctx, utils.TestAddress(0), "Farming Plan",
		[]types.RewardAllocation{
			{
				PairId:        2,
				RewardsPerDay: utils.ParseCoins("100_000000stake"),
			},
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"))
	s.Require().EqualError(
		err, "pair 2 not found: not found")
}
