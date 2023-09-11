package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

func TestBlockPublicPlanPairRewardAllocation(t *testing.T) {
	proposal := types.NewFarmingPlanProposal(
		"Title", "Description", []types.CreatePlanRequest{
			types.NewCreatePlanRequest(
				"Farming plan", utils.TestAddress(1), []types.RewardAllocation{
					types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
		}, nil)
	err := proposal.ValidateBasic()
	require.EqualError(t, err, "pair reward allocation for 1 is disabled: invalid request")
}
