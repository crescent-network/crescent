package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

//// NewStaking sets the index to a given staking
//func (k Keeper) NewStaking(ctx sdk.Context, staking types.Staking) types.Staking {
//	k.SetPlanIdByFarmerAddrIndex(ctx, staking.PlanId, staking.GetFarmerAddress())
//	return staking
//}

// GetStaking return a specific staking
func (k Keeper) GetStaking(ctx sdk.Context, id uint64) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingKey(id))
	if bz == nil {
		return staking, false
	}
	k.cdc.MustUnmarshal(bz, &staking)
	return staking, true
}

// GetStaking return a specific staking
func (k Keeper) GetStakingByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingByFarmerAddrIndexKey(farmerAcc))
	if bz == nil {
		return staking, false
	}
	k.cdc.MustUnmarshal(bz, &staking)
	return staking, true
}

// GetAllStakings returns all stakings in the Keeper.
func (k Keeper) GetAllStakings(ctx sdk.Context) (stakings []types.Staking) {
	k.IterateAllStakings(ctx, func(staking types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// GetStakingsByFarmer reads from kvstore and return a specific Staking indexed by given farmer address
func (k Keeper) GetStakingsByFarmer(ctx sdk.Context, farmer sdk.AccAddress) (stakings []types.Staking) {
	k.IterateStakingsByFarmer(ctx, farmer, func(staking types.Staking) bool {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// SetStaking implements Staking.
func (k Keeper) SetStaking(ctx sdk.Context, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&staking)
	store.Set(types.GetStakingKey(staking.Id), bz)
}

// RemoveStaking removes an staking for the staking mapper store.
func (k Keeper) RemoveStaking(ctx sdk.Context, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetStakingKey(staking.Id))
}

// IterateAllStakings iterates over all the stored stakings and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllStakings(ctx sdk.Context, cb func(staking types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.StakingKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var staking types.Staking
		k.cdc.MustUnmarshal(iterator.Value(), &staking)
		if cb(staking) {
			break
		}
	}
}

// IterateStakingsByFarmer iterates over all the stored stakings indexed by farmer and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateStakingsByFarmer(ctx sdk.Context, farmer sdk.AccAddress, cb func(staking types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetStakingByFarmerAddrIndexKey(farmer))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var staking types.Staking
		k.cdc.MustUnmarshal(iterator.Value(), &staking)
		if cb(staking) {
			break
		}
	}
}

// UnmarshalStaking unmarshals a Staking from bytes.
func (k Keeper) UnmarshalStaking(bz []byte) (types.Staking, error) {
	var staking types.Staking
	return staking, k.cdc.Unmarshal(bz, &staking)
}

// ReserveStakingCoins sends staking coins to the staking reserve account.
func (k Keeper) ReserveStakingCoins(ctx sdk.Context, farmer sdk.AccAddress, stakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, farmer, types.StakingReserveAcc, stakingCoins); err != nil {
		return err
	}
	return nil
}

// ReleaseStakingCoins sends staking coins back to the farmer.
func (k Keeper) ReleaseStakingCoins(ctx sdk.Context, farmer sdk.AccAddress, unstakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, types.StakingReserveAcc, farmer, unstakingCoins); err != nil {
		return err
	}
	return nil
}

// Stake stores staking coins to queued coins and it will be processed in the next epoch.
func (k Keeper) Stake(ctx sdk.Context, farmer sdk.AccAddress, amount sdk.Coins) (types.Staking, error) {
	if err := k.ReserveStakingCoins(ctx, farmer, amount); err != nil {
		return types.Staking{}, err
	}

	staking, found := k.GetStakingByFarmer(ctx, farmer)
	if !found {
		staking = types.Staking{
			Farmer:      farmer.String(),
			StakedCoins: sdk.NewCoins(),
			QueuedCoins: amount,
		}
	} else {
		staking.QueuedCoins = staking.QueuedCoins.Add(amount...)
	}

	k.SetStaking(ctx, staking)

	return staking, nil
}

// Unstake unstakes an amount of staking coins from the staking reserve account.
func (k Keeper) Unstake(ctx sdk.Context, farmer sdk.AccAddress, amount sdk.Coins) (types.Staking, error) {
	staking, found := k.GetStakingByFarmer(ctx, farmer)
	if !found {
		return types.Staking{}, types.ErrStakingNotExists
	}

	if err := k.ReleaseStakingCoins(ctx, farmer, amount); err != nil {
		return types.Staking{}, err
	}

	var hasNeg bool
	staking.QueuedCoins, hasNeg = staking.QueuedCoins.SafeSub(amount)
	if hasNeg {
		negativeCoins := sdk.NewCoins()
		for _, coin := range staking.QueuedCoins {
			if coin.IsNegative() {
				negativeCoins = negativeCoins.Add(coin)
				staking.QueuedCoins = staking.QueuedCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.Neg()))
				staking.StakedCoins = staking.StakedCoins.Add(coin)
			}
		}
		staking.QueuedCoins = staking.QueuedCoins.Sub(negativeCoins)
		staking.StakedCoins = staking.QueuedCoins.Add(negativeCoins...)
	}

	k.SetStaking(ctx, staking)

	return staking, nil
}
