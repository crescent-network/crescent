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

func (k Keeper) GetOrderSequence(ctx sdk.Context) (seq uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OrderSequenceKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetOrderSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OrderSequenceKey, sdk.Uint64ToBigEndian(seq))
}

func (k Keeper) GetNextOrderSequence(ctx sdk.Context) (seq uint64) {
	seq = k.GetOrderSequence(ctx)
	seq++
	k.SetOrderSequence(ctx, seq)
	return
}

func (k Keeper) SetSpotLimitOrder(ctx sdk.Context, order types.SpotLimitOrder) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&order)
	store.Set(types.GetSpotLimitOrderKey(order.MarketId, order.Id), bz)
}

func (k Keeper) GetSpotLimitOrder(ctx sdk.Context, marketId, orderId string) (order types.SpotLimitOrder, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSpotLimitOrderKey(marketId, orderId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &order)
	return order, true
}

func (k Keeper) DeleteSpotLimitOrder(ctx sdk.Context, order types.SpotLimitOrder) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetSpotLimitOrderKey(order.MarketId, order.Id))
}

func (k Keeper) SetSpotOrderBookOrder(ctx sdk.Context, order types.SpotLimitOrder) {
	store := ctx.KVStore(k.storeKey)
	var seq uint64
	if order.IsBuy {
		seq = -order.Sequence
	} else {
		seq = order.Sequence
	}
	store.Set(types.GetSpotOrderBookOrderKey(order.MarketId, order.IsBuy, order.Price, seq), []byte(order.Id))
}

func (k Keeper) IterateSpotOrderBook(ctx sdk.Context, marketId string, isBuy bool, priceLimit *sdk.Dec, cb func(order types.SpotLimitOrder) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	var iter sdk.Iterator
	if isBuy {
		var start []byte
		if priceLimit == nil {
			start = types.GetSpotOrderBookIteratorPrefix(marketId, true)
		} else {
			start = types.GetSpotOrderBookIteratorEndBytes(marketId, true, *priceLimit)
		}
		iter = store.ReverseIterator(
			start,
			sdk.PrefixEndBytes(types.GetSpotOrderBookIteratorPrefix(marketId, true)))
	} else {
		var end []byte
		if priceLimit == nil {
			end = sdk.PrefixEndBytes(types.GetSpotOrderBookIteratorPrefix(marketId, false))
		} else {
			end = types.GetSpotOrderBookIteratorEndBytes(marketId, false, *priceLimit)
		}
		iter = store.Iterator(
			types.GetSpotOrderBookIteratorPrefix(marketId, false),
			end)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := string(iter.Value())
		order, found := k.GetSpotLimitOrder(ctx, marketId, orderId)
		if !found {
			panic("order not found")
		}
		if cb(order) {
			break
		}
	}
}

func (k Keeper) DeleteSpotOrderBookOrder(ctx sdk.Context, order types.SpotLimitOrder) {
	store := ctx.KVStore(k.storeKey)
	var seq uint64
	if order.IsBuy {
		seq = -order.Sequence
	} else {
		seq = order.Sequence
	}
	store.Delete(types.GetSpotOrderBookOrderKey(order.MarketId, order.IsBuy, order.Price, seq))
}
