package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetLastPairId returns the global pair id counter.
func (k Keeper) GetLastPairId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PairIdKey)
	if bz == nil {
		id = 0 // initialize the pair id
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

// GetPair returns pair object for the given pair id.
func (k Keeper) GetPair(ctx sdk.Context, id uint64) (pair types.Pair, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPairKey(id)

	value := store.Get(key)
	if value == nil {
		return pair, false
	}

	pair = types.MustUnmarshalPair(k.cdc, value)

	return pair, true
}

// GetAllPairs returns all pairs in the store.
func (k Keeper) GetAllPairs(ctx sdk.Context) (pairs []types.Pair) {
	k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool) {
		pairs = append(pairs, pair)
		return false
	})

	return pairs
}

// SetLastPairId stores the global pair id counter.
func (k Keeper) SetLastPairId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.PairIdKey, bz)
}

// SetPair stores the particular pair.
func (k Keeper) SetPair(ctx sdk.Context, pair types.Pair) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalPair(k.cdc, pair)
	store.Set(types.GetPairKey(pair.Id), bz)
	k.SetPairIndex(ctx, pair.XCoinDenom, pair.YCoinDenom, pair.Id)
	k.SetPairIndex(ctx, pair.YCoinDenom, pair.XCoinDenom, pair.Id)
}

// SetPairIndex stores the particular denom pair.
func (k Keeper) SetPairIndex(ctx sdk.Context, denomA string, denomB string, pairId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPairIndexKey(denomA, denomB, pairId), []byte{})
}

// IterateAllPairs iterates over all the stored pairs and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllPairs(ctx sdk.Context, cb func(pair types.Pair) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.PairKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		pair := types.MustUnmarshalPair(k.cdc, iter.Value())
		if cb(pair) {
			break
		}
	}
}

// IteratePairsByDenom iterates over all the stored pairs by particular denomination and
// performs a callback function. Stops iteration when callback returns true.
func (k Keeper) IteratePairsByDenom(ctx sdk.Context, denom string, cb func(pair types.Pair) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.GetPairByDenomKeyPrefix(denom))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_, pairId := types.ParsePairByDenomIndexKey(iter.Key())
		pair, _ := k.GetPair(ctx, pairId)
		if cb(pair) {
			break
		}
	}
}

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

// GetDepositRequest returns the particular deposit request.
func (k Keeper) GetDepositRequest(ctx sdk.Context, poolId, id uint64) (state types.DepositRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDepositRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshalDepositRequest(k.cdc, value)
	return state, true
}

// GetWithdrawRequest returns the particular withdraw request.
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, poolId, id uint64) (state types.WithdrawRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshaWithdrawRequest(k.cdc, value)
	return state, true
}

// GetSwapRequest returns the particular swap request.
func (k Keeper) GetSwapRequest(ctx sdk.Context, poolId, id uint64) (state types.SwapRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetSwapRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshaSwapRequest(k.cdc, value)
	return state, true
}

// SetDepositRequest stores deposit request for the batch execution.
func (k Keeper) SetDepositRequest(ctx sdk.Context, poolId uint64, id uint64, state types.DepositRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalDepositRequest(k.cdc, state)
	store.Set(types.GetDepositRequestKey(poolId, id), bz)
}

// SetWithdrawRequest stores withdraw request for the batch execution.
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, poolId uint64, id uint64, state types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaWithdrawRequest(k.cdc, state)
	store.Set(types.GetWithdrawRequestKey(poolId, id), bz)
}

// SetSwapRequest stores swap request for the batch execution.
func (k Keeper) SetSwapRequest(ctx sdk.Context, poolId uint64, id uint64, state types.SwapRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaSwapRequest(k.cdc, state)
	store.Set(types.GetDepositRequestKey(poolId, id), bz)
}
