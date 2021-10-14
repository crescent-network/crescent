package keeper

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (k Keeper) GetHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64) (rewards types.HistoricalRewards, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &rewards)
	found = true
	return
}

func (k Keeper) SetHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch), bz)
}

func (k Keeper) DeleteHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch))
}

func (k Keeper) IterateHistoricalRewards(ctx sdk.Context, cb func(stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.HistoricalRewardsKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.HistoricalRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		stakingCoinDenom, epoch := types.ParseHistoricalRewardsKey(iter.Key())
		if cb(stakingCoinDenom, epoch, rewards) {
			break
		}
	}
}

func (k Keeper) GetCurrentEpoch(ctx sdk.Context, stakingCoinDenom string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCurrentEpochKey(stakingCoinDenom))
	var val gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &val)
	return val.GetValue()
}

func (k Keeper) SetCurrentEpoch(ctx sdk.Context, stakingCoinDenom string, currentEpoch uint64) {
	store := ctx.KVStore(k.storeKey)
	val := gogotypes.UInt64Value{Value: currentEpoch}
	bz := k.cdc.MustMarshal(&val)
	store.Set(types.GetCurrentEpochKey(stakingCoinDenom), bz)
}

func (k Keeper) IterateCurrentEpochs(ctx sdk.Context, cb func(stakingCoinDenom string, currentEpoch uint64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.CurrentEpochKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var val gogotypes.UInt64Value
		k.cdc.MustUnmarshal(iter.Value(), &val)
		stakingCoinDenom := types.ParseCurrentEpochKey(iter.Key())
		if cb(stakingCoinDenom, val.GetValue()) {
			break
		}
	}
}

func (k Keeper) GetOutstandingRewards(ctx sdk.Context, stakingCoinDenom string) (rewards types.OutstandingRewards, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOutstandingRewardsKey(stakingCoinDenom))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &rewards)
	found = true
	return
}

func (k Keeper) SetOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, rewards types.OutstandingRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetOutstandingRewardsKey(stakingCoinDenom), bz)
}

func (k Keeper) DeleteOutstandingRewards(ctx sdk.Context, stakingCoinDenom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOutstandingRewardsKey(stakingCoinDenom))
}

func (k Keeper) IterateOutstandingRewards(ctx sdk.Context, cb func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.OutstandingRewardsKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.OutstandingRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		stakingCoinDenom := types.ParseOutstandingRewardsKey(iter.Key())
		if cb(stakingCoinDenom, rewards) {
			break
		}
	}
}

func (k Keeper) IncreaseOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, amount sdk.DecCoins) {
	outstanding, _ := k.GetOutstandingRewards(ctx, stakingCoinDenom)
	outstanding.Rewards = outstanding.Rewards.Add(amount...)
	k.SetOutstandingRewards(ctx, stakingCoinDenom, outstanding)
}

func (k Keeper) DecreaseOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, amount sdk.DecCoins) {
	outstanding, found := k.GetOutstandingRewards(ctx, stakingCoinDenom)
	if !found {
		panic("outstanding rewards not found")
	}
	if outstanding.Rewards.IsEqual(amount) {
		k.DeleteOutstandingRewards(ctx, stakingCoinDenom)
	} else {
		outstanding.Rewards = outstanding.Rewards.Sub(amount)
		k.SetOutstandingRewards(ctx, stakingCoinDenom, outstanding)
	}
}

func (k Keeper) CalculateRewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string, endingEpoch uint64) (rewards sdk.DecCoins) {
	staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
	if !found {
		return sdk.NewDecCoins()
	}

	starting, _ := k.GetHistoricalRewards(ctx, stakingCoinDenom, staking.StartingEpoch-1)
	ending, _ := k.GetHistoricalRewards(ctx, stakingCoinDenom, endingEpoch)
	diff := ending.CumulativeUnitRewards.Sub(starting.CumulativeUnitRewards)
	rewards = diff.MulDecTruncate(staking.Amount.ToDec())
	return
}

func (k Keeper) Rewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) sdk.Coins {
	currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
	rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, currentEpoch-1)
	truncatedRewards, _ := rewards.TruncateDecimal()

	return truncatedRewards
}

func (k Keeper) AllRewards(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins {
	totalRewards := sdk.NewCoins()
	k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		rewards := k.Rewards(ctx, farmerAcc, stakingCoinDenom)
		totalRewards = totalRewards.Add(rewards...)
		return false
	})
	return totalRewards
}

