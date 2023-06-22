package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/app/params"
	"github.com/crescent-network/crescent/v5/x/budget/keeper"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateUpdateBudgetPlans = "op_weight_update_budget_plans"
)

// ProposalContents defines the module weighted proposals' contents for mocking param changes, other actions with keeper
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateUpdateBudgetPlans,
			params.DefaultWeightUpdateBudgetPlans,
			SimulateUpdateBudgetPlans(k),
		),
	}
}

// SimulateUpdateBudgetPlans generates random update budget plans param change proposal content.
func SimulateUpdateBudgetPlans(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)

		params.Budgets = GenBudgets(r, ctx, accs)
		params.EpochBlocks = GenEpochBlocks(r)

		// manually set params for simulation
		k.SetParams(ctx, params)

		return nil
	}
}
