package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

// CreatePrivatePlan creates a new private farming plan.
func (k Keeper) CreatePrivatePlan(
	ctx sdk.Context, creatorAddr sdk.AccAddress, description string,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time,
) (types.Plan, error) {
	if !k.CanCreatePrivatePlan(ctx) {
		return types.Plan{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"maximum number of active private plans reached")
	}

	fee := k.GetPrivatePlanCreationFee(ctx)
	feeCollectorAddr, err := sdk.AccAddressFromBech32(k.GetFeeCollector(ctx))
	if err != nil {
		return types.Plan{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, creatorAddr, feeCollectorAddr, fee); err != nil {
		return types.Plan{}, err
	}

	id, _ := k.GetLastPlanId(ctx)
	farmingPoolAddr := types.DeriveFarmingPoolAddress(id + 1)

	plan, err := k.createPlan(
		ctx, description, farmingPoolAddr, creatorAddr,
		rewardAllocs, startTime, endTime, true)
	if err != nil {
		return types.Plan{}, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCreatePrivatePlan{
		Creator:            creatorAddr.String(),
		PlanId:             plan.Id,
		FarmingPoolAddress: plan.FarmingPoolAddress,
	}); err != nil {
		return types.Plan{}, err
	}

	return plan, nil
}

// CreatePublicPlan creates a new public farming plan.
func (k Keeper) CreatePublicPlan(
	ctx sdk.Context, description string,
	farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time,
) (types.Plan, error) {
	return k.createPlan(
		ctx, description, farmingPoolAddr, termAddr,
		rewardAllocs, startTime, endTime, false)
}

func (k Keeper) createPlan(
	ctx sdk.Context, description string, farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time, isPrivate bool,
) (types.Plan, error) {
	// Check if end time > block time
	if !endTime.After(ctx.BlockTime()) {
		return types.Plan{}, sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest, "end time is past")
	}

	for _, rewardAlloc := range rewardAllocs {
		if rewardAlloc.PairId > 0 {
			_, found := k.liquidityKeeper.GetPair(ctx, rewardAlloc.PairId)
			if !found {
				return types.Plan{}, sdkerrors.Wrapf(
					sdkerrors.ErrNotFound, "pair %d not found", rewardAlloc.PairId)
			}
		}
	}

	// Generate the next plan id and update the last plan id.
	id, _ := k.GetLastPlanId(ctx)
	id++
	k.SetLastPlanId(ctx, id)

	plan := types.NewPlan(
		id, description, farmingPoolAddr, termAddr, rewardAllocs,
		startTime, endTime, isPrivate)
	k.SetPlan(ctx, plan)

	if plan.IsPrivate {
		k.SetNumPrivatePlans(ctx, k.GetNumPrivatePlans(ctx)+1)
	}

	return plan, nil
}

// TerminateEndedPlans iterates through all plans and terminate the plans
// which should be ended by the current block time.
func (k Keeper) TerminateEndedPlans(ctx sdk.Context) (err error) {
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsTerminated {
			return false
		}
		if !ctx.BlockTime().Before(plan.EndTime) {
			if err = k.TerminatePlan(ctx, plan); err != nil {
				return true
			}
		}
		return false
	})
	return err
}

// TerminatePlan mark the plan as terminated and send remaining balances
// in the farming pool to the termination address.
func (k Keeper) TerminatePlan(ctx sdk.Context, plan types.Plan) error {
	if plan.IsTerminated {
		return types.ErrPlanAlreadyTerminated
	}
	farmingPoolAddr := plan.GetFarmingPoolAddress()
	balances := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
	if !balances.IsZero() {
		if err := k.bankKeeper.SendCoins(
			ctx, farmingPoolAddr, plan.GetTerminationAddress(), balances); err != nil {
			return err
		}
	}
	plan.IsTerminated = true
	k.SetPlan(ctx, plan)
	if plan.IsPrivate {
		k.SetNumPrivatePlans(ctx, k.GetNumPrivatePlans(ctx)-1)
	}
	if err := ctx.EventManager().EmitTypedEvent(&types.EventTerminatePlan{
		PlanId: plan.Id,
	}); err != nil {
		return err
	}
	return nil
}

// AllocateRewards allocates the current block's rewards to the farms
// based on active plans.
func (k Keeper) AllocateRewards(ctx sdk.Context) error {
	lastBlockTime, found := k.GetLastBlockTime(ctx)
	if !found {
		// For the very first block, just skip it.
		return nil
	}

	blockDuration := ctx.BlockTime().Sub(lastBlockTime)
	// Constrain the block duration to the max block duration param.
	if maxBlockDuration := k.GetMaxBlockDuration(ctx); blockDuration > maxBlockDuration {
		blockDuration = maxBlockDuration
	}

	ck := newCachingKeeper(k)
	ra := newRewardAllocator(ctx, k, ck)
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsTerminated || !plan.IsActiveAt(ctx.BlockTime()) {
			return false // Skip
		}
		for _, rewardAlloc := range plan.RewardAllocations {
			rewards := types.RewardsForBlock(rewardAlloc.RewardsPerDay, blockDuration)
			// TODO: allocate sdk.DecCoins instead of sdk.Coins in future
			truncatedRewards, _ := rewards.TruncateDecimal()
			if truncatedRewards.IsAllPositive() {
				if rewardAlloc.Denom != "" {
					ra.allocateRewardsToDenom(plan.GetFarmingPoolAddress(), rewardAlloc.Denom, truncatedRewards)
				} else if rewardAlloc.PairId > 0 {
					pair, found := ck.getPair(ctx, rewardAlloc.PairId)
					if !found { // It should never happen
						panic("pair not found")
					}
					if pair.LastPrice == nil { // If the pair doesn't have the last price, skip.
						continue
					}
					ra.allocateRewardsToPair(plan.GetFarmingPoolAddress(), pair, truncatedRewards)
				}
			}
		}
		return false
	})

	rewardsByDenom := map[string]sdk.DecCoins{}
	// We keep this slice for deterministic iteration over the rewardsByDenom map.
	var denomsWithRewards []string
	for _, farmingPoolAddr := range ra.farmingPoolAddrs {
		farmingPool := farmingPoolAddr.String()
		totalRewards := ra.totalRewardsByFarmingPool[farmingPool]
		balances := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
		if !balances.IsAllGTE(totalRewards) {
			continue
		}
		if err := k.bankKeeper.SendCoins(
			ctx, farmingPoolAddr, types.RewardsPoolAddress, totalRewards); err != nil {
			return err
		}
		for denom, rewards := range ra.allocatedRewards[farmingPool] {
			if _, ok := rewardsByDenom[denom]; !ok {
				denomsWithRewards = append(denomsWithRewards, denom)
			}
			rewardsByDenom[denom] = rewardsByDenom[denom].Add(rewards...)
		}
	}

	sort.Strings(denomsWithRewards)
	for _, denom := range denomsWithRewards {
		farm, _ := ck.getFarm(ctx, denom)
		farm.CurrentRewards = farm.CurrentRewards.Add(rewardsByDenom[denom]...)
		farm.OutstandingRewards = farm.OutstandingRewards.Add(rewardsByDenom[denom]...)
		k.SetFarm(ctx, denom, farm)
	}

	return nil
}

// CanCreatePrivatePlan returns true if the current number of non-terminated
// private plans is less than the limit.
func (k Keeper) CanCreatePrivatePlan(ctx sdk.Context) bool {
	return k.GetNumPrivatePlans(ctx) < uint64(k.GetMaxNumPrivatePlans(ctx))
}