func (k Keeper) WithdrawRewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) (sdk.Coins, error) {
	staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
	if !found {
		return nil, types.ErrStakingNotExists
	}

	currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
	// TODO: handle if currentEpoch is 0
	rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, currentEpoch-1)
	truncatedRewards, _ := rewards.TruncateDecimal()

	if !rewards.IsZero() {
		if !truncatedRewards.IsZero() {
			if err := k.bankKeeper.SendCoins(ctx, k.GetRewardsReservePoolAcc(ctx), farmerAcc, truncatedRewards); err != nil {
				return nil, err
			}
		}

		k.DecreaseOutstandingRewards(ctx, stakingCoinDenom, rewards)
	}

	staking.StartingEpoch = currentEpoch
	k.SetStaking(ctx, stakingCoinDenom, farmerAcc, staking)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRewardsWithdrawn,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStakingCoinDenom, stakingCoinDenom),
			sdk.NewAttribute(types.AttributeKeyRewardCoins, truncatedRewards.String()),
		),
	})

	return truncatedRewards, nil
}

func (k Keeper) WithdrawAllRewards(ctx sdk.Context, farmerAcc sdk.AccAddress) (sdk.Coins, error) {
	totalRewards := sdk.NewCoins()
	k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
		rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, currentEpoch-1)
		truncatedRewards, _ := rewards.TruncateDecimal()
		totalRewards = totalRewards.Add(truncatedRewards...)

		if !rewards.IsZero() {
			k.DecreaseOutstandingRewards(ctx, stakingCoinDenom, rewards)
		}

		staking.StartingEpoch = currentEpoch
		k.SetStaking(ctx, stakingCoinDenom, farmerAcc, staking)

		return false
	})

	if !totalRewards.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, k.GetRewardsReservePoolAcc(ctx), farmerAcc, totalRewards); err != nil {
			return nil, err
		}
	}

	return totalRewards, nil
}

// Harvest claims farming rewards from the reward pool.
func (k Keeper) Harvest(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenoms []string) error {
	totalRewards := sdk.NewCoins()

	for _, denom := range stakingCoinDenoms {
		rewards, err := k.WithdrawRewards(ctx, farmerAcc, denom)
		if err != nil {
			return err
		}
		totalRewards = totalRewards.Add(rewards...)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeHarvest,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStakingCoinDenoms, strings.Join(stakingCoinDenoms, ",")),
		),
	})

	return nil
}

type AllocationInfo struct {
	Plan   types.PlanI
	Amount sdk.Coins
}

func (k Keeper) AllocationInfos(ctx sdk.Context) []AllocationInfo {
	farmingPoolBalances := make(map[string]sdk.Coins)   // farmingPoolAddress => sdk.Coins
	allocCoins := make(map[string]map[uint64]sdk.Coins) // farmingPoolAddress => (planId => sdk.Coins)

	plans := make(map[uint64]types.PlanI)
	for _, plan := range k.GetPlans(ctx) {
		// Filter plans by their start time and end time.
		if !plan.GetTerminated() && types.IsPlanActiveAt(plan, ctx.BlockTime()) {
			plans[plan.GetId()] = plan
		}
	}

	for _, plan := range plans {
		farmingPoolAcc := plan.GetFarmingPoolAddress()
		farmingPool := farmingPoolAcc.String()

		balances, ok := farmingPoolBalances[farmingPool]
		if !ok {
			balances = k.bankKeeper.GetAllBalances(ctx, farmingPoolAcc)
			farmingPoolBalances[farmingPool] = balances
		}

		ac, ok := allocCoins[farmingPool]
		if !ok {
			ac = make(map[uint64]sdk.Coins)
			allocCoins[farmingPool] = ac
		}

		switch plan := plan.(type) {
		case *types.FixedAmountPlan:
			ac[plan.GetId()] = plan.EpochAmount
		case *types.RatioPlan:
			ac[plan.GetId()], _ = sdk.NewDecCoinsFromCoins(balances...).MulDecTruncate(plan.EpochRatio).TruncateDecimal()
		}
	}

	var allocInfos []AllocationInfo
	for farmingPool, coins := range allocCoins {
		totalCoins := sdk.NewCoins()
		for _, amt := range coins {
			totalCoins = totalCoins.Add(amt...)
		}

		balances := farmingPoolBalances[farmingPool]
		if !totalCoins.IsAllLTE(balances) {
			continue
		}

		for planID, amt := range coins {
			allocInfos = append(allocInfos, AllocationInfo{
				Plan:   plans[planID],
				Amount: amt,
			})
		}
	}

	return allocInfos
}

