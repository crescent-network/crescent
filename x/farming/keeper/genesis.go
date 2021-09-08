package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

// InitGenesis initializes the farming module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	ctx, applyCache := ctx.CacheContext()
	k.SetParams(ctx, genState.Params)
	moduleAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	k.accountKeeper.SetModuleAccount(ctx, moduleAcc)

	_, err := sdk.AccAddressFromBech32(genState.Params.FarmingFeeCollector)
	if err != nil {
		panic(err)
	}

	// TODO: add f1 struct, queued staking
	for _, record := range genState.PlanRecords {
		plan, err := types.UnpackPlan(&record.Plan)
		if err != nil {
			panic(err)
		}
		k.SetPlan(ctx, plan)
		k.SetGlobalPlanId(ctx, plan.GetId())
	}
	//for _, staking := range genState.Stakings {
	//	k.SetStaking(ctx, staking)
	//	k.SetStakingIndex(ctx, staking)
	//}
	if err := k.ValidateRemainingRewardsAmount(ctx); err != nil {
		panic(err)
	}
	if err := k.ValidateStakingReservedAmount(ctx); err != nil {
		panic(err)
	}
	applyCache()
}

// ExportGenesis returns the farming module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	// TODO: unimplemented
	var planRecords []types.PlanRecord

	plans := k.GetAllPlans(ctx)
	// TODO: add f1 struct, queued staking
	//stakings := k.GetAllStakings(ctx)

	for _, plan := range plans {
		any, err := types.PackPlan(plan)
		if err != nil {
			panic(err)
		}
		planRecords = append(planRecords, types.PlanRecord{
			Plan:             *any,
			FarmingPoolCoins: k.bankKeeper.GetAllBalances(ctx, plan.GetFarmingPoolAddress()),
		})
	}

	epochTime, _ := k.GetLastEpochTime(ctx)
	return types.NewGenesisState(params, planRecords, nil, k.bankKeeper.GetAllBalances(ctx, types.StakingReserveAcc), k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc), epochTime)
}
