package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// GetLastPairId returns the last pair id.
func (k Keeper) GetLastPairId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPairIdKey)
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

// SetLastPairId stores the last pair id.
func (k Keeper) SetLastPairId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.LastPairIdKey, bz)
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

// GetPairByDenoms returns a types.Pair for given denoms.
func (k Keeper) GetPairByDenoms(ctx sdk.Context, baseCoinDenom, quoteCoinDenom string) (pair types.Pair, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPairIndexKey(baseCoinDenom, quoteCoinDenom))
	if bz == nil {
		return
	}

	var val gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &val)
	pair, found = k.GetPair(ctx, val.Value)
	return
}

// GetAllPairs returns all pairs in the store.
func (k Keeper) GetAllPairs(ctx sdk.Context) (pairs []types.Pair) {
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairs = append(pairs, pair)
		return false, nil
	})

	return pairs
}

// SetPair stores the particular pair.
func (k Keeper) SetPair(ctx sdk.Context, pair types.Pair) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalPair(k.cdc, pair)
	store.Set(types.GetPairKey(pair.Id), bz)
	k.SetPairIndex(ctx, pair.BaseCoinDenom, pair.QuoteCoinDenom, pair.Id)
	k.SetPairLookupIndex(ctx, pair.BaseCoinDenom, pair.QuoteCoinDenom, pair.Id)
	k.SetPairLookupIndex(ctx, pair.QuoteCoinDenom, pair.BaseCoinDenom, pair.Id)
}

// SetPairIndex stores a pair index.
func (k Keeper) SetPairIndex(ctx sdk.Context, baseCoinDenom, quoteCoinDenom string, pairId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: pairId})
	store.Set(types.GetPairIndexKey(baseCoinDenom, quoteCoinDenom), bz)
}

// SetPairLookupIndex stores a pair lookup index for given denoms.
func (k Keeper) SetPairLookupIndex(ctx sdk.Context, denomA string, denomB string, pairId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPairsByDenomsIndexKey(denomA, denomB, pairId), []byte{})
}

// IterateAllPairs iterates over all the stored pairs and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllPairs(ctx sdk.Context, cb func(pair types.Pair) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.PairKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		pair := types.MustUnmarshalPair(k.cdc, iter.Value())
		stop, err := cb(pair)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetLastPoolId returns the last pool id.
func (k Keeper) GetLastPoolId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPoolIdKey)
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

// SetLastPoolId stores the last pool id.
func (k Keeper) SetLastPoolId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.LastPoolIdKey, bz)
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

// GetPoolByReserveAddress returns pool object for the given reserve account address.
func (k Keeper) GetPoolByReserveAddress(ctx sdk.Context, reserveAddr sdk.AccAddress) (pool types.Pool, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPoolByReserveAddressIndexKey(reserveAddr)

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
	_ = k.IterateAllPools(ctx, func(pool types.Pool) (stop bool, err error) {
		pools = append(pools, pool)
		return false, nil
	})
	return
}

// SetPool stores the particular pool.
func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalPool(k.cdc, pool)
	store.Set(types.GetPoolKey(pool.Id), bz)
	k.SetPoolByReserveIndex(ctx, pool)
	k.SetPoolsByPairIndex(ctx, pool)
}

// SetPoolByReserveIndex stores a pool by reserve account index key.
func (k Keeper) SetPoolByReserveIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: pool.Id})
	store.Set(types.GetPoolByReserveAddressIndexKey(pool.GetReserveAddress()), bz)
}

// SetPoolsByPairIndex stores a pool by pair index key.
func (k Keeper) SetPoolsByPairIndex(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPoolsByPairIndexKey(pool.PairId, pool.Id), []byte{})
}

