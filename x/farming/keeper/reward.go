package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/farming/x/farming/types"
)

// GetReward return a specific reward
func (k Keeper) GetReward(ctx sdk.Context, planID uint64, farmerAcc sdk.AccAddress) (reward types.Reward, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRewardIndexKey(planID, farmerAcc))
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

// GetRewardsByPlanID reads from kvstore and return a specific Reward indexed by given plan id
func (k Keeper) GetRewardsByPlanID(ctx sdk.Context, planID uint64) (rewards []types.Reward) {
	k.IterateRewardsByPlanID(ctx, planID, func(reward types.Reward) bool {
		rewards = append(rewards, reward)
		return false
	})

	return rewards
}

// SetReward implements Reward.
func (k Keeper) SetReward(ctx sdk.Context, reward types.Reward) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&reward)
	store.Set(types.GetRewardIndexKey(reward.PlanId, reward.GetFarmerAddress()), bz)
}

// RemoveReward removes an reward for the reward mapper store.
func (k Keeper) RemoveReward(ctx sdk.Context, reward types.Reward) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRewardIndexKey(reward.PlanId, reward.GetFarmerAddress()))
}

// IterateAllRewards iterates over all the stored rewards and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllRewards(ctx sdk.Context, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reward types.Reward
		k.cdc.MustUnmarshal(iterator.Value(), &reward)
		if cb(reward) {
			break
		}
	}
}

// IterateRewardsByPlanID iterates over all the stored rewards and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateRewardsByPlanID(ctx sdk.Context, planID uint64, cb func(reward types.Reward) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardPrefix(planID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
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

// Claim claims farming rewards from the reward pool account.
func (k Keeper) Claim(ctx sdk.Context, msg *types.MsgClaim) (types.Reward, error) {
	plan, found := k.GetPlan(ctx, msg.PlanId)
	if !found {
		return types.Reward{}, types.ErrPlanNotExists
	}

	farmerAcc, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		return types.Reward{}, err
	}

	reward, found := k.GetReward(ctx, plan.GetId(), farmerAcc)
	if !found {
		return types.Reward{}, types.ErrRewardNotExists
	}

	if err := k.bankKeeper.SendCoins(ctx, plan.GetRewardPoolAddress(), reward.GetFarmerAddress(), reward.RewardCoins); err != nil {
		panic(err)
	}

	return types.Reward{}, nil
}
