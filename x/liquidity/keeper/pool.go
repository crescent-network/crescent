package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetLastPoolId returns the global pool id counter.
func (k Keeper) GetLastPoolId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PoolIdKey)
	if bz == nil {
		id = 0 // initialize the pool id
	} else {
		val := gogotypes.UInt64Value{}
		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}
		id = val.GetValue()
	}
	return id
}

// GetPool returns pool object for the given pool id.
func (k Keeper) GetPool(ctx sdk.Context, id uint64) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolKey(id)

	value := store.Get(key)
	if value == nil {
		return pool, false
	}

	pool = types.MustUnmarshalPool(k.cdc, value)

	return pool, true
}

// GetPoolByReserveAcc returns pool object for the givern reserve account address.
func (k Keeper) GetPoolByReserveAcc(ctx sdk.Context, reserveAcc sdk.AccAddress) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolByReserveAccKey(reserveAcc)

	value := store.Get(key)
	if value == nil {
		return pool, false
	}

	val := gogotypes.UInt64Value{}
	err := k.cdc.Unmarshal(value, &val)
	if err != nil {
		return pool, false
	}
	poolId := val.GetValue()
	return k.GetPool(ctx, poolId)
}

// GetAllPools returns all pairs in the store.
func (k Keeper) GetAllPools(ctx sdk.Context) (pools []types.Pool) {
	k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
		pools = append(pools, pool)
		return false
	})

	return pools
}

// GetNextPoolIdWithUpdate increments pool id by one and set it.
func (k Keeper) GetNextPoolIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastPoolId(ctx) + 1
	k.SetLastPoolId(ctx, id)
	return id
}

// SetLastPoolId stores the global pool id counter.
func (k Keeper) SetLastPoolId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.PoolIdKey, bz)
}

// SetPool stores the particular pool.
func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalPool(k.cdc, pool)
	store.Set(types.GetPoolKey(pool.Id), bz)
}

// SetPoolByReserveAccKey stores a pool by reserve account index key.
func (k Keeper) SetPoolByReserveAccKey(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: pool.Id})
	store.Set(types.GetPoolByReserveAccKey(pool.GetReserveAddress()), bz)
}

// SetPoolByPairIndexKey stores a pool by pair index key.
func (k Keeper) SetPoolByPairIndexKey(ctx sdk.Context, pairId uint64, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPoolsByPairIndexKey(pairId, poolId), []byte{})
}

// IterateAllPools iterates over all the stored pools and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllPools(ctx sdk.Context, cb func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		pool := types.MustUnmarshalPool(k.cdc, iter.Value())
		if cb(pool) {
			break
		}
	}
}

// IterateAllPools iterates over all the stored pools by the pair and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IteratePoolsByPair(ctx sdk.Context, pairId uint64, cb func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.GetPoolsByPairKey(pairId))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		poolId := types.ParsePoolsByPairIndexKey(iter.Key())
		pool, _ := k.GetPool(ctx, poolId)
		if cb(pool) {
			break
		}
	}
}
