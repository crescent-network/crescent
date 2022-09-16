package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// CreatePrivatePlan creates a new private farming plan.
func (k Keeper) CreatePrivatePlan(
	ctx sdk.Context, creatorAddr sdk.AccAddress, description string,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time,
) (types.Plan, error) {
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

	allocsByFarmingPool := map[string]map[uint64]sdk.Coins{} // farmingPoolAddr => (pairId => rewards)
	for _, plan := range activePlans {
		for _, rewardAlloc := range plan.RewardAllocations {
			allocs, ok := allocsByFarmingPool[plan.FarmingPoolAddress]
			if !ok {
				allocs = map[uint64]sdk.Coins{}
				allocsByFarmingPool[plan.FarmingPoolAddress] = allocs
			}
			rewards := types.RewardsForBlock(rewardAlloc.RewardsPerDay, blockDuration)
			truncatedRewards, _ := rewards.TruncateDecimal()
			allocs[rewardAlloc.PairId] = allocs[rewardAlloc.PairId].Add(truncatedRewards...)
		}
	}

	farmingPoolBalances := map[string]sdk.Coins{}
	rewardsDeltaByDenom := map[string]sdk.DecCoins{}
	farmCache := map[string]*types.Farm{}
	// We keep this slice for deterministic iteration over the rewardsDeltaByDenom map.
	var denomsWithRewardsChanged []string
	for _, plan := range activePlans {
		allocs := allocsByFarmingPool[plan.FarmingPoolAddress]
		totalRewards := sdk.Coins{}
		for _, rewards := range allocs {
			totalRewards = totalRewards.Add(rewards...)
		}

		farmingPoolAddr, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
		if err != nil {
			panic(err)
		}
		balances, ok := farmingPoolBalances[plan.FarmingPoolAddress]
		if !ok {
			balances = k.bankKeeper.SpendableCoins(ctx, farmingPoolAddr)
		}
		if !balances.IsAllGTE(totalRewards) {
			continue
		}

		if err := k.bankKeeper.SendCoins(ctx, farmingPoolAddr, types.RewardsPoolAddress, totalRewards); err != nil {
			return err
		}

		for _, rewardAlloc := range plan.RewardAllocations {
			_ = k.liquidityKeeper.IteratePoolsByPair(ctx, rewardAlloc.PairId, func(pool liquiditytypes.Pool) (stop bool, err error) {
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
				if _, found := k.GetFarm(ctx, pool.PoolCoinDenom); !found {
					// If there's no Farm yet(which means there's no farmer for the denom),
					// just skip this pool.
					return false, nil
				}
				// TODO: what if the farm exists but has zero total farming amount?
				rewards := sdk.NewDecCoinsFromCoins(allocs[rewardAlloc.PairId]...) // TODO: use weighted rewards
				if _, ok := rewardsDeltaByDenom[pool.PoolCoinDenom]; !ok {
					denomsWithRewardsChanged = append(denomsWithRewardsChanged, pool.PoolCoinDenom)
				}
				rewardsDeltaByDenom[pool.PoolCoinDenom] =
					rewardsDeltaByDenom[pool.PoolCoinDenom].Add(rewards...)
				return false, nil
			})
		}
	}

	for _, denom := range denomsWithRewardsChanged {
		farm := farmCache[denom]
		farm.CurrentRewards = farm.CurrentRewards.Add(rewardsDeltaByDenom[denom]...)
		farm.OutstandingRewards = farm.OutstandingRewards.Add(rewardsDeltaByDenom[denom]...)
		k.SetFarm(ctx, denom, *farm)
	}

	return nil
}
