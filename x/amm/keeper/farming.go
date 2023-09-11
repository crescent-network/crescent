package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) CanCreatePrivateFarmingPlan(ctx sdk.Context) bool {
	return k.GetNumPrivateFarmingPlans(ctx) < k.GetMaxNumPrivateFarmingPlans(ctx)
}

func (k Keeper) CreatePrivateFarmingPlan(
	ctx sdk.Context, creatorAddr sdk.AccAddress, description string,
	termAddr sdk.AccAddress, rewardAllocs []types.FarmingRewardAllocation, startTime, endTime time.Time,
) (plan types.FarmingPlan, err error) {
	if !k.CanCreatePrivateFarmingPlan(ctx) {
		return plan, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"maximum number of active private farming plans reached")
	}
	fee := k.GetPrivateFarmingPlanCreationFee(ctx)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, fee); err != nil {
		return
	}
	plan, err = k.createFarmingPlan(ctx, description, nil, termAddr, rewardAllocs, startTime, endTime, true)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventCreatePrivateFarmingPlan{
		Creator:            creatorAddr.String(),
		Description:        description,
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
		FarmingPlanId:      plan.Id,
		FarmingPoolAddress: plan.FarmingPoolAddress,
	}); err != nil {
		return
	}
	return plan, nil
}

func (k Keeper) CreatePublicFarmingPlan(
	ctx sdk.Context, description string,
	farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []types.FarmingRewardAllocation, startTime, endTime time.Time,
) (plan types.FarmingPlan, err error) {
	plan, err = k.createFarmingPlan(
		ctx, description, farmingPoolAddr, termAddr,
		rewardAllocs, startTime, endTime, false)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventCreatePublicFarmingPlan{
		Description:        description,
		FarmingPoolAddress: plan.FarmingPoolAddress,
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
		FarmingPlanId:      plan.Id,
	}); err != nil {
		return
	}
	return plan, nil
}

func (k Keeper) createFarmingPlan(
	ctx sdk.Context, description string, farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []types.FarmingRewardAllocation, startTime, endTime time.Time, isPrivate bool,
) (plan types.FarmingPlan, err error) {
	// Check if end time > block time
	if !endTime.After(ctx.BlockTime()) {
		return plan, sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest, "end time is past")
	}
	for _, rewardAlloc := range rewardAllocs {
		if found := k.LookupPool(ctx, rewardAlloc.PoolId); !found {
			return plan, sdkerrors.Wrapf(
				sdkerrors.ErrNotFound, "pool %d not found", rewardAlloc.PoolId)
		}
		for _, coin := range rewardAlloc.RewardsPerDay {
			if !k.bankKeeper.HasSupply(ctx, coin.Denom) {
				return plan, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidRequest, "denom %s has no supply", coin.Denom)
			}
		}
	}
	// Generate the next plan id and update the last plan id.
	id := k.GetNextFarmingPlanIdWithUpdate(ctx)
	if isPrivate {
		farmingPoolAddr = types.DeriveFarmingPoolAddress(id)
		k.SetNumPrivateFarmingPlans(ctx, k.GetNumPrivateFarmingPlans(ctx)+1)
	}
	plan = types.NewFarmingPlan(
		id, description, farmingPoolAddr, termAddr, rewardAllocs,
		startTime, endTime, isPrivate)
	k.SetFarmingPlan(ctx, plan)
	// TODO: emit event
	return plan, nil
}

func (k Keeper) TerminateEndedFarmingPlans(ctx sdk.Context) (err error) {
	k.IterateAllFarmingPlans(ctx, func(plan types.FarmingPlan) (stop bool) {
		if plan.IsTerminated {
			return false
		}
		if !ctx.BlockTime().Before(plan.EndTime) {
			if err = k.TerminateFarmingPlan(ctx, plan); err != nil {
				return true
			}
		}
		return false
	})
	return err
}

