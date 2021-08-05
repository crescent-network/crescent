package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

// GetReward returns a specific reward.
func (k Keeper) GetReward(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (reward types.Reward, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRewardKey(stakingCoinDenom, farmerAcc))
	if bz == nil {
		return reward, false
	}
	var rewardCoins types.RewardCoins
	k.cdc.MustUnmarshal(bz, &rewardCoins)
	return types.Reward{
		Farmer:           farmerAcc.String(),
		StakingCoinDenom: stakingCoinDenom,
		RewardCoins:      rewardCoins.RewardCoins,
	}, true
}

// GetRewardsByFarmer reads from kvstore and return a specific Reward indexed by given farmer's address
func (k Keeper) GetRewardsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) (rewards []types.Reward) {
	k.IterateRewardsByFarmer(ctx, farmerAcc, func(reward types.Reward) bool {
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}

// GetAllRewards returns all rewards in the Keeper.
func (k Keeper) GetAllRewards(ctx sdk.Context) (rewards []types.Reward) {
	k.IterateAllRewards(ctx, func(reward types.Reward) (stop bool) {
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}

// GetRewardsByStakingCoinDenom reads from kvstore and return a specific Reward indexed by given staking coin denom
func (k Keeper) GetRewardsByStakingCoinDenom(ctx sdk.Context, denom string) (rewards []types.Reward) {
	k.IterateRewardsByStakingCoinDenom(ctx, denom, func(reward types.Reward) bool {
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}

// SetReward implements Reward.
func (k Keeper) SetReward(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress, rewardCoins sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.RewardCoins{RewardCoins: rewardCoins})
	store.Set(types.GetRewardKey(stakingCoinDenom, farmerAcc), bz)
	store.Set(types.GetRewardByFarmerAndStakingCoinDenomIndexKey(farmerAcc, stakingCoinDenom), []byte{})
}

// DeleteReward deletes a reward for the reward mapper store.
func (k Keeper) DeleteReward(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRewardKey(stakingCoinDenom, farmerAcc))
	store.Delete(types.GetRewardByFarmerAndStakingCoinDenomIndexKey(farmerAcc, stakingCoinDenom))
}

// IterateAllRewards iterates over all the stored rewards and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllRewards(ctx sdk.Context, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		stakingCoinDenom, farmerAcc := types.ParseRewardKey(iterator.Key())
		var rewardCoins types.RewardCoins
		k.cdc.MustUnmarshal(iterator.Value(), &rewardCoins)
		if cb(types.Reward{Farmer: farmerAcc.String(), StakingCoinDenom: stakingCoinDenom, RewardCoins: rewardCoins.RewardCoins}) {
			break
		}
	}
}

// IterateRewardsByStakingCoinDenom iterates over all the stored rewards indexed by given staking coin denom and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateRewardsByStakingCoinDenom(ctx sdk.Context, stakingCoinDenom string, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardsByStakingCoinDenomKey(stakingCoinDenom))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, farmerAcc := types.ParseRewardKey(iterator.Key())
		var rewardCoins types.RewardCoins
		k.cdc.MustUnmarshal(iterator.Value(), &rewardCoins)
		if cb(types.Reward{Farmer: farmerAcc.String(), StakingCoinDenom: stakingCoinDenom, RewardCoins: rewardCoins.RewardCoins}) {
			break
		}
	}
}

// IterateRewardsByFarmer iterates over all the stored rewards indexed by given farmer's address and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateRewardsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardsByFarmerIndexKey(farmerAcc))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, stakingCoinDenom := types.ParseRewardsByFarmerIndexKey(iterator.Key())
		reward, _ := k.GetReward(ctx, stakingCoinDenom, farmerAcc)
		if cb(reward) {
			break
		}
	}
}

// UnmarshalRewardCoins unmarshals a RewardCoins from bytes.
func (k Keeper) UnmarshalRewardCoins(bz []byte) (rewardCoins types.RewardCoins, err error) {
	return rewardCoins, k.cdc.Unmarshal(bz, &rewardCoins)
}

// Harvest claims farming rewards from the reward pool account.
func (k Keeper) Harvest(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenoms []string) error {
	amount := sdk.NewCoins()
	for _, denom := range stakingCoinDenoms {
		reward, found := k.GetReward(ctx, denom, farmerAcc)
		if !found {
			return sdkerrors.Wrapf(types.ErrRewardNotExists, "no reward for staking coin denom %s", denom)
		}
		amount = amount.Add(reward.RewardCoins...)
	}

	// TODO: remove staking
	// TODO: send reward from the reward pool

	for _, denom := range stakingCoinDenoms {
		k.DeleteReward(ctx, denom, farmerAcc)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeHarvest,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyRewardCoins, amount.String()),
		),
	})

	return nil
}

