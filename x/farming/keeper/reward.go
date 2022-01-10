package keeper

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/crescent-network/crescent/x/farming/types"
)

// GetHistoricalRewards returns historical rewards for a given
// staking coin denom and an epoch number.
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

// SetHistoricalRewards sets historical rewards for a given
// staking coin denom and an epoch number.
func (k Keeper) SetHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch), bz)
}

// DeleteHistoricalRewards deletes historical rewards for a given
// staking coin denom and an epoch number.
func (k Keeper) DeleteHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch))
}

// DeleteAllHistoricalRewards deletes all historical rewards for a
// staking coin denom.
func (k Keeper) DeleteAllHistoricalRewards(ctx sdk.Context, stakingCoinDenom string) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetHistoricalRewardsPrefix(stakingCoinDenom))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// IterateHistoricalRewards iterates through all historical rewards
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
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

// GetCurrentEpoch returns the current epoch number for a given
// staking coin denom.
func (k Keeper) GetCurrentEpoch(ctx sdk.Context, stakingCoinDenom string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCurrentEpochKey(stakingCoinDenom))
	var val gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &val)
	return val.GetValue()
}

// SetCurrentEpoch sets the current epoch number for a given
// staking coin denom.
func (k Keeper) SetCurrentEpoch(ctx sdk.Context, stakingCoinDenom string, currentEpoch uint64) {
	store := ctx.KVStore(k.storeKey)
	val := gogotypes.UInt64Value{Value: currentEpoch}
	bz := k.cdc.MustMarshal(&val)
	store.Set(types.GetCurrentEpochKey(stakingCoinDenom), bz)
}

// DeleteCurrentEpoch deletes current epoch info for a given
// staking coin denom.
func (k Keeper) DeleteCurrentEpoch(ctx sdk.Context, stakingCoinDenom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetCurrentEpochKey(stakingCoinDenom))
}

// IterateCurrentEpochs iterates through all current epoch infos
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
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

// GetOutstandingRewards returns outstanding rewards for a given
// staking coin denom.
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

// SetOutstandingRewards sets outstanding rewards for a given
// staking coin denom.
func (k Keeper) SetOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, rewards types.OutstandingRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetOutstandingRewardsKey(stakingCoinDenom), bz)
}

// DeleteOutstandingRewards deletes outstanding rewards for a given
// staking coin denom.
func (k Keeper) DeleteOutstandingRewards(ctx sdk.Context, stakingCoinDenom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOutstandingRewardsKey(stakingCoinDenom))
}

// IterateOutstandingRewards iterates through all outstanding rewards
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
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

// IncreaseOutstandingRewards increases outstanding rewards for a given
// staking coin denom by given amount.
func (k Keeper) IncreaseOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, amount sdk.DecCoins) {
	outstanding, found := k.GetOutstandingRewards(ctx, stakingCoinDenom)
	if !found {
		panic("outstanding rewards not found")
	}
	outstanding.Rewards = outstanding.Rewards.Add(amount...)
	k.SetOutstandingRewards(ctx, stakingCoinDenom, outstanding)
}

// DecreaseOutstandingRewards decreases outstanding rewards for a given
// staking coin denom by given amount.
// If the resulting outstanding rewards is zero, then the outstanding rewards
// will be deleted, not updated.
func (k Keeper) DecreaseOutstandingRewards(ctx sdk.Context, stakingCoinDenom string, amount sdk.DecCoins) {
	outstanding, found := k.GetOutstandingRewards(ctx, stakingCoinDenom)
	if !found {
		panic("outstanding rewards not found")
	}
	outstanding.Rewards = outstanding.Rewards.Sub(amount)
	k.SetOutstandingRewards(ctx, stakingCoinDenom, outstanding)
}

// CalculateRewards returns rewards accumulated until endingEpoch
// for a farmer for a given staking coin denom.
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

// Rewards returns truncated rewards accumulated until the current epoch
// for a farmer for a given staking coin denom.
func (k Keeper) Rewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) sdk.Coins {
	currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
	rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, currentEpoch-1)
	truncatedRewards, _ := rewards.TruncateDecimal()

	return truncatedRewards
}

// AllRewards returns truncated total rewards accumulated until the
// current epoch for a farmer.
func (k Keeper) AllRewards(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins {
	totalRewards := sdk.NewCoins()
	k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		rewards := k.Rewards(ctx, farmerAcc, stakingCoinDenom)
		totalRewards = totalRewards.Add(rewards...)
		return false
	})
	return totalRewards
}

