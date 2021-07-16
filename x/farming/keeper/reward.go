package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

// GetReward return a specific reward
func (k Keeper) GetReward(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (reward types.Reward, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRewardKey(stakingCoinDenom, farmerAcc))
	if bz == nil {
		return reward, false
	}
	k.cdc.MustUnmarshal(bz, &reward)
	return reward, true
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

// GetRewardsByFarmer reads from kvstore and return a specific Reward indexed by given farmer's address
func (k Keeper) GetRewardsByFarmer(ctx sdk.Context, farmer sdk.AccAddress) (rewards []types.Reward) {
	k.IterateRewardsByFarmer(ctx, farmer, func(reward types.Reward) bool {
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}

// SetReward implements Reward.
func (k Keeper) SetReward(ctx sdk.Context, reward types.Reward) {
	store := ctx.KVStore(k.storeKey)
	// TODO: only rewardCoins
	// 0x31 | StakingCoinDenomAddrLen (1 byte) | StakingCoinDenom | FarmerAddrLen (1 byte) | FarmerAddr -> ProtocolBuffer(sdk.Coins) RewardCoins
	bz := k.cdc.MustMarshal(&reward)
	store.Set(types.GetRewardKey(reward.StakingCoinDenom, reward.GetFarmerAddress()), bz)
	store.Set(types.GetRewardByFarmerAddrIndexKey(reward.GetFarmerAddress(), reward.StakingCoinDenom), nil)
}

// RemoveReward removes an reward for the reward mapper store.
func (k Keeper) RemoveReward(ctx sdk.Context, reward types.Reward) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRewardKey(reward.StakingCoinDenom, reward.GetFarmerAddress()))
	store.Delete(types.GetRewardByFarmerAddrIndexKey(reward.GetFarmerAddress(), reward.StakingCoinDenom))
}

// IterateAllRewards iterates over all the stored rewards and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllRewards(ctx sdk.Context, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// TODO: unmarshal values by key
		// Reward: 0x31 | StakingCoinDenomAddrLen (1 byte) | StakingCoinDenom | FarmerAddrLen (1 byte) | FarmerAddr -> ProtocolBuffer(sdk.Coins) RewardCoins
		var reward types.Reward
		k.cdc.MustUnmarshal(iterator.Value(), &reward)
		if cb(reward) {
			break
		}
	}
}

// IterateRewardsByStakingCoinDenom iterates over all the stored rewards indexed by given staking coin denom and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateRewardsByStakingCoinDenom(ctx sdk.Context, denom string, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardByStakingCoinDenomPrefix(denom))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// TODO: unmarshal values by key
		// Reward: 0x31 | StakingCoinDenomAddrLen (1 byte) | StakingCoinDenom | FarmerAddrLen (1 byte) | FarmerAddr -> ProtocolBuffer(sdk.Coins) RewardCoins
		var reward types.Reward
		k.cdc.MustUnmarshal(iterator.Value(), &reward)
		if cb(reward) {
			break
		}
	}
}

// IterateRewardsByFarmer iterates over all the stored rewards indexed by given farmer's address and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateRewardsByFarmer(ctx sdk.Context, farmer sdk.AccAddress, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardByFarmerAddrIndexPrefix(farmer))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// TODO: unmarshal values by key
		// RewardByFarmerAddrIndex: 0x32 | FarmerAddrLen (1 byte) | FarmerAddr | StakingCoinDenomAddrLen (1 byte) | StakingCoinDenom -> nil
		var reward types.Reward
		k.cdc.MustUnmarshal(iterator.Value(), &reward)
		if cb(reward) {
			break
		}
	}
}

// UnmarshalReward unmarshals a Reward from bytes.
func (k Keeper) UnmarshalReward(bz []byte) (types.Reward, error) {
	var reward types.Reward
	return reward, k.cdc.Unmarshal(bz, &reward)
}

// Harvest claims farming rewards from the reward pool account.
func (k Keeper) Harvest(ctx sdk.Context, msg *types.MsgHarvest) (types.Reward, error) {
	farmerAcc, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		return types.Reward{}, err
	}

	reward, found := k.GetReward(ctx, msg.StakingCoinDenom, farmerAcc)
	if !found {
		return types.Reward{}, types.ErrRewardNotExists
	}
	// TODO: add detailed core logic and validation
	if err := k.ReleaseStakingCoins(ctx, reward.GetFarmerAddress(), reward.RewardCoins); err != nil {
		panic(err)
	}

	return types.Reward{}, nil
}
