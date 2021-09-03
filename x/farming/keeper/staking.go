package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

// GetStaking returns a specific staking identified by id.
func (k Keeper) GetStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingKey(stakingCoinDenom, farmerAcc))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &staking)
	found = true
	return
}

//// GetAllStakings returns all stakings in the Keeper.
//func (k Keeper) GetAllStakings(ctx sdk.Context) (stakings []types.Staking) {
//	k.IterateAllStakings(ctx, func(staking types.Staking) (stop bool) {
//		stakings = append(stakings, staking)
//		return false
//	})
//
//	return stakings
//}

//// GetStakingsByStakingCoinDenom reads from kvstore and return a specific Staking indexed by given staking coin denomination
//func (k Keeper) GetStakingsByStakingCoinDenom(ctx sdk.Context, denom string) (stakings []types.Staking) {
//	k.IterateStakingsByStakingCoinDenom(ctx, denom, func(staking types.Staking) bool {
//		stakings = append(stakings, staking)
//		return false
//	})
//	return stakings
//}

// SetStaking implements Staking.
func (k Keeper) SetStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&staking)
	store.Set(types.GetStakingKey(stakingCoinDenom, farmerAcc), bz)
}

//// SetStakingIndex implements Staking.
//func (k Keeper) SetStakingIndex(ctx sdk.Context, staking types.Staking) {
//	store := ctx.KVStore(k.storeKey)
//	store.Set(types.GetStakingByFarmerIndexKey(staking.GetFarmer()), sdk.Uint64ToBigEndian(staking.Id))
//	for _, denom := range staking.StakingCoinDenoms() {
//		store.Set(types.GetStakingByStakingCoinDenomIndexKey(denom, staking.Id), []byte{})
//	}
//}

//// DeleteStaking deletes a staking.
//func (k Keeper) DeleteStaking(ctx sdk.Context, staking types.Staking) {
//	store := ctx.KVStore(k.storeKey)
//	store.Delete(types.GetStakingKey(staking.Id))
//	store.Delete(types.GetStakingByFarmerIndexKey(staking.GetFarmer()))
//	if denoms := staking.StakingCoinDenoms(); len(denoms) > 0 {
//		k.DeleteStakingCoinDenomsIndex(ctx, staking.Id, denoms)
//	}
//}

//// DeleteStakingCoinDenomsIndex removes an staking for the staking mapper store.
//func (k Keeper) DeleteStakingCoinDenomsIndex(ctx sdk.Context, id uint64, denoms []string) {
//	store := ctx.KVStore(k.storeKey)
//	for _, denom := range denoms {
//		store.Delete(types.GetStakingByStakingCoinDenomIndexKey(denom, id))
//	}
//}

//// IterateAllStakings iterates over all the stored stakings and performs a callback function.
//// Stops iteration when callback returns true.
//func (k Keeper) IterateAllStakings(ctx sdk.Context, cb func(staking types.Staking) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iterator := sdk.KVStorePrefixIterator(store, types.StakingKeyPrefix)
//
//	defer iterator.Close()
//	for ; iterator.Valid(); iterator.Next() {
//		var staking types.Staking
//		k.cdc.MustUnmarshal(iterator.Value(), &staking)
//		if cb(staking) {
//			break
//		}
//	}
//}

//// IterateStakingsByStakingCoinDenom iterates over all the stored stakings indexed by staking coin denomination and performs a callback function.
//// Stops iteration when callback returns true.
//func (k Keeper) IterateStakingsByStakingCoinDenom(ctx sdk.Context, denom string, cb func(staking types.Staking) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iterator := sdk.KVStorePrefixIterator(store, types.GetStakingsByStakingCoinDenomIndexKey(denom))
//	defer iterator.Close()
//	for ; iterator.Valid(); iterator.Next() {
//		_, id := types.ParseStakingsByStakingCoinDenomIndexKey(iterator.Key())
//		staking, _ := k.GetStaking(ctx, id)
//		if cb(staking) {
//			break
//		}
//	}
//}

func (k Keeper) GetQueuedStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (queuedStaking types.QueuedStaking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueuedStakingKey(stakingCoinDenom, farmerAcc))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &queuedStaking)
	found = true
	return
}

func (k Keeper) SetQueuedStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&queuedStaking)
	store.Set(types.GetQueuedStakingKey(stakingCoinDenom, farmerAcc), bz)
}

func (k Keeper) DeleteQueuedStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetQueuedStakingKey(stakingCoinDenom, farmerAcc))
}

func (k Keeper) IterateQueuedStakings(ctx sdk.Context, cb func(stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.QueuedStakingKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var queuedStaking types.QueuedStaking
		k.cdc.MustUnmarshal(iter.Value(), &queuedStaking)
		stakingCoinDenom, farmerAcc := types.ParseQueuedStakingKey(iter.Key())
		if cb(stakingCoinDenom, farmerAcc, queuedStaking) {
			break
		}
	}
}

func (k Keeper) GetTotalStaking(ctx sdk.Context, stakingCoinDenom string) (totalStaking types.TotalStaking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalStakingKey(stakingCoinDenom))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &totalStaking)
	found = true
	return
}

