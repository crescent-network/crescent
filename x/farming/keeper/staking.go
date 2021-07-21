package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (k Keeper) GetNextStakingID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	var lastID uint64
	bz := store.Get(types.GlobalStakingIdKey)
	if bz == nil {
		lastID = 0
	} else {
		var value gogotypes.UInt64Value
		k.cdc.MustUnmarshal(bz, &value)
		lastID = value.Value
	}
	return lastID + 1
}

func (k Keeper) GetNextStakingIDWithUpdate(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	id := k.GetNextStakingID(ctx)
	store.Set(types.GlobalStakingIdKey, k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id}))
	return id
}

// NewStaking sets the index to a given staking
func (k Keeper) NewStaking(ctx sdk.Context, farmer sdk.AccAddress) types.Staking {
	id := k.GetNextStakingIDWithUpdate(ctx)
	return types.Staking{
		Id:          id,
		Farmer:      farmer.String(),
		StakedCoins: sdk.NewCoins(),
		QueuedCoins: sdk.NewCoins(),
	}
}

// GetStaking returns a specific staking identified by id.
func (k Keeper) GetStaking(ctx sdk.Context, id uint64) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingKey(id))
	if bz == nil {
		return staking, false
	}
	k.cdc.MustUnmarshal(bz, &staking)
	return staking, true
}

func (k Keeper) GetStakingIDByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) (id uint64, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingByFarmerIndexKey(farmerAcc))
	if bz == nil {
		return 0, false
	}
	id = binary.BigEndian.Uint64(bz)
	return id, true
}

// GetStakingByFarmer returns a specific staking identified by farmer address.
func (k Keeper) GetStakingByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) (staking types.Staking, found bool) {
	id, found := k.GetStakingIDByFarmer(ctx, farmerAcc)
	if !found {
		return staking, false
	}
	return k.GetStaking(ctx, id)
}

// GetAllStakings returns all stakings in the Keeper.
func (k Keeper) GetAllStakings(ctx sdk.Context) (stakings []types.Staking) {
	k.IterateAllStakings(ctx, func(staking types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// GetStakingsByStakingCoinDenom reads from kvstore and return a specific Staking indexed by given staking coin denomination
func (k Keeper) GetStakingsByStakingCoinDenom(ctx sdk.Context, denom string) (stakings []types.Staking) {
	k.IterateStakingsByStakingCoinDenom(ctx, denom, func(staking types.Staking) bool {
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

// SetStakingIndex implements Staking.
func (k Keeper) SetStakingIndex(ctx sdk.Context, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetStakingByFarmerIndexKey(staking.GetFarmer()), sdk.Uint64ToBigEndian(staking.Id))
	for _, denom := range staking.StakingCoinDenoms() {
		store.Set(types.GetStakingByStakingCoinDenomIndexKey(denom, staking.Id), []byte{})
	}
}

// DeleteStaking deletes a staking.
func (k Keeper) DeleteStaking(ctx sdk.Context, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetStakingKey(staking.Id))
	store.Delete(types.GetStakingByFarmerIndexKey(staking.GetFarmer()))
	k.DeleteStakingCoinDenomsIndex(ctx, staking.Id, staking.StakingCoinDenoms())
}

// DeleteStakingCoinDenomsIndex removes an staking for the staking mapper store.
func (k Keeper) DeleteStakingCoinDenomsIndex(ctx sdk.Context, id uint64, denoms []string) {
	store := ctx.KVStore(k.storeKey)
	for _, denom := range denoms {
		store.Delete(types.GetStakingByStakingCoinDenomIndexKey(denom, id))
	}
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

// IterateStakingsByStakingCoinDenom iterates over all the stored stakings indexed by staking coin denomination and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateStakingsByStakingCoinDenom(ctx sdk.Context, denom string, cb func(staking types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetStakingsByStakingCoinDenomIndexKey(denom))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, id := types.ParseStakingsByStakingCoinDenomIndexKey(iterator.Key())
		staking, _ := k.GetStaking(ctx, id)
		if cb(staking) {
			break
		}
	}
}

// UnmarshalStaking unmarshals a Staking from bytes.
func (k Keeper) UnmarshalStaking(bz []byte) (staking types.Staking, err error) {
	return staking, k.cdc.Unmarshal(bz, &staking)
}

// ReserveStakingCoins sends staking coins to the staking reserve account.
func (k Keeper) ReserveStakingCoins(ctx sdk.Context, farmer sdk.AccAddress, stakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, farmer, k.GetStakingStakingReservePoolAcc(ctx), stakingCoins); err != nil {
		return err
	}
	return nil
}

// ReleaseStakingCoins sends staking coins back to the farmer.
func (k Keeper) ReleaseStakingCoins(ctx sdk.Context, farmer sdk.AccAddress, unstakingCoins sdk.Coins) error {
	if err := k.bankKeeper.SendCoins(ctx, k.GetStakingStakingReservePoolAcc(ctx), farmer, unstakingCoins); err != nil {
		return err
	}
	return nil
}

// Stake stores staking coins to queued coins and it will be processed in the next epoch.
func (k Keeper) Stake(ctx sdk.Context, farmer sdk.AccAddress, amount sdk.Coins) error {
	if err := k.ReserveStakingCoins(ctx, farmer, amount); err != nil {
		return err
	}

	params := k.GetParams(ctx)

	staking, found := k.GetStakingByFarmer(ctx, farmer)
	if !found {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, farmer, types.ModuleName, params.StakingCreationFee); err != nil {
			return err
		}

		staking = k.NewStaking(ctx, farmer)
	}
	staking.QueuedCoins = staking.QueuedCoins.Add(amount...)

	k.SetStaking(ctx, staking)
	k.SetStakingIndex(ctx, staking)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyStakingCoins, amount.String()),
		),
	})

	return nil
}