func (k Keeper) TerminateFarmingPlan(ctx sdk.Context, plan types.FarmingPlan) error {
	if plan.IsTerminated {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "plan is already terminated")
	}
	if plan.FarmingPoolAddress != plan.TerminationAddress {
		farmingPoolAddr := sdk.MustAccAddressFromBech32(plan.FarmingPoolAddress)
		balances := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
		if balances.IsAllPositive() {
			if err := k.bankKeeper.SendCoins(
				ctx, farmingPoolAddr,
				sdk.MustAccAddressFromBech32(plan.TerminationAddress), balances); err != nil {
				return err
			}
		}
	}
	plan.IsTerminated = true
	k.SetFarmingPlan(ctx, plan)
	if plan.IsPrivate {
		numPlans := k.GetNumPrivateFarmingPlans(ctx)
		if numPlans > 1 {
			k.SetNumPrivateFarmingPlans(ctx, numPlans-1)
		} else {
			k.DeleteNumPrivateFarmingPlans(ctx)
		}
	}
	if err := ctx.EventManager().EmitTypedEvent(&types.EventFarmingPlanTerminated{
		FarmingPlanId: plan.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) AllocateFarmingRewards(ctx sdk.Context) error {
	lastBlockTime, found := k.markerKeeper.GetLastBlockTime(ctx)
	if !found {
		// For the very first block, just skip it.
		return nil
	}
	elapsed := ctx.BlockTime().Sub(lastBlockTime)
	// Constrain the elapsed block time to the max rewards block time parameter.
	if maxBlockTime := k.GetMaxFarmingBlockTime(ctx); elapsed > maxBlockTime {
		elapsed = maxBlockTime
	}
	// If the elapsed block time is 0, skip this block for rewards allocation.
	if elapsed == 0 {
		return nil
	}
	totalRewardsByFarmingPool := map[string]sdk.Coins{}                // farming pool => rewards
	allocatedRewardsByFarmingPool := map[string]map[uint64]sdk.Coins{} // farming pool => (pool id => rewards)
	var farmingPools []string
	k.IterateAllFarmingPlans(ctx, func(plan types.FarmingPlan) (stop bool) {
		if plan.IsTerminated || !plan.IsActiveAt(ctx.BlockTime()) {
			return false // Skip
		}
		for _, rewardAlloc := range plan.RewardAllocations {
			rewards := types.RewardsForBlock(rewardAlloc.RewardsPerDay, elapsed)
			// TODO: allocate sdk.DecCoins instead of sdk.Coins in future
			truncatedRewards, _ := rewards.TruncateDecimal()
			if truncatedRewards.IsAllPositive() {
				pool := k.MustGetPool(ctx, rewardAlloc.PoolId)
				poolState := k.MustGetPoolState(ctx, pool.Id)
				if !poolState.CurrentLiquidity.IsPositive() {
					continue
				}
				if _, ok := totalRewardsByFarmingPool[plan.FarmingPoolAddress]; !ok {
					farmingPools = append(farmingPools, plan.FarmingPoolAddress)
				}
				totalRewardsByFarmingPool[plan.FarmingPoolAddress] =
					totalRewardsByFarmingPool[plan.FarmingPoolAddress].Add(truncatedRewards...)
				allocatedRewardsByPool, ok := allocatedRewardsByFarmingPool[plan.FarmingPoolAddress]
				if !ok {
					allocatedRewardsByPool = map[uint64]sdk.Coins{}
					allocatedRewardsByFarmingPool[plan.FarmingPoolAddress] = allocatedRewardsByPool
				}
				allocatedRewardsByPool[pool.Id] = allocatedRewardsByPool[pool.Id].Add(truncatedRewards...)
			}
		}
		return false
	})
	totalRewardsByPool := map[uint64]sdk.Coins{}
	var rewardedPools []uint64
	for _, farmingPool := range farmingPools {
		farmingPoolAddr := sdk.MustAccAddressFromBech32(farmingPool)
		spendable := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
		totalRewards := totalRewardsByFarmingPool[farmingPool]
		if !spendable.IsAllGTE(totalRewards) {
			continue
		}
		if err := k.bankKeeper.SendCoins(
			ctx, farmingPoolAddr, types.RewardsPoolAddress, totalRewards); err != nil {
			return err
		}
		for poolId, rewards := range allocatedRewardsByFarmingPool[farmingPool] {
			if _, ok := totalRewardsByPool[poolId]; !ok {
				rewardedPools = append(rewardedPools, poolId)
			}
			totalRewardsByPool[poolId] = totalRewardsByPool[poolId].Add(rewards...)
		}
	}
	for _, poolId := range rewardedPools {
		poolState := k.MustGetPoolState(ctx, poolId)
		rewardsGrowth := sdk.NewDecCoinsFromCoins(totalRewardsByPool[poolId]...).
			MulDecTruncate(types.DecMulFactor).
			QuoDecTruncate(poolState.CurrentLiquidity.ToDec())
		poolState.FarmingRewardsGrowthGlobal = poolState.FarmingRewardsGrowthGlobal.Add(rewardsGrowth...)
		k.SetPoolState(ctx, poolId, poolState)
	}
	return nil
}