// IterateAllPools iterates over all the stored pools and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllPools(ctx sdk.Context, cb func(pool types.Pool) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		pool := types.MustUnmarshalPool(k.cdc, iter.Value())
		stop, err := cb(pool)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// IteratePoolsByPair iterates over all the stored pools by the pair and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IteratePoolsByPair(ctx sdk.Context, pairId uint64, cb func(pool types.Pool) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.GetPoolsByPairIndexKeyPrefix(pairId))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		poolId := types.ParsePoolsByPairIndexKey(iter.Key())
		pool, _ := k.GetPool(ctx, poolId)
		stop, err := cb(pool)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
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

// SetDepositRequest stores deposit request for the batch execution.
func (k Keeper) SetDepositRequest(ctx sdk.Context, req types.DepositRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalDepositRequest(k.cdc, req)
	store.Set(types.GetDepositRequestKey(req.PoolId, req.Id), bz)
}

// DeleteDepositRequest deletes a deposit request.
func (k Keeper) DeleteDepositRequest(ctx sdk.Context, poolId, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDepositRequestKey(poolId, id))
}

func (k Keeper) GetAllDepositRequests(ctx sdk.Context) (reqs []types.DepositRequest) {
	_ = k.IterateAllDepositRequests(ctx, func(req types.DepositRequest) (stop bool, err error) {
		reqs = append(reqs, req)
		return false, nil
	})
	return
}

func (k Keeper) IterateAllDepositRequests(ctx sdk.Context, cb func(req types.DepositRequest) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DepositRequestKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var req types.DepositRequest
		k.cdc.MustUnmarshal(iter.Value(), &req)
		stop, err := cb(req)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetWithdrawRequest returns the particular withdraw request.
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, poolId, id uint64) (state types.WithdrawRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshalWithdrawRequest(k.cdc, value)
	return state, true
}

// SetWithdrawRequest stores withdraw request for the batch execution.
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, req types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaWithdrawRequest(k.cdc, req)
	store.Set(types.GetWithdrawRequestKey(req.PoolId, req.Id), bz)
}

// DeleteWithdrawRequest deletes a withdraw request.
func (k Keeper) DeleteWithdrawRequest(ctx sdk.Context, poolId, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetWithdrawRequestKey(poolId, id))
}

func (k Keeper) GetAllWithdrawRequests(ctx sdk.Context) (reqs []types.WithdrawRequest) {
	_ = k.IterateAllWithdrawRequests(ctx, func(req types.WithdrawRequest) (stop bool, err error) {
		reqs = append(reqs, req)
		return false, nil
	})
	return
}

func (k Keeper) IterateAllWithdrawRequests(ctx sdk.Context, cb func(req types.WithdrawRequest) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.WithdrawRequestKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var req types.WithdrawRequest
		k.cdc.MustUnmarshal(iter.Value(), &req)
		stop, err := cb(req)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetOrder returns the particular order.
func (k Keeper) GetOrder(ctx sdk.Context, pairId, id uint64) (order types.Order, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(pairId, id)

	value := store.Get(key)
	if value == nil {
		return order, false
	}

	order = types.MustUnmarshalOrder(k.cdc, value)
	return order, true
}

// SetOrder stores an order for the batch execution.
func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaOrder(k.cdc, order)
	store.Set(types.GetOrderKey(order.PairId, order.Id), bz)
}

// DeleteOrder deletes an order.
func (k Keeper) DeleteOrder(ctx sdk.Context, pairId, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOrderKey(pairId, id))
}

func (k Keeper) GetAllOrders(ctx sdk.Context) (reqs []types.Order) {
	_ = k.IterateAllOrders(ctx, func(order types.Order) (stop bool, err error) {
		reqs = append(reqs, order)
		return false, nil
	})
	return
}

func (k Keeper) IterateAllOrders(ctx sdk.Context, cb func(order types.Order) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var req types.Order
		k.cdc.MustUnmarshal(iter.Value(), &req)
		stop, err := cb(req)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

func (k Keeper) IterateOrdersByPair(ctx sdk.Context, pairId uint64, cb func(req types.Order) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrdersByPairKeyPrefix(pairId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var req types.Order
		k.cdc.MustUnmarshal(iter.Value(), &req)
		stop, err := cb(req)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}
