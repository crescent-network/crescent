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
	// Check if end time > block time
	if !endTime.After(ctx.BlockTime()) {
		return types.Plan{}, sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest, "end time is past")
	}

	// Check if the number of non-terminated private plans is not greater than
	// the MaxNumPrivatePlans param.
	// TODO: store the counter separately to optimize gas usage?
	numPrivatePlans := 0
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsPrivate && !plan.IsTerminated {
			numPrivatePlans++
		}
		return false
	})
	if maxNum := k.GetMaxNumPrivatePlans(ctx); uint32(numPrivatePlans) >= maxNum {
		return types.Plan{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"maximum number of active private plans reached: %d", maxNum)
	}

	for _, rewardAlloc := range rewardAllocs {
		_, found := k.liquidityKeeper.GetPair(ctx, rewardAlloc.PairId)
		if !found {
			return types.Plan{}, sdkerrors.Wrapf(
				sdkerrors.ErrNotFound, "pair %d not found", rewardAlloc.PairId)
		}
	}

	fee := k.GetPrivatePlanCreationFee(ctx)
	feeCollectorAddr, err := sdk.AccAddressFromBech32(k.GetFeeCollector(ctx))
	if err != nil {
		return types.Plan{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, creatorAddr, feeCollectorAddr, fee); err != nil {
		return types.Plan{}, err
	}

	// Generate the next plan id and update the last plan id.
	id, _ := k.GetLastPlanId(ctx)
	id++
	k.SetLastPlanId(ctx, id)

	farmingPoolAddr := types.DeriveFarmingPoolAddress(id)
	plan := types.NewPlan(
		id, description, farmingPoolAddr, creatorAddr, rewardAllocs,
		startTime, endTime, true)
	k.SetPlan(ctx, plan)

	return plan, nil
}

// TerminateEndedPlans iterates through all plans and terminate the plans
// which should be ended by the current block time.
func (k Keeper) TerminateEndedPlans(ctx sdk.Context) (err error) {
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsTerminated {
			return false
		}
		if !ctx.BlockTime().After(plan.EndTime) {
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
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "already terminated plan")
	}
	farmingPoolAddr := plan.GetFarmingPoolAddress()
	balances := k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
	if err := k.bankKeeper.SendCoins(
		ctx, farmingPoolAddr, plan.GetTerminationAddress(), balances); err != nil {
		return err
	}
	plan.IsTerminated = true
	k.SetPlan(ctx, plan)
	return nil
}