// WithdrawRewards withdraws accumulated rewards for a farmer for a given
// staking coin denom.
// It decreases outstanding rewards and set the starting epoch of a
// staking.
func (k Keeper) WithdrawRewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) (sdk.Coins, error) {
	staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
	if !found {
		return nil, types.ErrStakingNotExists
	}

	currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
	rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, currentEpoch-1)
	truncatedRewards, _ := rewards.TruncateDecimal()

	if !rewards.IsZero() {
		if !truncatedRewards.IsZero() {
			if err := k.bankKeeper.SendCoins(ctx, types.RewardsReserveAcc, farmerAcc, truncatedRewards); err != nil {
				return nil, err
			}

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeRewardsWithdrawn,
					sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
					sdk.NewAttribute(types.AttributeKeyStakingCoinDenom, stakingCoinDenom),
					sdk.NewAttribute(types.AttributeKeyRewardCoins, truncatedRewards.String()),
				),
			})
		}

		k.DecreaseOutstandingRewards(ctx, stakingCoinDenom, rewards)
	}

	staking.StartingEpoch = currentEpoch
	k.SetStaking(ctx, stakingCoinDenom, farmerAcc, staking)

	return truncatedRewards, nil
}

// WithdrawAllRewards withdraws all accumulated rewards for a farmer.
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
		if err := k.bankKeeper.SendCoins(ctx, types.RewardsReserveAcc, farmerAcc, totalRewards); err != nil {
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
			sdk.NewAttribute(types.AttributeKeyRewardCoins, totalRewards.String()),
		),
	})

	return nil
}

// AllocationInfo holds information about an allocation for a plan.
type AllocationInfo struct {
	Plan   types.PlanI
	Amount sdk.Coins
}

// AllocationInfos returns allocation infos for the end
// of the current epoch.
// When total allocated coins for a farming pool exceeds the pool's
// balance, then allocation will not happen.
func (k Keeper) AllocationInfos(ctx sdk.Context) []AllocationInfo {
	// farmingPoolBalances is a cache for balances of each farming pool,
	// to reduce number of BankKeeper.GetAllBalances calls.
	// It maps farmingPoolAddress to the pool's balance.
	farmingPoolBalances := map[string]sdk.Coins{}

	// allocCoins is a table that records which farming pool allocates
	// how many coins to which plan.
	// It maps farmingPoolAddress to a map that maps planId to amount of
	// coins to allocate.
	allocCoins := map[string]map[uint64]sdk.Coins{}

	plans := map[uint64]types.PlanI{} // it maps planId to plan.
	for _, plan := range k.GetPlans(ctx) {
		// Add plans that are not terminated and active to the map.
		if !plan.GetTerminated() && types.IsPlanActiveAt(plan, ctx.BlockTime()) {
			plans[plan.GetId()] = plan
		}
	}

	// Calculate how many coins the plans want to allocate rewards from farming pools.
	// Note that in this step, we don't check if the farming pool has
	// sufficient balance for all allocations. We'll do that check in the next step.
	for _, plan := range plans {
		farmingPoolAcc := plan.GetFarmingPoolAddress()
		farmingPool := farmingPoolAcc.String()

		// Lookup if we already have the farming pool's balance in the cache.
		// If not, call BankKeeper.GetAllBalances and add the result to the cache.
		balances, ok := farmingPoolBalances[farmingPool]
		if !ok {
			balances = k.bankKeeper.GetAllBalances(ctx, farmingPoolAcc)
			farmingPoolBalances[farmingPool] = balances
		}

		// Lookup if we already have the farming pool's allocation map in the cache.
		// If not, create new allocation map and add it to the cache.
		ac, ok := allocCoins[farmingPool]
		if !ok {
			ac = map[uint64]sdk.Coins{}
			allocCoins[farmingPool] = ac
		}

		// Based on the plan's type, record how many coins the plan wants to
		// allocate in the allocation map.
		switch plan := plan.(type) {
		case *types.FixedAmountPlan:
			ac[plan.GetId()] = plan.EpochAmount
		case *types.RatioPlan:
			ac[plan.GetId()], _ = sdk.NewDecCoinsFromCoins(balances...).MulDecTruncate(plan.EpochRatio).TruncateDecimal()
		}
	}

	// In this step, we check if farming pools have sufficient balance for allocations.
	// If not, we don't allocate rewards from that farming pool for this epoch.
	var allocInfos []AllocationInfo
	for farmingPool, planCoins := range allocCoins {
		totalCoins := sdk.NewCoins()
		for _, amt := range planCoins {
			totalCoins = totalCoins.Add(amt...)
		}

		balances := farmingPoolBalances[farmingPool]
		if !totalCoins.IsAllLTE(balances) {
			continue
		}

		for planID, amt := range planCoins {
			allocInfos = append(allocInfos, AllocationInfo{
				Plan:   plans[planID],
				Amount: amt,
			})
		}
	}

	return allocInfos
}