func (k Keeper) SetTotalStaking(ctx sdk.Context, stakingCoinDenom string, totalStaking types.TotalStaking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalStaking)
	store.Set(types.GetTotalStakingKey(stakingCoinDenom), bz)
}

// ReserveStakingCoins sends staking coins to the staking reserve account.
func (k Keeper) ReserveStakingCoins(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, farmerAcc, k.GetStakingReservePoolAcc(ctx), stakingCoins); err != nil {
		return err
	}
	return nil
}

// ReleaseStakingCoins sends staking coins back to the farmer.
func (k Keeper) ReleaseStakingCoins(ctx sdk.Context, farmerAcc sdk.AccAddress, unstakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, k.GetStakingReservePoolAcc(ctx), farmerAcc, unstakingCoins); err != nil {
		return err
	}
	return nil
}

// Stake stores staking coins to queued coins, and it will be processed in the next epoch.
func (k Keeper) Stake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error {
	if err := k.ReserveStakingCoins(ctx, farmerAcc, amount); err != nil {
		return err
	}

	for _, coin := range amount {
		queuedStaking, found := k.GetQueuedStaking(ctx, coin.Denom, farmerAcc)
		if !found {
			queuedStaking.Amount = sdk.ZeroInt()
		}
		queuedStaking.Amount = queuedStaking.Amount.Add(coin.Amount)
		k.SetQueuedStaking(ctx, coin.Denom, farmerAcc, queuedStaking)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStakingCoins, amount.String()),
		),
	})

	return nil
}

// Unstake unstakes an amount of staking coins from the staking reserve account.
func (k Keeper) Unstake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error {
	//staking, found := k.GetStakingByFarmer(ctx, farmerAcc)
	//if !found {
	//	return types.Staking{}, types.ErrStakingNotExists
	//}
	//
	//if err := k.ReleaseStakingCoins(ctx, farmerAcc, amount); err != nil {
	//	return types.Staking{}, err
	//}
	//
	//prevDenoms := staking.StakingCoinDenoms()
	//
	//var hasNeg bool
	//staking.QueuedCoins, hasNeg = staking.QueuedCoins.SafeSub(amount)
	//if hasNeg {
	//	negativeCoins := sdk.NewCoins()
	//	for _, coin := range staking.QueuedCoins {
	//		if coin.IsNegative() {
	//			negativeCoins = negativeCoins.Add(coin)
	//		}
	//	}
	//	staking.QueuedCoins = staking.QueuedCoins.Sub(negativeCoins)
	//	staking.StakedCoins = staking.StakedCoins.Add(negativeCoins...)
	//}
	//
	//// Remove the Staking object from the kvstore when all coins has been unstaked
	//// and there's no rewards left.
	//// TODO: find more efficient way to check if the farmerAcc has no rewards
	//if staking.StakedCoins.IsZero() && staking.QueuedCoins.IsZero() {
	//	k.DeleteStaking(ctx, staking)
	//} else {
	//	k.SetStaking(ctx, staking)
	//
	//	denomSet := make(map[string]struct{})
	//	for _, denom := range staking.StakingCoinDenoms() {
	//		denomSet[denom] = struct{}{}
	//	}
	//
	//	var removedDenoms []string
	//	for _, denom := range prevDenoms {
	//		if _, ok := denomSet[denom]; !ok {
	//			removedDenoms = append(removedDenoms, denom)
	//		}
	//	}
	//
	//	if len(removedDenoms) > 0 {
	//		k.DeleteStakingCoinDenomsIndex(ctx, staking.Id, removedDenoms)
	//	}
	//}
	//
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypeUnstake,
	//		sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
	//		sdk.NewAttribute(types.AttributeKeyUnstakingCoins, amount.String()),
	//	),
	//})
	//
	//// We're returning a Staking even if it has deleted.
	//return staking, nil
	return nil
}

// ProcessQueuedCoins moves queued coins into staked coins.
func (k Keeper) ProcessQueuedCoins(ctx sdk.Context) {
	k.IterateQueuedStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
		if found {
			if _, err := k.WithdrawRewards(ctx, farmerAcc, stakingCoinDenom); err != nil {
				panic(err)
			}
		} else {
			staking.Amount = sdk.ZeroInt()
		}

		k.DeleteQueuedStaking(ctx, stakingCoinDenom, farmerAcc)
		k.SetStaking(ctx, stakingCoinDenom, farmerAcc, types.Staking{
			Amount:        staking.Amount.Add(queuedStaking.Amount),
			StartingEpoch: k.GetCurrentRewards(ctx, stakingCoinDenom).Epoch,
		})

		totalStaking, found := k.GetTotalStaking(ctx, stakingCoinDenom)
		if !found {
			totalStaking.Amount = sdk.ZeroInt()
		}
		k.SetTotalStaking(ctx, stakingCoinDenom, types.TotalStaking{
			Amount: totalStaking.Amount.Add(queuedStaking.Amount),
		})

		return false
	})
}
