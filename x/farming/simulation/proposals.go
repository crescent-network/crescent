package simulation

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	liquiditytypes "github.com/cosmosquad-labs/squad/x/liquidity/types"

	"github.com/cosmosquad-labs/squad/app/params"
	"github.com/cosmosquad-labs/squad/x/farming/keeper"
	"github.com/cosmosquad-labs/squad/x/farming/types"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateAddPublicPlanProposal    = "op_weight_add_public_plan_proposal"
	OpWeightSimulateUpdatePublicPlanProposal = "op_weight_update_public_plan_proposal"
	OpWeightSimulateDeletePublicPlanProposal = "op_weight_delete_public_plan_proposal"
	OpWeightSimulateAdvanceEpoch             = "op_weight_advance_epoch"
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
			SimulateModifyPublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateDeletePublicPlanProposal,
			params.DefaultWeightDeletePublicPlanProposal,
			SimulateDeletePublicPlanProposal(ak, bk, k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateAdvanceEpoch,
			params.DefaultWeightAdvanceEpoch,
			SimulateAdvanceEpoch(k, bk),
		),
	}
}

// SimulateAddPublicPlanProposal generates random public add plan proposal content.
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

		addPlanReqs := ranAddPlanRequests(r, ctx, simAccount, poolCoins)

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			addPlanReqs,
			[]types.ModifyPlanRequest{},
			[]types.DeletePlanRequest{},
		)
	}
}

// SimulateModifyPublicPlanProposal generates random public modify plan proposal content.
func SimulateModifyPublicPlanProposal(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.ContentSimulatorFn {
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

		req := types.ModifyPlanRequest{}

		plans := k.GetPlans(ctx)
		for _, p := range plans {
			if p.GetType() == types.PlanTypePublic {
				startTime := ctx.BlockTime()
				endTime := startTime.AddDate(0, simtypes.RandIntBetween(r, 1, 28), 0)

				switch plan := p.(type) {
				case *types.FixedAmountPlan:
					req.PlanId = plan.GetId()
					req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochAmount = sdk.NewCoins(
						sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))),
					)
				case *types.RatioPlan:
					req.PlanId = plan.GetId()
					req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
					req.FarmingPoolAddress = plan.GetFarmingPoolAddress().String()
					req.TerminationAddress = plan.GetTerminationAddress().String()
					req.StakingCoinWeights = plan.GetStakingCoinWeights()
					req.StartTime = &startTime
					req.EndTime = &endTime
					req.EpochRatio = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 5)), 1)
				}
				break
			}
		}

		if req.PlanId == 0 {
			return nil
		}

		modifyPlanReqs := []types.ModifyPlanRequest{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]types.AddPlanRequest{},
			modifyPlanReqs,
			[]types.DeletePlanRequest{},
		)
	}
}

// SimulateDeletePublicPlanProposal generates random public delete plan proposal content.
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

		req := types.DeletePlanRequest{}

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

		deletePlanReqs := []types.DeletePlanRequest{req}

		return types.NewPublicPlanProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			[]types.AddPlanRequest{},
			[]types.ModifyPlanRequest{},
			deletePlanReqs,
		)
	}
}

// SimulateAdvanceEpoch manually advance epoch for simulation palns
func SimulateAdvanceEpoch(k keeper.Keeper, bk types.BankKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		plans := k.GetPlans(ctx)
		plansLen := len(plans)
		for i := 0; i < plansLen; i++ {
			plan := plans[rand.Intn(plansLen)]
			if plan.GetType() == types.PlanTypePrivate {
				mintCoins := sdk.NewCoins(sdk.NewInt64Coin("farmingreward", int64(simtypes.RandIntBetween(r, 0, 1e15))))
				if err := bk.MintCoins(ctx, liquiditytypes.ModuleName, mintCoins); err != nil {
					return nil
				}
				if err := bk.SendCoinsFromModuleToAccount(ctx, liquiditytypes.ModuleName, plan.GetFarmingPoolAddress(), mintCoins); err != nil {
					return nil
				}
				fmt.Println("[minted farming pool]", mintCoins, plan.String())
				break
			}

		}
		err := k.AdvanceEpoch(ctx)
		if err != nil {
			return nil
		}
		fmt.Println("[AdvanceEpoch]")
		return nil
	}
}

// ranAddPlanRequests returns randomized add request proposals.
func ranAddPlanRequests(r *rand.Rand, ctx sdk.Context, simAccount simtypes.Account, poolCoins sdk.Coins) []types.AddPlanRequest {
	ranProposals := make([]types.AddPlanRequest, 0)

	// Generate a random number of proposals with random values of each parameter
	for i := 0; i < simtypes.RandIntBetween(r, 1, 3); i++ {
		req := types.AddPlanRequest{}
		req.Name = "simulation-test-" + simtypes.RandStringOfLength(r, 5)
		req.FarmingPoolAddress = simAccount.Address.String()
		req.TerminationAddress = simAccount.Address.String()
		req.StakingCoinWeights = sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
		req.StartTime = ctx.BlockTime()
		req.EndTime = ctx.BlockTime().AddDate(0, simtypes.RandIntBetween(r, 1, 28), 0)

		// Generate a fixed amount plan if pseudo-random integer is an even number and
		// generate a ratio plan if it is an odd number
		if r.Int()%2 == 0 {
			req.EpochAmount = sdk.NewCoins(
				sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 100_000_000))),
			)
		} else {
			req.EpochRatio = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 2) // 1% ~ 10%
		}
		ranProposals = append(ranProposals, req)
	}
	return ranProposals
}