// Unstake unstakes an amount of staking coins from the staking reserve account.
func (k Keeper) Unstake(ctx sdk.Context, farmer sdk.AccAddress, amount sdk.Coins) error {
	staking, found := k.GetStakingByFarmer(ctx, farmer)
	if !found {
		return types.ErrStakingNotExists
	}

	if err := k.ReleaseStakingCoins(ctx, farmer, amount); err != nil {
		return err
	}

	prevDenoms := staking.StakingCoinDenoms()

	var hasNeg bool
	staking.QueuedCoins, hasNeg = staking.QueuedCoins.SafeSub(amount)
	if hasNeg {
		negativeCoins := sdk.NewCoins()
		for _, coin := range staking.QueuedCoins {
			if coin.IsNegative() {
				negativeCoins = negativeCoins.Add(coin)
			}
		}
		staking.QueuedCoins = staking.QueuedCoins.Sub(negativeCoins)
		staking.StakedCoins = staking.StakedCoins.Add(negativeCoins...)
	}

	// Remove the Staking object from the kvstore when all coins has been unstaked
	// and there's no rewards left.
	if staking.StakedCoins.IsZero() && staking.QueuedCoins.IsZero() && len(k.GetRewardsByFarmer(ctx, farmer)) == 0 {
		k.DeleteStaking(ctx, staking)
	} else {
		k.SetStaking(ctx, staking)

		denomSet := make(map[string]struct{})
		for _, denom := range staking.StakingCoinDenoms() {
			denomSet[denom] = struct{}{}
		}

		var removedDenoms []string
		for _, denom := range prevDenoms {
			if _, ok := denomSet[denom]; !ok {
				removedDenoms = append(removedDenoms, denom)
			}
		}

		k.DeleteStakingCoinDenomsIndex(ctx, staking.Id, removedDenoms)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyUnstakingCoins, amount.String()),
		),
	})

	return nil
}