// AllocateRewards updates historical rewards and current epoch info
// based on the allocation infos.
func (k Keeper) AllocateRewards(ctx sdk.Context) error {
	// unitRewardsByDenom is a table that records how much unit rewards should
	// be increased in this epoch, for each staking coin denom.
	// It maps staking coin denom to unit rewards.
	unitRewardsByDenom := map[string]sdk.DecCoins{}

	// Get allocation information first.
	allocInfos := k.AllocationInfos(ctx)

	for _, allocInfo := range allocInfos {
		totalAllocCoins := sdk.NewCoins()

		// Calculate how many coins are allocated based on each staking coin weight.
		// It is calculated with the following formula:
		// (unit rewards for this epoch) = (weighted rewards for the denom) / (total staking amount for the denom)
		for _, weight := range allocInfo.Plan.GetStakingCoinWeights() {
			// Check if there are any coins staked for this denom.
			// If not, skip this denom for rewards allocation.
			totalStakings, found := k.GetTotalStakings(ctx, weight.Denom)
			if !found {
				continue
			}

			allocCoins, _ := sdk.NewDecCoinsFromCoins(allocInfo.Amount...).MulDecTruncate(weight.Amount).TruncateDecimal()
			allocCoinsDec := sdk.NewDecCoinsFromCoins(allocCoins...)

			// Multiple plans can have same denom in their staking coin weights,
			// so we accumulate all unit rewards for this denom in the table.
			unitRewardsByDenom[weight.Denom] = unitRewardsByDenom[weight.Denom].Add(allocCoinsDec.QuoDecTruncate(totalStakings.Amount.ToDec())...)

			k.IncreaseOutstandingRewards(ctx, weight.Denom, allocCoinsDec)

			totalAllocCoins = totalAllocCoins.Add(allocCoins...)
		}

		// If total allocated amount for this plan is zero, then skip allocation
		// for this plan.
		if totalAllocCoins.IsZero() {
			continue
		}

		rewardsReserveAcc := types.RewardsReserveAcc
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

	// For each staking coin denom in the table, increase cumulative unit rewards
	// and increment current epoch number by 1.
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

// ValidateRemainingRewardsAmount checks that the balance of the
// rewards reserve pool is greater than the total amount of
// unwithdrawn rewards.
func (k Keeper) ValidateRemainingRewardsAmount(ctx sdk.Context) error {
	remainingRewards := sdk.NewCoins()
	k.IterateStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool) {
		rewards := k.Rewards(ctx, farmerAcc, stakingCoinDenom)
		remainingRewards = remainingRewards.Add(rewards...)
		return false
	})

	rewardsReservePoolBalances := k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc)
	if !rewardsReservePoolBalances.IsAllGTE(remainingRewards) {
		return types.ErrInvalidRemainingRewardsAmount
	}

	return nil
}

// ValidateOutstandingRewardsAmount checks that the balance of the
// rewards reserve pool is greater than the total amount of
// outstanding rewards.
func (k Keeper) ValidateOutstandingRewardsAmount(ctx sdk.Context) error {
	totalOutstandingRewards := sdk.NewDecCoins()
	k.IterateOutstandingRewards(ctx, func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool) {
		totalOutstandingRewards = totalOutstandingRewards.Add(rewards.Rewards...)
		return false
	})

	rewardsReservePoolBalances := sdk.NewDecCoinsFromCoins(k.bankKeeper.GetAllBalances(ctx, types.RewardsReserveAcc)...)
	_, hasNeg := rewardsReservePoolBalances.SafeSub(totalOutstandingRewards)
	if hasNeg {
		return types.ErrInvalidOutstandingRewardsAmount
	}

	return nil
}
