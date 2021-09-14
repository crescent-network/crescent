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

	var plans []types.PlanRecord
	for _, plan := range k.GetAllPlans(ctx) {
		any, err := types.PackPlan(plan)
		if err != nil {
			panic(err)
		}
		plans = append(plans, types.PlanRecord{
			Plan:             *any,
			FarmingPoolCoins: k.bankKeeper.GetAllBalances(ctx, plan.GetFarmingPoolAddress()),
		})
	}

	var stakings []types.StakingRecord
	k.IterateStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool) {
		stakings = append(stakings, types.StakingRecord{
			StakingCoinDenom: stakingCoinDenom,
			Farmer:           farmerAcc.String(),
			Staking:          staking,
		})
		return false
	})

	var queuedStakings []types.QueuedStakingRecord
	k.IterateQueuedStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		queuedStakings = append(queuedStakings, types.QueuedStakingRecord{
			StakingCoinDenom: stakingCoinDenom,
			Farmer:           farmerAcc.String(),
			QueuedStaking:    queuedStaking,
		})
		return false
	})

	var historicalRewards []types.HistoricalRewardsRecord
	k.IterateHistoricalRewards(ctx, func(stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) (stop bool) {
		historicalRewards = append(historicalRewards, types.HistoricalRewardsRecord{
			StakingCoinDenom:  stakingCoinDenom,
			Epoch:             epoch,
			HistoricalRewards: rewards,
		})
		return false
	})

	var currentEpochs []types.CurrentEpochRecord
	k.IterateCurrentEpochs(ctx, func(stakingCoinDenom string, currentEpoch uint64) (stop bool) {
		currentEpochs = append(currentEpochs, types.CurrentEpochRecord{
			StakingCoinDenom: stakingCoinDenom,
			CurrentEpoch:     currentEpoch,
		})
		return false
	})

	epochTime, _ := k.GetLastEpochTime(ctx)
	return types.NewGenesisState(
		params,
		plans,
		stakings,
		queuedStakings,
		historicalRewards,
		currentEpochs,
		k.bankKeeper.GetAllBalances(ctx, types.StakingReserveAcc),
		k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc),
		epochTime)
}