func (k Keeper) AllocateRewards(ctx sdk.Context) error {
	unitRewardsByDenom := map[string]sdk.DecCoins{} // (staking coin denom) => (unit rewards)

	for _, allocInfo := range k.AllocationInfos(ctx) {
		totalWeight := sdk.ZeroDec()
		for _, weight := range allocInfo.Plan.GetStakingCoinWeights() {
			totalWeight = totalWeight.Add(weight.Amount)
		}

		totalAllocCoins := sdk.NewCoins()
		for _, weight := range allocInfo.Plan.GetStakingCoinWeights() {
			totalStakings, found := k.GetTotalStakings(ctx, weight.Denom)
			if !found {
				continue
			}
			if !totalStakings.Amount.IsPositive() {
				continue
			}

			weightProportion := weight.Amount.QuoTruncate(totalWeight)
			allocCoins, _ := sdk.NewDecCoinsFromCoins(allocInfo.Amount...).MulDecTruncate(weightProportion).TruncateDecimal()
			allocCoinsDec := sdk.NewDecCoinsFromCoins(allocCoins...)

			unitRewardsByDenom[weight.Denom] = unitRewardsByDenom[weight.Denom].Add(allocCoinsDec.QuoDecTruncate(totalStakings.Amount.ToDec())...)

			k.IncreaseOutstandingRewards(ctx, weight.Denom, allocCoinsDec)

			totalAllocCoins = totalAllocCoins.Add(allocCoins...)
		}

		if totalAllocCoins.IsZero() {
			continue
		}

		rewardsReserveAcc := k.GetRewardsReservePoolAcc(ctx)
		if err := k.bankKeeper.SendCoins(ctx, allocInfo.Plan.GetFarmingPoolAddress(), rewardsReserveAcc, totalAllocCoins); err != nil {
			return err
		}

		t := ctx.BlockTime()
		_ = allocInfo.Plan.SetLastDistributionTime(&t)
		_ = allocInfo.Plan.SetDistributedCoins(allocInfo.Plan.GetDistributedCoins().Add(totalAllocCoins...))
		k.SetPlan(ctx, allocInfo.Plan)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRewardsAllocated,
				sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(allocInfo.Plan.GetId(), 10)),
				sdk.NewAttribute(types.AttributeKeyAmount, totalAllocCoins.String()),
			),
		})
	}

	for stakingCoinDenom, unitRewards := range unitRewardsByDenom {
		currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
		historical, _ := k.GetHistoricalRewards(ctx, stakingCoinDenom, currentEpoch-1)
		k.SetHistoricalRewards(ctx, stakingCoinDenom, currentEpoch, types.HistoricalRewards{
			CumulativeUnitRewards: historical.CumulativeUnitRewards.Add(unitRewards...),
		})
		k.SetCurrentEpoch(ctx, stakingCoinDenom, currentEpoch+1)
	}

	return nil
}

// ValidateRemainingRewardsAmount checks that the balance of the RewardPoolAddresses of all plans greater than the total amount of unwithdrawn reward coins in all reward objects
// TODO: correct comment above
func (k Keeper) ValidateRemainingRewardsAmount(ctx sdk.Context) error {
	remainingRewards := sdk.NewCoins()
	k.IterateStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool) {
		rewards := k.Rewards(ctx, farmerAcc, stakingCoinDenom)
		remainingRewards = remainingRewards.Add(rewards...)
		return false
	})

	rewardsReservePoolBalances := k.bankKeeper.GetAllBalances(ctx, k.GetRewardsReservePoolAcc(ctx))
	if !rewardsReservePoolBalances.IsAllGTE(remainingRewards) {
		return types.ErrInvalidRemainingRewardsAmount
	}

	return nil
}

func (k Keeper) ValidateOutstandingRewards(ctx sdk.Context) error {
	totalOutstandingRewards := sdk.NewDecCoins()
	k.IterateOutstandingRewards(ctx, func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool) {
		totalOutstandingRewards = totalOutstandingRewards.Add(rewards.Rewards...)
		return false
	})

	rewardsReservePoolBalances := sdk.NewDecCoinsFromCoins(k.bankKeeper.GetAllBalances(ctx, k.GetRewardsReservePoolAcc(ctx))...)
	_, hasNeg := rewardsReservePoolBalances.SafeSub(totalOutstandingRewards)
	if hasNeg {
		return types.ErrInvalidStakingReservedAmount // TODO: use different error type
	}

	return nil
}
