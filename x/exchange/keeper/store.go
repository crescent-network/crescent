package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) GetSpotMarket(ctx sdk.Context, marketId string) (market types.SpotMarket, found bool) {
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

func (k Keeper) GetSpotMarketState(ctx sdk.Context, marketId string) (state types.SpotMarketState, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotMarketStateKey(marketId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &state)
	return state, true
}

func (k Keeper) MustGetSpotMarketState(ctx sdk.Context, marketId string) types.SpotMarketState {
	state, found := k.GetSpotMarketState(ctx, marketId)
	if !found {
		panic("spot market state not found")
	}
	return state
}

func (k Keeper) SetSpotMarketState(ctx sdk.Context, marketId string, state types.SpotMarketState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.GetSpotMarketStateKey(marketId), bz)
}

func (k Keeper) GetLastOrderId(ctx sdk.Context) (orderId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastOrderIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastOrderId(ctx sdk.Context, orderId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastOrderIdKey, sdk.Uint64ToBigEndian(orderId))
}

func (k Keeper) GetNextOrderIdWithUpdate(ctx sdk.Context) (orderId uint64) {
	orderId = k.GetLastOrderId(ctx)
	orderId++
	k.SetLastOrderId(ctx, orderId)
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

func (k Keeper) IterateSpotOrderBook(ctx sdk.Context, marketId string, cb func(order types.SpotOrder) (stop bool)) {
	k.IterateSpotOrderBookSide(ctx, marketId, false, cb)
	k.IterateSpotOrderBookSide(ctx, marketId, true, cb)
}

func (k Keeper) IterateSpotOrderBookSide(ctx sdk.Context, marketId string, isBuy bool, cb func(order types.SpotOrder) (stop bool)) {
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

func (k Keeper) IterateTransientSpotOrderBookSide(ctx sdk.Context, marketId string, isBuy bool, cb func(order types.TransientSpotOrder) (stop bool)) {
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

func (k Keeper) IterateTransientSpotOrderBook(ctx sdk.Context, marketId string, cb func(order types.TransientSpotOrder) (stop bool)) {
	k.IterateTransientSpotOrderBookSide(ctx, marketId, false, cb)
	k.IterateTransientSpotOrderBookSide(ctx, marketId, true, cb)
}

func (k Keeper) DeleteTransientSpotOrderBookOrder(ctx sdk.Context, order types.TransientSpotOrder) {
	store := ctx.TransientStore(k.tsKey)
	store.Delete(types.GetSpotOrderBookOrderKey(order.Order.MarketId, order.Order.IsBuy, order.Order.Price, order.Order.Id))
}