func (k Keeper) AllocateRewards(ctx sdk.Context) error {
	lastBlockTime, found := k.GetLastBlockTime(ctx)
	if !found {
		// Skip this block.
		k.SetLastBlockTime(ctx, ctx.BlockTime())
		return nil
	}

	blockDuration := ctx.BlockTime().Sub(lastBlockTime)
	if maxBlockDuration := k.GetMaxBlockDuration(ctx); blockDuration > maxBlockDuration {
		// Constrain the block duration to the max block duration param.
		blockDuration = maxBlockDuration
	}

	var activePlans []types.Plan
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		if plan.IsTerminated || !plan.IsActiveAt(ctx.BlockTime()) {
			return false // Skip
		}
		activePlans = append(activePlans, plan)
		return false
	})

	// farmingPoolAddr => (pairId => rewards)
	allocsByFarmingPool := map[string]map[uint64]sdk.Coins{}
	pairCache := map[uint64]liquiditytypes.Pair{}
	for _, plan := range activePlans {
		allocs, ok := allocsByFarmingPool[plan.FarmingPoolAddress]
		if !ok {
			allocs = map[uint64]sdk.Coins{}
			allocsByFarmingPool[plan.FarmingPoolAddress] = allocs
		}

		for _, rewardAlloc := range plan.RewardAllocations {
			pair, ok := pairCache[rewardAlloc.PairId]
			if !ok {
				pair, found = k.liquidityKeeper.GetPair(ctx, rewardAlloc.PairId)
				if !found { // It never happens
					panic("pair not found")
				}
				pairCache[rewardAlloc.PairId] = pair
			}
			if pair.LastPrice == nil { // If the pair doesn't have the last price, skip.
				continue
			}

			rewards := types.RewardsForBlock(rewardAlloc.RewardsPerDay, blockDuration)
			// TODO: allocate sdk.DecCoins instead of sdk.Coins?
			truncatedRewards, _ := rewards.TruncateDecimal()
			allocs[rewardAlloc.PairId] = allocs[rewardAlloc.PairId].Add(truncatedRewards...)
		}
	}

	farmingPoolBalancesCache := map[string]sdk.Coins{}
	rewardsDiffByDenom := map[string]sdk.DecCoins{}
	farmCache := map[string]*types.Farm{}
	// We keep this slice for deterministic iteration over the rewardsDiffByDenom map.
	var denomsWithRewardsDiff []string
	for _, plan := range activePlans {
		allocs := allocsByFarmingPool[plan.FarmingPoolAddress]
		totalRewards := sdk.Coins{}
		for _, rewards := range allocs {
			totalRewards = totalRewards.Add(rewards...)
		}

		farmingPoolAddr, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
		if err != nil { // It never happens
			return err
		}
		balances, ok := farmingPoolBalancesCache[plan.FarmingPoolAddress]
		if !ok {
			balances = k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
			farmingPoolBalancesCache[plan.FarmingPoolAddress] = balances
		}
		if !balances.IsAllGTE(totalRewards) {
			continue
		}

		if err := k.bankKeeper.SendCoins(
			ctx, farmingPoolAddr, types.RewardsPoolAddress, totalRewards); err != nil {
			return err
		}

		for _, rewardAlloc := range plan.RewardAllocations {
			pair := pairCache[rewardAlloc.PairId]
			if pair.LastPrice == nil {
				continue
			}

			var poolCache []liquiditytypes.Pool
			rewardWeightByPool := map[uint64]sdk.Dec{}
			totalRewardWeight := sdk.ZeroDec()
			_ = k.liquidityKeeper.IteratePoolsByPair(
				ctx, rewardAlloc.PairId, func(pool liquiditytypes.Pool) (stop bool, err error) {
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

					// Load the farm object either from the cache or the store.
					farm, ok := farmCache[pool.PoolCoinDenom]
					if !ok {
						f, found := k.GetFarm(ctx, pool.PoolCoinDenom)
						if found {
							farm = &f
						}
						// If farm wasn't found, then nil will be set for the key.
						farmCache[pool.PoolCoinDenom] = farm
					}

					// If there's no farm yet(which means there's no farmer yet),
					// just skip this pool.
					if farm == nil {
						return false, nil
					}
					// TODO: what if the farm exists but has zero total farming amount?

					rewardWeight := k.PoolRewardWeight(ctx, pool, pair)
					rewardWeightByPool[pool.Id] = rewardWeight
					totalRewardWeight = totalRewardWeight.Add(rewardWeight)
					poolCache = append(poolCache, pool)
					return false, nil
				})

			for _, pool := range poolCache {
				rewardProportion := rewardWeightByPool[pool.Id].QuoTruncate(totalRewardWeight)
				rewards := sdk.NewDecCoinsFromCoins(allocs[rewardAlloc.PairId]...).
					MulDecTruncate(rewardProportion)

				if _, ok := rewardsDiffByDenom[pool.PoolCoinDenom]; !ok {
					denomsWithRewardsDiff = append(denomsWithRewardsDiff, pool.PoolCoinDenom)
				}
				rewardsDiffByDenom[pool.PoolCoinDenom] =
					rewardsDiffByDenom[pool.PoolCoinDenom].Add(rewards...)
			}
		}
	}

	for _, denom := range denomsWithRewardsDiff {
		farm := farmCache[denom]
		farm.CurrentRewards = farm.CurrentRewards.Add(rewardsDiffByDenom[denom]...)
		farm.OutstandingRewards = farm.OutstandingRewards.Add(rewardsDiffByDenom[denom]...)
		k.SetFarm(ctx, denom, *farm)
	}

	return nil
}
