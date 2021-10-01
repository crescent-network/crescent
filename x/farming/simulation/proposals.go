package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/tendermint/farming/app/params"
	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

/*
[TODO]:
	We need to come up with better ways to simulate public plan proposals.
	Currently, the details are igored and only basic logics are written to simulate.

	These are some of the following considerations that i think need to be discussed and addressed:
	1. Randomize staking coin weights (single or multiple denoms)
	2. Simulate multiple proposals (add new weighted proposal content for multiple plans?)
*/

// Simulation operation weights constants.
const (
	OpWeightSimulateAddPublicPlanProposal    = "op_weight_add_public_plan_proposal"
	OpWeightSimulateUpdatePublicPlanProposal = "op_weight_update_public_plan_proposal"
	OpWeightSimulateDeletePublicPlanProposal = "op_weight_delete_public_plan_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateAddPublicPlanProposal,
			params.DefaultWeightAddPublicPlanProposal,
			SimulateAddPublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateUpdatePublicPlanProposal,
			params.DefaultWeightUpdatePublicPlanProposal,
			SimulateUpdatePublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateDeletePublicPlanProposal,
			params.DefaultWeightDeletePublicPlanProposal,
			SimulateDeletePublicPlanProposal(ak, bk, k),
		),
	}
}

// SimulateAddPublicPlanProposal generates random public plan proposal content
func SimulateAddPublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		poolCoins, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return nil
		}

		// add request proposal
		req := &types.AddRequestProposal{
			Name:               "simulation-test-" + simtypes.RandStringOfLength(r, 5),
			FarmingPoolAddress: simAccount.Address.String(),
			TerminationAddress: simAccount.Address.String(),
			StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1)),
			StartTime:          ctx.BlockTime(),
			EndTime:            ctx.BlockTime().AddDate(0, 1, 0),
			EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000)))),
			EpochRatio:         sdk.ZeroDec(),
		}
		addRequests := []*types.AddRequestProposal{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			addRequests,
			[]*types.UpdateRequestProposal{},
			[]*types.DeleteRequestProposal{},
		)
	}
}

// SimulateUpdatePublicPlanProposal generates random public plan proposal content
func SimulateUpdatePublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		poolCoins, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return nil
		}

		req := &types.UpdateRequestProposal{}

		// TODO: decide which values of fields to randomize
		plans := k.GetPlans(ctx)
		for _, p := range plans {
			if p.GetType() == types.PlanTypePublic {
				startTime := ctx.BlockTime()
				endTime := startTime.AddDate(0, 1, 0)

				switch plan := p.(type) {
				case *types.FixedAmountPlan:
					req.PlanId = plan.GetId()
					req.Name = plan.GetName()
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochAmount = sdk.NewCoins(sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))))
				case *types.RatioPlan:
					req.PlanId = plan.GetId()
					req.Name = plan.GetName()
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochRatio = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 1)
				}
				break
			}
		}

		if req.PlanId == 0 {
			return nil
		}

		updateRequests := []*types.UpdateRequestProposal{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]*types.AddRequestProposal{},
			updateRequests,
			[]*types.DeleteRequestProposal{},
		)
	}
}

// SimulateDeletePublicPlanProposal generates random public plan proposal content
func SimulateDeletePublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil
		}

		req := &types.DeleteRequestProposal{}

		plans := k.GetPlans(ctx)
		for _, p := range plans {
			if p.GetType() == types.PlanTypePublic {
				req.PlanId = p.GetId()
				break
			}
		}

		if req.PlanId == 0 {
			return nil
		}

		deleteRequest := []*types.DeleteRequestProposal{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]*types.AddRequestProposal{},
			[]*types.UpdateRequestProposal{},
			deleteRequest,
		)
	}
}
