package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestPoolParameterChangeProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(p *types.PoolParameterChangeProposal)
		expectedErr string
	}{
		{
			"valid",
			func(p *types.PoolParameterChangeProposal) {},
			"",
		},
		{
			"empty changes",
			func(p *types.PoolParameterChangeProposal) {
				p.Changes = nil
			},
			"changes must not be empty: invalid request",
		},
		{
			"invalid pool id",
			func(p *types.PoolParameterChangeProposal) {
				p.Changes = []types.PoolParameterChange{
					types.NewPoolParameterChange(0, 5),
				}
			},
			"pool id must not be 0: invalid request",
		},
		{
			"not allowed tick spacing",
			func(p *types.PoolParameterChangeProposal) {
				p.Changes = []types.PoolParameterChange{
					types.NewPoolParameterChange(1, 7),
				}
			},
			"tick spacing 7 is not allowed: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewPoolParameterChangeProposal(
				"Title", "Description", []types.PoolParameterChange{
					types.NewPoolParameterChange(1, 5),
					types.NewPoolParameterChange(2, 10),
				})
			tc.malleate(p)
			err := p.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestPublicFarmingPlanProposal_ValidateBasic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(p *types.PublicFarmingPlanProposal)
		expectedErr string
	}{
		{
			"valid",
			func(p *types.PublicFarmingPlanProposal) {},
			"",
		},
		{
			"empty requests",
			func(p *types.PublicFarmingPlanProposal) {
				p.CreateRequests = nil
				p.TerminateRequests = nil
			},
			"requests must not be empty: invalid request",
		},
		{
			"invalid farming plan id",
			func(p *types.PublicFarmingPlanProposal) {
				p.TerminateRequests = []types.TerminateFarmingPlanRequest{
					types.NewTerminateFarmingPlanRequest(0),
				}
			},
			"farming plan id must not be zero: invalid request",
		},
		// the rest of the checks done performed in TestFarmingPlan_Validate
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := types.NewPublicFarmingPlanProposal(
				"Title", "Description", []types.CreatePublicFarmingPlanRequest{
					types.NewCreatePublicFarmingPlanRequest(
						"Farming plan", utils.TestAddress(1), utils.TestAddress(2),
						[]types.FarmingRewardAllocation{
							types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ucre")),
						},
						utils.ParseTime("2023-01-01T00:00:00Z"),
						utils.ParseTime("2023-07-01T00:00:00Z")),
				}, []types.TerminateFarmingPlanRequest{
					types.NewTerminateFarmingPlanRequest(1),
				})
			tc.malleate(p)
			err := p.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func ExamplePoolParameterChangeProposal_String() {
	p := types.NewPoolParameterChangeProposal(
		"Title", "Description", []types.PoolParameterChange{
			types.NewPoolParameterChange(1, 5),
			types.NewPoolParameterChange(2, 10),
		})
	fmt.Println(p.String())

	// Output:
	// Pool Parameter Change Proposal:
	//   Title:       Title
	//   Description: Description
	//   Changes:
	//     Pool Parameter Change:
	//       Pool Id:      1
	//       Tick Spacing: 5
	//     Pool Parameter Change:
	//       Pool Id:      2
	//       Tick Spacing: 10
}

func ExamplePublicFarmingPlanProposal_String() {
	farmingPoolAddr1 := utils.TestAddress(10000)
	farmingPoolAddr2 := utils.TestAddress(20000)
	termAddr1 := utils.TestAddress(30000)
	p := types.NewPublicFarmingPlanProposal(
		"Title", "Description", []types.CreatePublicFarmingPlanRequest{
			types.NewCreatePublicFarmingPlanRequest(
				"First plan", farmingPoolAddr1, farmingPoolAddr1, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ucre,50_000000uatom")),
					types.NewFarmingRewardAllocation(2, utils.ParseCoins("200_000000ucre,50_000000uatom")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
			types.NewCreatePublicFarmingPlanRequest(
				"Second plan", farmingPoolAddr2, termAddr1, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(3, utils.ParseCoins("500_000000ucre")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2023-07-01T00:00:00Z")),
		}, []types.TerminateFarmingPlanRequest{
			types.NewTerminateFarmingPlanRequest(1),
			types.NewTerminateFarmingPlanRequest(2),
		})
	fmt.Println(p.String())

	// Output:
	// Public Farming Plan Proposal:
	//   Title:       Title
	//   Description: Description
	//   Create Requests:
	//     Create Public Farming Plan Request:
	//       Description:          First plan
	//       Farming Pool Address: cosmos15zwqzqqqqqqqqqqqqqqqqqqqqqqqqqqqg6q7lu
	//       Termination Address:  cosmos15zwqzqqqqqqqqqqqqqqqqqqqqqqqqqqqg6q7lu
	//       Start Time:           2023-01-01 00:00:00 +0000 UTC
	//       End Time:             2024-01-01 00:00:00 +0000 UTC
	//       Reward Allocations:
	//         Reward Allocation:
	//           Pool Id:         1
	//           Rewards Per Day: 50000000uatom,100000000ucre
	//         Reward Allocation:
	//           Pool Id:         2
	//           Rewards Per Day: 50000000uatom,200000000ucre
	//     Create Public Farming Plan Request:
	//       Description:          Second plan
	//       Farming Pool Address: cosmos1czuqyqqqqqqqqqqqqqqqqqqqqqqqqqqqqvvkq5
	//       Termination Address:  cosmos1ur2qxqqqqqqqqqqqqqqqqqqqqqqqqqqqa33g34
	//       Start Time:           2023-01-01 00:00:00 +0000 UTC
	//       End Time:             2023-07-01 00:00:00 +0000 UTC
	//       Reward Allocations:
	//         Reward Allocation:
	//           Pool Id:         3
	//           Rewards Per Day: 500000000ucre
	//   Terminate Farming Plan Request:
	//     Terminate Public Farming Plan Request:
	//       Farming Plan Id: 1
	//     Terminate Public Farming Plan Request:
	//       Farming Plan Id: 2
}
