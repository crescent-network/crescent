package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
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

func (k Keeper) GetNextPoolIdWithUpdate(ctx sdk.Context) (poolId uint64) {
	poolId = k.GetLastPoolId(ctx)
	poolId++
	k.SetLastPoolId(ctx, poolId)
	return poolId
}

func (k Keeper) DeleteLastPoolId(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.LastPoolIdKey)
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

func (k Keeper) GetNextPositionIdWithUpdate(ctx sdk.Context) (positionId uint64) {
	positionId = k.GetLastPositionId(ctx)
	positionId++
	k.SetLastPositionId(ctx, positionId)
	return positionId
}

func (k Keeper) DeleteLastPositionId(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.LastPositionIdKey)
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

func (k Keeper) MustGetPool(ctx sdk.Context, poolId uint64) (pool types.Pool) {
	pool, found := k.GetPool(ctx, poolId)
	if !found {
		panic("pool not found")
	}
	return pool
}

func (k Keeper) LookupPool(ctx sdk.Context, poolId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPoolKey(poolId))
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.GetPoolKey(pool.Id), bz)
}

func (k Keeper) GetPoolByMarket(ctx sdk.Context, marketId uint64) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolByMarketIndexKey(marketId))
	if bz == nil {
		return
	}
	return k.GetPool(ctx, sdk.BigEndianToUint64(bz))
}

func (k Keeper) LookupPoolByMarket(ctx sdk.Context, marketId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPoolByMarketIndexKey(marketId))
}

func (k Keeper) SetPoolByMarketIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPoolByMarketIndexKey(pool.MarketId), sdk.Uint64ToBigEndian(pool.Id))
}

func (k Keeper) GetPoolByReserveAddress(ctx sdk.Context, reserveAddr sdk.AccAddress) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolByReserveAddressIndexKey(reserveAddr))
	if bz == nil {
		return
	}
	return k.GetPool(ctx, sdk.BigEndianToUint64(bz))
}

func (k Keeper) MustGetPoolByReserveAddress(ctx sdk.Context, reserveAddr sdk.AccAddress) (pool types.Pool) {
	pool, found := k.GetPoolByReserveAddress(ctx, reserveAddr)
	if !found {
		panic("pool not found")
	}
	return pool
}

func (k Keeper) DeletePool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPoolKey(pool.Id))
}

func (k Keeper) SetPoolByReserveAddressIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetPoolByReserveAddressIndexKey(pool.MustGetReserveAddress()),
		sdk.Uint64ToBigEndian(pool.Id))
}

func (k Keeper) DeletePoolByReserveAddressIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(
		types.GetPoolByReserveAddressIndexKey(pool.MustGetReserveAddress()))
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

func (k Keeper) DeletePoolState(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPoolStateKey(poolId))
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

func (k Keeper) MustGetPosition(ctx sdk.Context, positionId uint64) (position types.Position) {
	position, found := k.GetPosition(ctx, positionId)
	if !found {
		panic("position not found")
	}
	return position
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
		position.MustGetOwnerAddress(), position.PoolId,
		position.LowerTick, position.UpperTick),
		sdk.Uint64ToBigEndian(position.Id))
}

func (k Keeper) SetPositionsByPoolIndex(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPositionsByPoolIndexKey(position.PoolId, position.Id), []byte{})
}

func (k Keeper) IterateAllPositions(ctx sdk.Context, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PositionKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var pool types.Position // pool -> position(mis-spell)
		k.cdc.MustUnmarshal(iter.Value(), &pool)
		if cb(pool) {
			break
		}
	}
}

func (k Keeper) IteratePositionsByPool(ctx sdk.Context, poolId uint64, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetPositionsByPoolIteratorPrefix(poolId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, positionId := types.ParsePositionsByPoolIndexKey(iter.Key())
		position := k.MustGetPosition(ctx, positionId)
		if cb(position) {
			break
		}
	}
}

func (k Keeper) IteratePositionsByOwner(ctx sdk.Context, ownerAddr sdk.AccAddress, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetPositionsByOwnerIteratorPrefix(ownerAddr))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		positionId := sdk.BigEndianToUint64(iter.Value())
		position := k.MustGetPosition(ctx, positionId)
		if cb(position) {
			break
		}
	}
}

func (k Keeper) IteratePositionsByOwnerAndPool(ctx sdk.Context, ownerAddr sdk.AccAddress, poolId uint64, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetPositionsByOwnerAndPoolIteratorPrefix(ownerAddr, poolId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		positionId := sdk.BigEndianToUint64(iter.Value())
		position := k.MustGetPosition(ctx, positionId)
		if cb(position) {
			break
		}
	}
}

func (k Keeper) DeletePosition(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPositionKey(position.Id))
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

func (k Keeper) MustGetTickInfo(ctx sdk.Context, poolId uint64, tick int32) (tickInfo types.TickInfo) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found {
		panic("tick info not found")
	}
	return tickInfo
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

func (k Keeper) IterateTickInfosByPool(ctx sdk.Context, poolId uint64, cb func(tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetTickInfosByPoolIteratorPrefix(poolId))
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

func (k Keeper) IterateTickInfosBelow(ctx sdk.Context, poolId uint64, currentTick int32, inclusive bool, cb func(tick int32, tickInfo types.TickInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	end := types.GetTickInfoKey(poolId, currentTick)
	if inclusive {
		end = sdk.PrefixEndBytes(end)
	}
	iter := store.ReverseIterator(types.GetTickInfosByPoolIteratorPrefix(poolId), end)
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
		sdk.PrefixEndBytes(types.GetTickInfosByPoolIteratorPrefix(poolId)))
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

func (k Keeper) GetLastFarmingPlanId(ctx sdk.Context) (planId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastFarmingPlanIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastFarmingPlanId(ctx sdk.Context, planId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastFarmingPlanIdKey, sdk.Uint64ToBigEndian(planId))
}

func (k Keeper) GetNextFarmingPlanIdWithUpdate(ctx sdk.Context) (planId uint64) {
	planId = k.GetLastFarmingPlanId(ctx)
	planId++
	k.SetLastFarmingPlanId(ctx, planId)
	return planId
}

func (k Keeper) DeleteLastFarmingPlanId(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.LastFarmingPlanIdKey)
}

func (k Keeper) GetFarmingPlan(ctx sdk.Context, planId uint64) (plan types.FarmingPlan, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFarmingPlanKey(planId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &plan)
	return plan, true
}

func (k Keeper) SetFarmingPlan(ctx sdk.Context, plan types.FarmingPlan) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFarmingPlanKey(plan.Id), k.cdc.MustMarshal(&plan))
}

func (k Keeper) IterateAllFarmingPlans(ctx sdk.Context, cb func(plan types.FarmingPlan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FarmingPlanKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var plan types.FarmingPlan
		k.cdc.MustUnmarshal(iter.Value(), &plan)
		if cb(plan) {
			break
		}
	}
}

func (k Keeper) DeleteFarmingPlan(ctx sdk.Context, plan types.FarmingPlan) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetFarmingPlanKey(plan.Id))
}

func (k Keeper) GetNumPrivateFarmingPlans(ctx sdk.Context) (num uint32) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NumPrivateFarmingPlansKey)
	if bz == nil {
		return 0
	}
	return utils.BigEndianToUint32(bz)
}

func (k Keeper) SetNumPrivateFarmingPlans(ctx sdk.Context, num uint32) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NumPrivateFarmingPlansKey, utils.Uint32ToBigEndian(num))
}

func (k Keeper) DeleteNumPrivateFarmingPlans(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.NumPrivateFarmingPlansKey)
}
