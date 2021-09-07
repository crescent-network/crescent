package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (k Keeper) GetHistoricalRewards(ctx sdk.Context, stakingCoinDenom string, epoch uint64) (rewards types.HistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoricalRewardsKey(stakingCoinDenom, epoch))
	k.cdc.MustUnmarshal(bz, &rewards)
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

func (k Keeper) GetCurrentRewards(ctx sdk.Context, stakingCoinDenom string) (rewards types.CurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCurrentRewardsKey(stakingCoinDenom))
	k.cdc.MustUnmarshal(bz, &rewards)
	return
}

func (k Keeper) SetCurrentRewards(ctx sdk.Context, stakingCoinDenom string, rewards types.CurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetCurrentRewardsKey(stakingCoinDenom), bz)
}

func (k Keeper) CalculateRewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string, endingEpoch uint64) (rewards sdk.Coins) {
	staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
	if !found {
		staking.Amount = sdk.ZeroInt()
	}

	starting := k.GetHistoricalRewards(ctx, stakingCoinDenom, staking.StartingEpoch-1)
	ending := k.GetHistoricalRewards(ctx, stakingCoinDenom, endingEpoch)
	diff := ending.CumulativeUnitRewards.Sub(starting.CumulativeUnitRewards)
	rewards, _ = diff.MulDecTruncate(staking.Amount.ToDec()).TruncateDecimal()
	return
}

func (k Keeper) WithdrawRewards(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) (sdk.Coins, error) {
	staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
	if !found {
		return nil, fmt.Errorf("empty starting info") // TODO: use correct error
	}

	current := k.GetCurrentRewards(ctx, stakingCoinDenom)
	rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, current.Epoch-1)

	if !rewards.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, k.GetRewardsReservePoolAcc(ctx), farmerAcc, rewards); err != nil {
			return nil, err
		}
	}

	staking.StartingEpoch = current.Epoch
	k.SetStaking(ctx, stakingCoinDenom, farmerAcc, staking)

	return rewards, nil
}

func (k Keeper) WithdrawAllRewards(ctx sdk.Context, farmerAcc sdk.AccAddress) (sdk.Coins, error) {
	totalRewards := sdk.NewCoins()

	k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		current := k.GetCurrentRewards(ctx, stakingCoinDenom)
		rewards := k.CalculateRewards(ctx, farmerAcc, stakingCoinDenom, current.Epoch-1)

		if !rewards.IsZero() {
			totalRewards = totalRewards.Add(rewards...)
		}

		staking.StartingEpoch = current.Epoch
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
			sdk.NewAttribute(types.AttributeKeyRewardCoins, totalRewards.String()),
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
	for _, plan := range k.GetAllPlans(ctx) {
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
		if !totalCoins.IsAllLT(balances) {
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
	for _, allocInfo := range k.AllocationInfos(ctx) {
		totalWeight := sdk.ZeroDec()
		for _, weight := range allocInfo.Plan.GetStakingCoinWeights() {
			totalWeight = totalWeight.Add(weight.Amount)
		}

		totalAllocCoins := sdk.NewDecCoins()
		for _, weight := range allocInfo.Plan.GetStakingCoinWeights() {
			totalStaking, found := k.GetTotalStaking(ctx, weight.Denom)
			if !found {
				continue
			}
			if !totalStaking.Amount.IsPositive() {
				continue
			}

			weightProportion := weight.Amount.QuoTruncate(totalWeight)
			allocCoins := sdk.NewDecCoinsFromCoins(allocInfo.Amount...).MulDecTruncate(weightProportion)

			current := k.GetCurrentRewards(ctx, weight.Denom)
			historical := k.GetHistoricalRewards(ctx, weight.Denom, current.Epoch-1)
			k.SetHistoricalRewards(ctx, weight.Denom, current.Epoch, types.HistoricalRewards{
				CumulativeUnitRewards: historical.CumulativeUnitRewards.Add(allocCoins.QuoDecTruncate(totalStaking.Amount.ToDec())...),
			})
			k.SetCurrentRewards(ctx, weight.Denom, types.CurrentRewards{
				Epoch: current.Epoch + 1,
			})

			totalAllocCoins = totalAllocCoins.Add(allocCoins...)
		}

		if totalAllocCoins.IsZero() {
			continue
		}

		truncatedAllocCoins, _ := totalAllocCoins.TruncateDecimal()

		rewardsReserveAcc := k.GetRewardsReservePoolAcc(ctx)
		if err := k.bankKeeper.SendCoins(ctx, allocInfo.Plan.GetFarmingPoolAddress(), rewardsReserveAcc, truncatedAllocCoins); err != nil {
			return err
		}

		t := ctx.BlockTime()
		_ = allocInfo.Plan.SetLastDistributionTime(&t)
		_ = allocInfo.Plan.SetDistributedCoins(allocInfo.Plan.GetDistributedCoins().Add(truncatedAllocCoins...))
		k.SetPlan(ctx, allocInfo.Plan)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRewardsAllocated,
				sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(allocInfo.Plan.GetId(), 10)),
				sdk.NewAttribute(types.AttributeKeyAmount, truncatedAllocCoins.String()),
			),
		})
	}

	return nil
}
