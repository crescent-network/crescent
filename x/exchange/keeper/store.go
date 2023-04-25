package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) GetLastSpotMarketId(ctx sdk.Context) (marketId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastSpotMarketIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastSpotMarketId(ctx sdk.Context, marketId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastSpotMarketIdKey, sdk.Uint64ToBigEndian(marketId))
}

func (k Keeper) GetNextSpotMarketIdWithUpdate(ctx sdk.Context) (marketId uint64) {
	marketId = k.GetLastSpotMarketId(ctx)
	marketId++
	k.SetLastSpotMarketId(ctx, marketId)
	return
}

func (k Keeper) GetSpotMarket(ctx sdk.Context, marketId uint64) (market types.SpotMarket, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotMarketKey(marketId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &market)
	return market, true
}

func (k Keeper) SetSpotMarket(ctx sdk.Context, market types.SpotMarket) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&market)
	store.Set(types.GetSpotMarketKey(market.Id), bz)
}

func (k Keeper) IterateAllSpotMarkets(ctx sdk.Context, cb func(market types.SpotMarket) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.SpotMarketKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var market types.SpotMarket
		k.cdc.MustUnmarshal(iter.Value(), &market)
		if cb(market) {
			break
		}
	}
}

func (k Keeper) GetSpotMarketState(ctx sdk.Context, marketId uint64) (state types.SpotMarketState, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotMarketStateKey(marketId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &state)
	return state, true
}

func (k Keeper) MustGetSpotMarketState(ctx sdk.Context, marketId uint64) types.SpotMarketState {
	state, found := k.GetSpotMarketState(ctx, marketId)
	if !found {
		panic("spot market state not found")
	}
	return state
}

func (k Keeper) SetSpotMarketState(ctx sdk.Context, marketId uint64, state types.SpotMarketState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.GetSpotMarketStateKey(marketId), bz)
}

func (k Keeper) GetSpotMarketByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (market types.SpotMarket, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotMarketByDenomsIndexKey(baseDenom, quoteDenom))
	if bz == nil {
		return
	}
	return k.GetSpotMarket(ctx, sdk.BigEndianToUint64(bz))
}

func (k Keeper) SetSpotMarketByDenomsIndex(ctx sdk.Context, market types.SpotMarket) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetSpotMarketByDenomsIndexKey(market.BaseDenom, market.QuoteDenom), sdk.Uint64ToBigEndian(market.Id))
}

func (k Keeper) GetLastSpotOrderId(ctx sdk.Context) (orderId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastSpotOrderIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastSpotOrderId(ctx sdk.Context, orderId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastSpotOrderIdKey, sdk.Uint64ToBigEndian(orderId))
}

func (k Keeper) GetNextSpotOrderIdWithUpdate(ctx sdk.Context) (orderId uint64) {
	orderId = k.GetLastSpotOrderId(ctx)
	orderId++
	k.SetLastSpotOrderId(ctx, orderId)
	return
}

func (k Keeper) SetSpotOrder(ctx sdk.Context, order types.SpotOrder) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&order)
	store.Set(types.GetSpotOrderKey(order.Id), bz)
}

func (k Keeper) GetSpotOrder(ctx sdk.Context, orderId uint64) (order types.SpotOrder, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotOrderKey(orderId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &order)
	return order, true
}

func (k Keeper) IterateAllSpotOrders(ctx sdk.Context, cb func(order types.SpotOrder) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.SpotOrderKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order types.SpotOrder
		k.cdc.MustUnmarshal(iter.Value(), &order)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) DeleteSpotOrder(ctx sdk.Context, order types.SpotOrder) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetSpotOrderKey(order.Id))
}

func (k Keeper) SetSpotOrderBookOrder(ctx sdk.Context, order types.SpotOrder) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetSpotOrderBookOrderKey(order.MarketId, order.IsBuy, order.Price, order.Id),
		sdk.Uint64ToBigEndian(order.Id))
}

func (k Keeper) IterateSpotOrderBook(ctx sdk.Context, marketId uint64, cb func(order types.SpotOrder) (stop bool)) {
	k.IterateSpotOrderBookSide(ctx, marketId, false, cb)
	k.IterateSpotOrderBookSide(ctx, marketId, true, cb)
}

func (k Keeper) IterateSpotOrderBookSide(ctx sdk.Context, marketId uint64, isBuy bool, cb func(order types.SpotOrder) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	var iter sdk.Iterator
	if isBuy {
		iter = sdk.KVStoreReversePrefixIterator(store, types.GetSpotOrderBookIteratorPrefix(marketId, true))
	} else {
		iter = sdk.KVStorePrefixIterator(store, types.GetSpotOrderBookIteratorPrefix(marketId, false))
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := sdk.BigEndianToUint64(iter.Value())
		order, found := k.GetSpotOrder(ctx, orderId)
		if !found { // sanity check
			panic("order not found")
		}
		if cb(order) {
			break
		}
	}
}

func (k Keeper) DeleteSpotOrderBookOrder(ctx sdk.Context, order types.SpotOrder) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(
		types.GetSpotOrderBookOrderKey(order.MarketId, order.IsBuy, order.Price, order.Id))
}

func (k Keeper) SetTransientSpotOrderBookOrder(ctx sdk.Context, order types.TransientSpotOrder) {
	store := ctx.TransientStore(k.tsKey)
	bz := k.cdc.MustMarshal(&order)
	store.Set(types.GetSpotOrderBookOrderKey(order.Order.MarketId, order.Order.IsBuy, order.Order.Price, order.Order.Id), bz)
}

func (k Keeper) IterateTransientSpotOrderBookSide(ctx sdk.Context, marketId uint64, isBuy bool, cb func(order types.TransientSpotOrder) (stop bool)) {
	store := ctx.TransientStore(k.tsKey)
	var iter sdk.Iterator
	if isBuy {
		iter = sdk.KVStoreReversePrefixIterator(store, types.GetSpotOrderBookIteratorPrefix(marketId, true))
	} else {
		iter = sdk.KVStorePrefixIterator(store, types.GetSpotOrderBookIteratorPrefix(marketId, false))
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order types.TransientSpotOrder
		k.cdc.MustUnmarshal(iter.Value(), &order)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) IterateTransientSpotOrderBook(ctx sdk.Context, marketId uint64, cb func(order types.TransientSpotOrder) (stop bool)) {
	k.IterateTransientSpotOrderBookSide(ctx, marketId, false, cb)
	k.IterateTransientSpotOrderBookSide(ctx, marketId, true, cb)
}

func (k Keeper) DeleteTransientSpotOrderBookOrder(ctx sdk.Context, order types.TransientSpotOrder) {
	store := ctx.TransientStore(k.tsKey)
	store.Delete(types.GetSpotOrderBookOrderKey(order.Order.MarketId, order.Order.IsBuy, order.Order.Price, order.Order.Id))
}
