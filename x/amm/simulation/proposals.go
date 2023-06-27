package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

const (
	OpWeightSubmitPoolParameterChangeProposal = "op_weight_submit_pool_parameter_change_proposal"
	OpWeightSubmitPublicFarmingPlanProposal   = "op_weight_submit_public_farming_plan_proposal"

	DefaultWeightPoolParameterChangeProposal = 5
	DefaultWeightPublicFarmingPlanProposal   = 5
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(bk types.BankKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitPoolParameterChangeProposal,
			DefaultWeightPoolParameterChangeProposal,
			SimulatePoolParameterChangeProposal(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSubmitPublicFarmingPlanProposal,
			DefaultWeightPublicFarmingPlanProposal,
			SimulatePublicFarmingPlanProposal(bk, k),
		),
	}
}

func SimulatePoolParameterChangeProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		var changes []types.PoolParameterChange
		k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
			if r.Float64() <= 0.5 {
				allowedTickSpacings := utils.Filter(types.AllowedTickSpacings, func(tickSpacing uint32) bool {
					return tickSpacing < pool.TickSpacing && pool.TickSpacing%tickSpacing == 0
				})
				if len(allowedTickSpacings) > 0 {
					changes = append(changes,
						types.NewPoolParameterChange(
							pool.Id, allowedTickSpacings[r.Intn(len(allowedTickSpacings))]))
				}
			}
			return false
		})
		if len(changes) == 0 {
			return nil
		}
		return types.NewPoolParameterChangeProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			changes,
		)
	}
}

func SimulatePublicFarmingPlanProposal(bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		acc, _ := simtypes.RandomAcc(r, accs)
		var pools []types.Pool
		k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
			pools = append(pools, pool)
			return false
		})
		numCreateReqs := 1 + r.Intn(3)
		var createReqs []types.CreatePublicFarmingPlanRequest
		for i := 0; i < numCreateReqs; i++ {
			startTime := ctx.BlockTime().AddDate(0, 0, r.Intn(5))
			endTime := startTime.AddDate(0, 0, 1+r.Intn(10))
			var rewardAllocs []types.FarmingRewardAllocation
			for _, pool := range utils.Filter(pools, func(types.Pool) bool {
				return r.Float64() <= 0.5
			}) {
				balances := sdk.NewDecCoinsFromCoins(bk.SpendableCoins(ctx, acc.Address)...)
				rewardsPerDay, _ := balances.QuoDec(sdk.NewDec(1000)).TruncateDecimal()
				rewardsPerDay = simtypes.RandSubsetCoins(r, rewardsPerDay)
				if rewardsPerDay.IsAllPositive() {
					rewardAllocs = append(rewardAllocs, types.NewFarmingRewardAllocation(
						pool.Id, rewardsPerDay))
				}
			}
			if len(rewardAllocs) > 0 {
				createReqs = append(createReqs,
					types.NewCreatePublicFarmingPlanRequest(
						"Farming plan", acc.Address, acc.Address,
						rewardAllocs, startTime, endTime))
			}
		}
		var termReqs []types.TerminateFarmingPlanRequest
		k.IterateAllFarmingPlans(ctx, func(plan types.FarmingPlan) (stop bool) {
			if r.Float64() <= 0.2 {
				termReqs = append(termReqs, types.NewTerminateFarmingPlanRequest(plan.Id))
			}
			return false
		})
		if len(createReqs) == 0 && len(termReqs) == 0 {
			return nil
		}
		return types.NewPublicFarmingPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			createReqs,
			termReqs,
		)
	}
}