type DistributionInfo struct {
	Plan   types.PlanI
	Amount sdk.Coins
}

func (k Keeper) DistributionInfos(ctx sdk.Context) []DistributionInfo {
	farmingPoolBalances := make(map[string]sdk.Coins)   // farmingPoolAddress => sdk.Coins
	distrCoins := make(map[string]map[uint64]sdk.Coins) // farmingPoolAddress => (planId => sdk.Coins)

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

		dc, ok := distrCoins[farmingPool]
		if !ok {
			dc = make(map[uint64]sdk.Coins)
			distrCoins[farmingPool] = dc
		}

		switch plan := plan.(type) {
		case *types.FixedAmountPlan:
			dc[plan.GetId()] = plan.EpochAmount
		case *types.RatioPlan:
			dc[plan.GetId()], _ = sdk.NewDecCoinsFromCoins(balances...).MulDecTruncate(plan.EpochRatio).TruncateDecimal()
		}
	}

	var distrInfos []DistributionInfo
	for farmingPool, coins := range distrCoins {
		totalCoins := sdk.NewCoins()
		for _, amt := range coins {
			totalCoins = totalCoins.Add(amt...)
		}

		balances := farmingPoolBalances[farmingPool]
		if !totalCoins.IsAllLT(balances) {
			continue
		}

		for planID, amt := range coins {
			distrInfos = append(distrInfos, DistributionInfo{
				Plan:   plans[planID],
				Amount: amt,
			})
		}
	}

	return distrInfos
}

func (k Keeper) DistributeRewards(ctx sdk.Context) error {
	stakingsByDenom := make(map[string][]types.Staking)
	totalStakedAmtByDenom := make(map[string]sdk.Int)

	for _, distrInfo := range k.DistributionInfos(ctx) {
		stakingCoinWeights := distrInfo.Plan.GetStakingCoinWeights()
		totalDistrAmt := sdk.NewCoins()

		totalWeight := sdk.ZeroDec()
		for _, coinWeight := range stakingCoinWeights {
			totalWeight = totalWeight.Add(coinWeight.Amount)
		}

		for _, coinWeight := range stakingCoinWeights {
			stakings, ok := stakingsByDenom[coinWeight.Denom]
			if !ok {
				stakings = k.GetStakingsByStakingCoinDenom(ctx, coinWeight.Denom)
				stakingsByDenom[coinWeight.Denom] = stakings

				for _, staking := range stakings {
					totalStakedAmt, ok := totalStakedAmtByDenom[coinWeight.Denom]
					if !ok {
						totalStakedAmt = sdk.ZeroInt()
					}
					totalStakedAmtByDenom[coinWeight.Denom] = totalStakedAmt.Add(staking.StakedCoins.AmountOf(coinWeight.Denom))
				}
			}

			totalStakedAmt := totalStakedAmtByDenom[coinWeight.Denom]

			for _, staking := range stakings {
				stakedAmt := staking.StakedCoins.AmountOf(coinWeight.Denom)
				if !stakedAmt.IsPositive() {
					continue
				}

				stakedProportion := stakedAmt.ToDec().QuoTruncate(totalStakedAmt.ToDec())
				weightProportion := coinWeight.Amount.QuoTruncate(totalWeight)
				distrAmt, _ := sdk.NewDecCoinsFromCoins(distrInfo.Amount...).MulDecTruncate(stakedProportion.MulTruncate(weightProportion)).TruncateDecimal()

				reward, _ := k.GetReward(ctx, coinWeight.Denom, staking.GetFarmer())
				reward.RewardCoins = reward.RewardCoins.Add(distrAmt...)
				k.SetReward(ctx, coinWeight.Denom, staking.GetFarmer(), reward.RewardCoins)
				totalDistrAmt = totalDistrAmt.Add(distrAmt...)
			}
		}

		if !totalDistrAmt.IsZero() {
			k.SetLastDistributedTime(ctx, distrInfo.Plan.GetId(), ctx.BlockTime())
			totalDistributedRewardCoins := k.GetTotalDistributedRewardCoins(ctx, distrInfo.Plan.GetId())
			totalDistributedRewardCoins = totalDistributedRewardCoins.Add(totalDistrAmt...)
			k.SetTotalDistributedRewardCoins(ctx, distrInfo.Plan.GetId(), totalDistributedRewardCoins)

			if err := k.bankKeeper.SendCoins(ctx, distrInfo.Plan.GetFarmingPoolAddress(), distrInfo.Plan.GetRewardPoolAddress(), totalDistrAmt); err != nil {
				return err
			}
		}
	}

	// TODO: emit an endblock event

	return nil
}
