package types_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func TestPlan_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(plan *types.Plan)
		expectedErr string
	}{
		{
			"happy case",
			func(plan *types.Plan) {},
			"",
		},
		{
			"too long description",
			func(plan *types.Plan) {
				plan.Description = strings.Repeat("x", types.MaxPlanDescriptionLen+1)
			},
			"too long plan description, maximum 200",
		},
		{
			"invalid farming pool address",
			func(plan *types.Plan) {
				plan.FarmingPoolAddress = "invalidaddr"
			},
			"invalid farming pool address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid termination address",
			func(plan *types.Plan) {
				plan.TerminationAddress = "invalidaddr"
			},
			"invalid termination address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"same farming pool address and termination address",
			func(plan *types.Plan) {
				plan.FarmingPoolAddress = utils.TestAddress(0).String()
				plan.TerminationAddress = utils.TestAddress(1).String()
				plan.IsPrivate = false
			},
			"farming pool address and termination address of a public plan must be same: cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a != cosmos1qgqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqggwm7m",
		},
		{
			"empty reward allocations",
			func(plan *types.Plan) {
				plan.RewardAllocations = []types.RewardAllocation{}
			},
			"invalid reward allocations: empty reward allocations",
		},
		{
			"zero pair id",
			func(plan *types.Plan) {
				plan.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        0,
						RewardsPerDay: utils.ParseCoins("1_000000stake"),
					},
				}
			},
			"invalid reward allocations: pair id must not be zero",
		},
		{
			"invalid rewards per day",
			func(plan *types.Plan) {
				plan.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("0stake")},
					},
				}
			},
			"invalid reward allocations: invalid rewards per day: coin 0stake amount is not positive",
		},
		{
			"too much rewards per day",
			func(plan *types.Plan) {
				plan.RewardAllocations = []types.RewardAllocation{
					{
						PairId: 1,
						RewardsPerDay: utils.ParseCoins("57896044618658097711785492504343953926634992332820282019728792003956564819967stake"),
					},
				}
			},
			"invalid reward allocations: too much rewards per day",
		},
		{
			"duplicate pair id",
			func(plan *types.Plan) {
				plan.RewardAllocations = []types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("100_000000stake")},
					},
					{
						PairId:        1,
						RewardsPerDay: sdk.Coins{utils.ParseCoin("200_000000stake")},
					},
				}
			},
			"invalid reward allocations: duplicate pair id: 1",
		},
		{
			"invalid start/end time",
			func(plan *types.Plan) {
				plan.StartTime = utils.ParseTime("2023-01-01T00:00:00Z")
				plan.EndTime = utils.ParseTime("2023-01-01T00:00:00Z")
			},
			"end time must be after start time",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := types.NewPlan(
				1, "Farming Plan", utils.TestAddress(0), utils.TestAddress(1),
				[]types.RewardAllocation{
					{
						PairId:        1,
						RewardsPerDay: utils.ParseCoins("100_000000stake"),
					},
					{
						PairId:        2,
						RewardsPerDay: utils.ParseCoins("200_000000stake"),
					},
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
	plan := types.NewPlan(
		1, "Farming Plan", utils.TestAddress(0), utils.TestAddress(0),
		[]types.RewardAllocation{
			{
				PairId:        1,
				RewardsPerDay: utils.ParseCoins("100_000000stake"),
			},
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"), false)
	require.False(t, plan.IsActiveAt(utils.ParseTime("2021-12-31T23:59:59Z")))
	require.True(t, plan.IsActiveAt(utils.ParseTime("2022-01-01T00:00:00Z")))
	require.True(t, plan.IsActiveAt(utils.ParseTime("2022-12-31T23:59:59Z")))
	require.False(t, plan.IsActiveAt(utils.ParseTime("2023-01-01T00:00:00Z")))
}
