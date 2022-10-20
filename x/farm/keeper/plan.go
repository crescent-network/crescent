package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
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
		_, found := k.liquidityKeeper.GetPair(ctx, rewardAlloc.PairId)
		if !found {
			return types.Plan{}, sdkerrors.Wrapf(
				sdkerrors.ErrNotFound, "pair %d not found", rewardAlloc.PairId)
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
	if plan.FarmingPoolAddress != plan.TerminationAddress {
		balances := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
		if !balances.IsZero() {
			if err := k.bankKeeper.SendCoins(
				ctx, farmingPoolAddr, plan.GetTerminationAddress(), balances); err != nil {
				return err
			}
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

	// An active plan means that the plan is not terminated and the current
	// block's time is between the time range of the plan.
	var activePlans []types.Plan
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsTerminated || !plan.IsActiveAt(ctx.BlockTime()) {
			return false // Skip
		}
		activePlans = append(activePlans, plan)
		return false
	})

	// CacheKeeper acts like a proxy to underlying keeper methods,
	// but caches the result to avoid unnecessary gas consumptions and
	// store read operations.
	cachedKeeper := NewCachedKeeper(k)
	// allocsByFarmingPool keeps track of the allocation information
	// grouped by farming pool address.
	// Each entry of the map is another map which holds mappings from
	// pair id to rewards for this block.
	allocsByFarmingPool := map[string]map[uint64]sdk.Coins{}
	var farmingPoolAddrs []sdk.AccAddress // For deterministic iteration
	// eligiblePoolsByPairs maps pair id to pools belong to the pair,
	// which are eligible for rewards allocation.
	// An eligible pool means that:
	// - The pool is not disabled
	// - The pool's price range includes the pair's last price
	// - Total farming amount of the pool's pool coin is positive
	eligiblePoolsByPair := map[uint64][]liquiditytypes.Pool{}
	for _, plan := range activePlans {
		for _, rewardAlloc := range plan.RewardAllocations {
			pair, found := cachedKeeper.GetPair(ctx, rewardAlloc.PairId)
			if !found { // It should never happen
				panic("pair not found")
			}
			if pair.LastPrice == nil { // If the pair doesn't have the last price, skip.
				continue
			}

			eligiblePools, ok := eligiblePoolsByPair[pair.Id]
			if !ok {
				// Collect pools eligible for reward allocation.
				_ = k.liquidityKeeper.IteratePoolsByPair(
					ctx, pair.Id, func(pool liquiditytypes.Pool) (stop bool, err error) {
						if pool.Disabled {
							return false, nil
						}
						// If the pool is a ranged pool and its pair's last price is out of
						// its price range, skip the pool.
						// This is because the amplification factor would be zero
						// so its reward weight would eventually be zero, too.
						if pool.Type == liquiditytypes.PoolTypeRanged &&
							(pair.LastPrice.LT(*pool.MinPrice) || pair.LastPrice.GT(*pool.MaxPrice)) {
							return false, nil
						}

						farm, found := cachedKeeper.GetFarm(ctx, pool.PoolCoinDenom)
						if !found || !farm.TotalFarmingAmount.IsPositive() {
							return false, nil
						}

						eligiblePools = append(eligiblePools, pool)
						return false, nil
					},
				)
				eligiblePoolsByPair[pair.Id] = eligiblePools
			}

			// Allocate rewards only when there is at least one eligible pool
			// belonging to the pair.
			if len(eligiblePools) > 0 {
				rewards := types.RewardsForBlock(rewardAlloc.RewardsPerDay, blockDuration)
				// TODO: allocate sdk.DecCoins instead of sdk.Coins in future
				truncatedRewards, _ := rewards.TruncateDecimal()
				if truncatedRewards.IsAllPositive() {
					allocs, ok := allocsByFarmingPool[plan.FarmingPoolAddress]
					if !ok {
						allocs = map[uint64]sdk.Coins{}
						allocsByFarmingPool[plan.FarmingPoolAddress] = allocs
						farmingPoolAddrs = append(farmingPoolAddrs, plan.GetFarmingPoolAddress())
					}
					allocs[pair.Id] = allocs[pair.Id].Add(truncatedRewards...)
				}
			}
		}
	}

	rewardsByPairId := map[uint64]sdk.Coins{}
	// We keep this slice for deterministic iteration over the rewardsByPairId map.
	var pairIdsWithRewards []uint64
	for _, farmingPoolAddr := range farmingPoolAddrs {
		allocs, ok := allocsByFarmingPool[farmingPoolAddr.String()]
		if !ok {
			continue
		}

		totalRewards := sdk.Coins{}
		for _, rewards := range allocs {
			totalRewards = totalRewards.Add(rewards...)
		}
		balances := cachedKeeper.SpendableCoins(ctx, farmingPoolAddr)
		if !balances.IsAllGTE(totalRewards) {
			continue
		}
		if err := k.bankKeeper.SendCoins(
			ctx, farmingPoolAddr, types.RewardsPoolAddress, totalRewards); err != nil {
			return err
		}

		for pairId, rewards := range allocs {
			if _, ok := rewardsByPairId[pairId]; !ok {
				pairIdsWithRewards = append(pairIdsWithRewards, pairId)
			}
			rewardsByPairId[pairId] = rewardsByPairId[pairId].Add(rewards...)
		}
	}

	rewardsByDenom := map[string]sdk.DecCoins{}
	// We keep this slice for deterministic iteration over the rewardsByDenom map.
	var denomsWithRewards []string
	for _, pairId := range pairIdsWithRewards {
		pair, _ := cachedKeeper.GetPair(ctx, pairId)
		rewards := rewardsByPairId[pairId]

		rewardWeightByPool := map[uint64]sdk.Dec{}
		totalRewardWeight := sdk.ZeroDec()
		for _, pool := range eligiblePoolsByPair[pairId] {
			rewardWeight := k.PoolRewardWeight(ctx, pool, pair)
			rewardWeightByPool[pool.Id] = rewardWeight
			totalRewardWeight = totalRewardWeight.Add(rewardWeight)
		}

		for _, pool := range eligiblePoolsByPair[pairId] {
			rewardProportion := rewardWeightByPool[pool.Id].QuoTruncate(totalRewardWeight)
			rewards := sdk.NewDecCoinsFromCoins(rewards...).
				MulDecTruncate(rewardProportion)

			if _, ok := rewardsByDenom[pool.PoolCoinDenom]; !ok {
				denomsWithRewards = append(denomsWithRewards, pool.PoolCoinDenom)
			}
			rewardsByDenom[pool.PoolCoinDenom] =
				rewardsByDenom[pool.PoolCoinDenom].Add(rewards...)
		}
	}

	for _, denom := range denomsWithRewards {
		farm, _ := cachedKeeper.GetFarm(ctx, denom)
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
