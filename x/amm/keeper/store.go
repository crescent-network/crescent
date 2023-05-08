package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) GetLastPoolId(ctx sdk.Context) (poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPoolIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastPoolIdKey, sdk.Uint64ToBigEndian(poolId))
}

func (k Keeper) GetNextPoolIdWithUpdate(ctx sdk.Context) uint64 {
	poolId := k.GetLastPoolId(ctx)
	poolId++
	k.SetLastPoolId(ctx, poolId)
	return poolId
}

func (k Keeper) GetLastPositionId(ctx sdk.Context) (positionId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPositionIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastPositionId(ctx sdk.Context, positionId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastPositionIdKey, sdk.Uint64ToBigEndian(positionId))
}

func (k Keeper) GetNextPositionIdWithUpdate(ctx sdk.Context) uint64 {
	positionId := k.GetLastPositionId(ctx)
	positionId++
	k.SetLastPositionId(ctx, positionId)
	return positionId
}

func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolKey(poolId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &pool)
	return pool, true
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.GetPoolKey(pool.Id), bz)
}

func (k Keeper) SetPoolsByMarketIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPoolsByMarketIndexKey(pool.MarketId, pool.Id), []byte{})
}

func (k Keeper) SetPoolByReserveAddressIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetPoolByReserveAddressIndexKey(sdk.MustAccAddressFromBech32(pool.ReserveAddress)),
		sdk.Uint64ToBigEndian(pool.Id))
}

func (k Keeper) IterateAllPools(ctx sdk.Context, cb func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var pool types.Pool
		k.cdc.MustUnmarshal(iter.Value(), &pool)
		if cb(pool) {
			break
		}
	}
}

func (k Keeper) IteratePoolsByMarket(ctx sdk.Context, marketId uint64, cb func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetPoolsByMarketIteratorPrefix(marketId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, poolId := types.ParsePoolsByMarketIndexKey(iter.Key())
		pool, found := k.GetPool(ctx, poolId)
		if !found { // sanity check
			panic("pool not found")
		}
		if cb(pool) {
			break
		}
	}
}

func (k Keeper) GetPoolByReserveAddress(ctx sdk.Context, reserveAddr sdk.AccAddress) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolByReserveAddressIndexKey(reserveAddr))
	if bz == nil {
		return
	}
	return k.GetPool(ctx, sdk.BigEndianToUint64(bz))
}

func (k Keeper) GetPoolState(ctx sdk.Context, poolId uint64) (state types.PoolState, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolStateKey(poolId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &state)
	return state, true
}

func (k Keeper) MustGetPoolState(ctx sdk.Context, poolId uint64) types.PoolState {
	state, found := k.GetPoolState(ctx, poolId)
	if !found {
		panic("pool state not found")
	}
	return state
}

func (k Keeper) SetPoolState(ctx sdk.Context, poolId uint64, state types.PoolState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.GetPoolStateKey(poolId), bz)
}

func (k Keeper) GetPosition(ctx sdk.Context, positionId uint64) (position types.Position, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPositionKey(positionId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &position)
	return position, true
}

func (k Keeper) GetPositionByParams(ctx sdk.Context, ownerAddr sdk.AccAddress, poolId uint64, lowerTick, upperTick int32) (position types.Position, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPositionByParamsIndexKey(ownerAddr, poolId, lowerTick, upperTick))
	if bz == nil {
		return
	}
	return k.GetPosition(ctx, sdk.BigEndianToUint64(bz))
}

func (k Keeper) SetPosition(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&position)
	store.Set(types.GetPositionKey(position.Id), bz)
}

func (k Keeper) SetPositionByParamsIndex(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPositionByParamsIndexKey(
		sdk.MustAccAddressFromBech32(position.Owner), position.PoolId,
		position.LowerTick, position.UpperTick),
		sdk.Uint64ToBigEndian(position.Id))
}

func (k Keeper) IterateAllPositions(ctx sdk.Context, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PositionKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var pool types.Position
		k.cdc.MustUnmarshal(iter.Value(), &pool)
		if cb(pool) {
			break
		}
	}
}

func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tick int32) (tickInfo types.TickInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTickInfoKey(poolId, tick))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &tickInfo)
	return tickInfo, true
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tick int32, tickInfo types.TickInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&tickInfo)
	store.Set(types.GetTickInfoKey(poolId, tick), bz)
}

func (k Keeper) IterateAllTickInfos(ctx sdk.Context, cb func(poolId uint64, tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.TickInfoKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		poolId, tick := types.ParseTickInfoKey(iter.Key())
		var tickInfo types.TickInfo
		k.cdc.MustUnmarshal(iter.Value(), &tickInfo)
		if cb(poolId, tick, tickInfo) {
			break
		}
	}
}

func (k Keeper) IterateTickInfosBelow(ctx sdk.Context, poolId uint64, currentTick int32, cb func(tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.ReverseIterator(
		types.GetTickInfoKeyPrefix(poolId),
		types.GetTickInfoKey(poolId, currentTick))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, tick := types.ParseTickInfoKey(iter.Key())
		var tickInfo types.TickInfo
		k.cdc.MustUnmarshal(iter.Value(), &tickInfo)
		if cb(tick, tickInfo) {
			break
		}
	}
}

func (k Keeper) IterateTickInfosBelowInclusive(ctx sdk.Context, poolId uint64, currentTick int32, cb func(tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.ReverseIterator(
		types.GetTickInfoKeyPrefix(poolId),
		types.GetTickInfoKey(poolId, currentTick+1))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, tick := types.ParseTickInfoKey(iter.Key())
		var tickInfo types.TickInfo
		k.cdc.MustUnmarshal(iter.Value(), &tickInfo)
		if cb(tick, tickInfo) {
			break
		}
	}
}

func (k Keeper) IterateTickInfosAbove(ctx sdk.Context, poolId uint64, currentTick int32, cb func(tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator(
		types.GetTickInfoKey(poolId, currentTick+1),
		sdk.PrefixEndBytes(types.GetTickInfoKeyPrefix(poolId)))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, tick := types.ParseTickInfoKey(iter.Key())
		var tickInfo types.TickInfo
		k.cdc.MustUnmarshal(iter.Value(), &tickInfo)
		if cb(tick, tickInfo) {
			break
		}
	}
}

func (k Keeper) DeleteTickInfo(ctx sdk.Context, poolId uint64, tick int32) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTickInfoKey(poolId, tick))
}
