package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestFarmingPlan_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(plan *types.FarmingPlan)
		expectedErr string
	}{
		{
			"happy case",
			func(plan *types.FarmingPlan) {},
			"",
		},
		{
			"too long description",
			func(plan *types.FarmingPlan) {
				plan.Description = strings.Repeat("x", types.MaxPlanDescriptionLen+1)
			},
			"too long plan description, maximum 200",
		},
		{
			"invalid farming pool address",
			func(plan *types.FarmingPlan) {
				plan.FarmingPoolAddress = "invalidaddr"
			},
			"invalid farming pool address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid termination address",
			func(plan *types.FarmingPlan) {
				plan.TerminationAddress = "invalidaddr"
			},
			"invalid termination address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"empty reward allocations",
			func(plan *types.FarmingPlan) {
				plan.RewardAllocations = []types.FarmingRewardAllocation{}
			},
			"invalid reward allocations: empty reward allocations",
		},
		{
			"invalid pool id",
			func(plan *types.FarmingPlan) {
				plan.RewardAllocations = []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(0, utils.ParseCoins("100_000000ucre")),
				}
			},
			"invalid reward allocations: pool id must not be 0",
		},
		{
			"invalid rewards per day",
			func(plan *types.FarmingPlan) {
				plan.RewardAllocations = []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, sdk.Coins{utils.ParseCoin("0ucre")}),
				}
			},
			"invalid reward allocations: invalid rewards per day: coin 0ucre amount is not positive",
		},
		{
			"too much rewards per day",
			func(plan *types.FarmingPlan) {
				plan.RewardAllocations = []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(
						1,
						utils.ParseCoins("57896044618658097711785492504343953926634992332820282019728792003956564819967ucre")),
				}
			},
			"invalid reward allocations: too much rewards per day",
		},
		{
			"duplicate pool id",
			func(plan *types.FarmingPlan) {
				plan.RewardAllocations = []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000stake")),
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("200_000000stake")),
				}
			},
			"invalid reward allocations: duplicate pool id: 1",
		},
		{
			"invalid start/end time",
			func(plan *types.FarmingPlan) {
				plan.StartTime = utils.ParseTime("2023-01-01T00:00:00Z")
				plan.EndTime = utils.ParseTime("2023-01-01T00:00:00Z")
			},
			"end time must be after start time",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := types.NewFarmingPlan(
				1, "Farming Plan", utils.TestAddress(0), utils.TestAddress(1),
				[]types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000stake")),
					types.NewFarmingRewardAllocation(2, utils.ParseCoins("200_000000stake")),
				},
				utils.ParseTime("2022-01-01T00:00:00Z"),
				utils.ParseTime("2023-01-01T00:00:00Z"), true)
			tc.malleate(&plan)
			err := plan.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestPlan_IsActiveAt(t *testing.T) {
	plan := types.NewFarmingPlan(
		1, "Farming Plan", utils.TestAddress(0), utils.TestAddress(0),
		[]types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"), false)
	require.False(t, plan.IsActiveAt(utils.ParseTime("2021-12-31T23:59:59Z")))
	require.True(t, plan.IsActiveAt(utils.ParseTime("2022-01-01T00:00:00Z")))
	require.True(t, plan.IsActiveAt(utils.ParseTime("2022-12-31T23:59:59Z")))
	require.False(t, plan.IsActiveAt(utils.ParseTime("2023-01-01T00:00:00Z")))
}
