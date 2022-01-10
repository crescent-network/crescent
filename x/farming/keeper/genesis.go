package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/farming/types"
)

// InitGenesis initializes the farming module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := types.ValidateGenesis(genState); err != nil {
		panic(err)
	}

	ctx, writeCache := ctx.CacheContext()

	k.SetParams(ctx, genState.Params)
	k.SetCurrentEpochDays(ctx, genState.CurrentEpochDays)
	if addr := k.accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	for i, record := range genState.PlanRecords {
		plan, err := types.UnpackPlan(&record.Plan)
		if err != nil {
			panic(err)
		}
		k.SetPlan(ctx, plan)
		if i == len(genState.PlanRecords)-1 {
			k.SetGlobalPlanId(ctx, plan.GetId())
		}
	}

	totalStakings := map[string]sdk.Int{} // (staking coin denom) => (amount)

	for _, record := range genState.StakingRecords {
		farmerAcc, err := sdk.AccAddressFromBech32(record.Farmer)
		if err != nil {
			panic(err)
		}
		k.SetStaking(ctx, record.StakingCoinDenom, farmerAcc, record.Staking)

		amt, ok := totalStakings[record.StakingCoinDenom]
		if !ok {
			amt = sdk.ZeroInt()
		}
		amt = amt.Add(record.Staking.Amount)
		totalStakings[record.StakingCoinDenom] = amt
	}

	for _, record := range genState.TotalStakingsRecords {
		if !record.Amount.Equal(totalStakings[record.StakingCoinDenom]) {
			panic(fmt.Sprintf("TotalStaking for %s differs from the actual value; have %s, want %s",
				record.StakingCoinDenom, totalStakings[record.StakingCoinDenom], record.Amount))
		}
		stakingReserveCoins := k.bankKeeper.GetAllBalances(ctx, types.StakingReserveAcc(record.StakingCoinDenom))
		if !record.StakingReserveCoins.IsEqual(stakingReserveCoins) {
			panic(fmt.Sprintf("StakingReserveCoins differs from the actual value; have %s, want %s",
				stakingReserveCoins, record.StakingReserveCoins))
		}
	}

	if len(totalStakings) != len(genState.TotalStakingsRecords) {
		panic(fmt.Sprintf("the number of TotalStaking differs from the actual value; have %d, want %d",
			len(totalStakings), len(genState.TotalStakingsRecords)))
	}

	for _, record := range genState.TotalStakingsRecords {
		k.SetTotalStakings(ctx, record.StakingCoinDenom, types.TotalStakings{Amount: record.Amount})
	}

	for _, record := range genState.QueuedStakingRecords {
		farmerAcc, err := sdk.AccAddressFromBech32(record.Farmer)
		if err != nil {
			panic(err)
		}
		k.SetQueuedStaking(ctx, record.StakingCoinDenom, farmerAcc, record.QueuedStaking)
	}

	for _, record := range genState.HistoricalRewardsRecords {
		k.SetHistoricalRewards(ctx, record.StakingCoinDenom, record.Epoch, record.HistoricalRewards)
	}

	for _, record := range genState.OutstandingRewardsRecords {
		k.SetOutstandingRewards(ctx, record.StakingCoinDenom, record.OutstandingRewards)
	}

	for _, record := range genState.CurrentEpochRecords {
		k.SetCurrentEpoch(ctx, record.StakingCoinDenom, record.CurrentEpoch)
	}

	if genState.LastEpochTime != nil {
		k.SetLastEpochTime(ctx, *genState.LastEpochTime)
	}

	err := k.ValidateRemainingRewardsAmount(ctx)
	if err != nil {
		panic(err)
	}
	rewardsPoolCoins := k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc)
	if !genState.RewardPoolCoins.IsEqual(rewardsPoolCoins) {
		panic(fmt.Sprintf("RewardPoolCoins differs from the actual value; have %s, want %s",
			rewardsPoolCoins, genState.RewardPoolCoins))
	}

	err = k.ValidateStakingReservedAmount(ctx)
	if err != nil {
		panic(err)
	}

	if err := k.ValidateOutstandingRewardsAmount(ctx); err != nil {
		panic(err)
	}

	writeCache()
}

// ExportGenesis returns the farming module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	plans := []types.PlanRecord{}
	for _, plan := range k.GetPlans(ctx) {
		any, err := types.PackPlan(plan)
		if err != nil {
			panic(err)
		}
		plans = append(plans, types.PlanRecord{
			Plan:             *any,
			FarmingPoolCoins: k.bankKeeper.GetAllBalances(ctx, plan.GetFarmingPoolAddress()),
		})
	}

	stakings := []types.StakingRecord{}
	k.IterateStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool) {
		stakings = append(stakings, types.StakingRecord{
			StakingCoinDenom: stakingCoinDenom,
			Farmer:           farmerAcc.String(),
			Staking:          staking,
		})
		return false
	})

	queuedStakings := []types.QueuedStakingRecord{}
	k.IterateQueuedStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		queuedStakings = append(queuedStakings, types.QueuedStakingRecord{
			StakingCoinDenom: stakingCoinDenom,
			Farmer:           farmerAcc.String(),
			QueuedStaking:    queuedStaking,
		})
		return false
	})

	totalStakings := []types.TotalStakingsRecord{}
	k.IterateTotalStakings(ctx, func(stakingCoinDenom string, ts types.TotalStakings) (stop bool) {
		totalStakings = append(totalStakings, types.TotalStakingsRecord{
			StakingCoinDenom:    stakingCoinDenom,
			Amount:              ts.Amount,
			StakingReserveCoins: k.bankKeeper.GetAllBalances(ctx, types.StakingReserveAcc(stakingCoinDenom)),
		})
		return false
	})

	historicalRewards := []types.HistoricalRewardsRecord{}
	k.IterateHistoricalRewards(ctx, func(stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) (stop bool) {
		historicalRewards = append(historicalRewards, types.HistoricalRewardsRecord{
			StakingCoinDenom:  stakingCoinDenom,
			Epoch:             epoch,
			HistoricalRewards: rewards,
		})
		return false
	})

	outstandingRewards := []types.OutstandingRewardsRecord{}
	k.IterateOutstandingRewards(ctx, func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool) {
		outstandingRewards = append(outstandingRewards, types.OutstandingRewardsRecord{
			StakingCoinDenom:   stakingCoinDenom,
			OutstandingRewards: rewards,
		})
		return false
	})

	currentEpochs := []types.CurrentEpochRecord{}
	k.IterateCurrentEpochs(ctx, func(stakingCoinDenom string, currentEpoch uint64) (stop bool) {
		currentEpochs = append(currentEpochs, types.CurrentEpochRecord{
			StakingCoinDenom: stakingCoinDenom,
			CurrentEpoch:     currentEpoch,
		})
		return false
	})

	var epochTime *time.Time
	tempEpochTime, found := k.GetLastEpochTime(ctx)
	if found {
		epochTime = &tempEpochTime
	}

	return types.NewGenesisState(
		params,
		plans,
		stakings,
		queuedStakings,
		totalStakings,
		historicalRewards,
		outstandingRewards,
		currentEpochs,
		k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc),
		epochTime,
		k.GetCurrentEpochDays(ctx),
	)
}
