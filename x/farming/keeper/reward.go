package keeper

import (
	"fmt"

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
	k.cdc.MustUnmarshal(bz, &reward)
	return reward, true
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

	if err := k.ReleaseStakingCoins(ctx, farmerAcc, amount); err != nil {
		return err
	}

	for _, denom := range stakingCoinDenoms {
		k.DeleteReward(ctx, denom, farmerAcc)
	}

	if len(k.GetRewardsByFarmer(ctx, farmerAcc)) == 0 {
		staking, found := k.GetStakingByFarmer(ctx, farmerAcc)
		if !found { // TODO: remove this check
			return fmt.Errorf("staking not found")
		}
		if staking.StakedCoins.IsZero() && staking.QueuedCoins.IsZero() {
			k.DeleteStaking(ctx, staking)
		}
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
