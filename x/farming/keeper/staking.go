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

// GetStaking return a specific staking by farmer address
func (k Keeper) GetStakingByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingByFarmerAddrIndexKey(farmerAcc))
	if bz == nil {
		return staking, false
	}
	id := binary.BigEndian.Uint64(bz)
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
	store.Set(types.GetStakingByFarmerAddrIndexKey(staking.GetFarmerAddress()), staking.IdBytes())
	for _, denom := range staking.Denoms() {
		store.Set(types.GetStakingByStakingCoinDenomIdIndexKey(denom, staking.Id), []byte{})
	}
}

// RemoveStaking removes an staking for the staking mapper store.
func (k Keeper) RemoveStaking(ctx sdk.Context, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetStakingKey(staking.Id))
	store.Delete(types.GetStakingByFarmerAddrIndexKey(staking.GetFarmerAddress()))
	for _, denom := range staking.Denoms() {
		store.Delete(types.GetStakingByStakingCoinDenomIdIndexKey(denom, staking.Id))
	}
}

// RemoveStaking removes an staking for the staking mapper store.
func (k Keeper) RemoveStakingCoinDenomsIndex(ctx sdk.Context, id uint64, denoms []string) {
	store := ctx.KVStore(k.storeKey)
	for _, denom := range denoms {
		store.Delete(types.GetStakingByStakingCoinDenomIdIndexKey(denom, id))
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
	iterator := sdk.KVStorePrefixIterator(store, types.GetStakingByStakingCoinDenomIdIndexPrefix(denom))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		_, id := types.ParseStakingByStakingCoinDenomIdIndexKey(iterator.Key())
		staking, _ := k.GetStaking(ctx, id)
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

	return nil
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
		// TODO: why added coins on Unstake?
		staking.StakedCoins = staking.QueuedCoins.Add(negativeCoins...)
	}

	stakedQueuedCoins := staking.StakedCoins.Add(staking.QueuedCoins...)
	var removedDenoms []string
	for _, coin := range amount {
		if !stakedQueuedCoins.AmountOf(coin.Denom).IsPositive() {
			removedDenoms = append(removedDenoms, coin.Denom)
		}
	}

	k.SetStaking(ctx, staking)
	k.RemoveStakingCoinDenomsIndex(ctx, staking.Id, removedDenoms)

	return staking, nil
}
