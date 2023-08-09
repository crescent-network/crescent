package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) GetLastMarketId(ctx sdk.Context) (marketId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastMarketIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastMarketId(ctx sdk.Context, marketId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastMarketIdKey, sdk.Uint64ToBigEndian(marketId))
}

func (k Keeper) GetNextMarketIdWithUpdate(ctx sdk.Context) (marketId uint64) {
	marketId = k.GetLastMarketId(ctx)
	marketId++
	k.SetLastMarketId(ctx, marketId)
	return
}

func (k Keeper) GetMarket(ctx sdk.Context, marketId uint64) (market types.Market, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMarketKey(marketId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &market)
	return market, true
}

func (k Keeper) MustGetMarket(ctx sdk.Context, marketId uint64) (market types.Market) {
	market, found := k.GetMarket(ctx, marketId)
	if !found {
		panic("market not found")
	}
	return market
}

func (k Keeper) LookupMarket(ctx sdk.Context, marketId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetMarketKey(marketId))
}

func (k Keeper) SetMarket(ctx sdk.Context, market types.Market) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&market)
	store.Set(types.GetMarketKey(market.Id), bz)
}

func (k Keeper) IterateAllMarkets(ctx sdk.Context, cb func(market types.Market) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.MarketKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var market types.Market
		k.cdc.MustUnmarshal(iter.Value(), &market)
		if cb(market) {
			break
		}
	}
}

func (k Keeper) GetMarketState(ctx sdk.Context, marketId uint64) (state types.MarketState, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMarketStateKey(marketId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &state)
	return state, true
}

func (k Keeper) MustGetMarketState(ctx sdk.Context, marketId uint64) types.MarketState {
	state, found := k.GetMarketState(ctx, marketId)
	if !found {
		panic(" market state not found")
	}
	return state
}

func (k Keeper) SetMarketState(ctx sdk.Context, marketId uint64, state types.MarketState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.GetMarketStateKey(marketId), bz)
}

func (k Keeper) GetMarketIdByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (marketId uint64, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMarketByDenomsIndexKey(baseDenom, quoteDenom))
	if bz == nil {
		return
	}
	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetMarketByDenomsIndex(ctx sdk.Context, market types.Market) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetMarketByDenomsIndexKey(market.BaseDenom, market.QuoteDenom), sdk.Uint64ToBigEndian(market.Id))
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

func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&order)
	store.Set(types.GetOrderKey(order.Id), bz)
}

func (k Keeper) GetOrder(ctx sdk.Context, orderId uint64) (order types.Order, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOrderKey(orderId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &order)
	return order, true
}

func (k Keeper) MustGetOrder(ctx sdk.Context, orderId uint64) (order types.Order) {
	order, found := k.GetOrder(ctx, orderId)
	if !found {
		panic("order not found")
	}
	return order
}

func (k Keeper) LookupOrder(ctx sdk.Context, orderId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetOrderKey(orderId))
}

func (k Keeper) IterateAllOrders(ctx sdk.Context, cb func(order types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order types.Order
		k.cdc.MustUnmarshal(iter.Value(), &order)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) IterateOrdersByMarket(ctx sdk.Context, marketId uint64, cb func(order types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrdersByMarketIteratorPrefix(marketId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrderBookOrderIndexKey(iter.Key())
		order := k.MustGetOrder(ctx, orderId)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) IterateOrdersByOrderer(ctx sdk.Context, ordererAddr sdk.AccAddress, cb func(order types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrdersByOrdererIteratorPrefix(ordererAddr))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrdersByOrdererIndexKey(iter.Key())
		order := k.MustGetOrder(ctx, orderId)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) IterateOrdersByOrdererAndMarket(ctx sdk.Context, ordererAddr sdk.AccAddress, marketId uint64, cb func(order types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrdersByOrdererAndMarketIteratorPrefix(ordererAddr, marketId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrdersByOrdererIndexKey(iter.Key())
		order := k.MustGetOrder(ctx, orderId)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) DeleteOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOrderKey(order.Id))
}

func (k Keeper) SetOrderBookOrderIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetOrderBookOrderIndexKey(order.MarketId, order.IsBuy, order.Price, order.Id),
		sdk.Uint64ToBigEndian(order.Id))
}

func (k Keeper) LookupOrderBookOrderIndex(ctx sdk.Context, marketId uint64, isBuy bool, price sdk.Dec, orderId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetOrderBookOrderIndexKey(marketId, isBuy, price, orderId))
}

func (k Keeper) IterateAllOrderBookOrderIds(ctx sdk.Context, cb func(orderId uint64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.OrderBookOrderIndexKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrderBookOrderIndexKey(iter.Key())
		if cb(orderId) {
			break
		}
	}
}

func (k Keeper) IterateOrderBookSideByMarket(ctx sdk.Context, marketId uint64, isBuy bool, cb func(order types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	var iter sdk.Iterator
	if isBuy {
		iter = sdk.KVStoreReversePrefixIterator(
			store, types.GetOrderBookSideIteratorPrefix(marketId, true))
	} else {
		iter = sdk.KVStorePrefixIterator(
			store, types.GetOrderBookSideIteratorPrefix(marketId, false))
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrderBookOrderIndexKey(iter.Key())
		order := k.MustGetOrder(ctx, orderId)
		if cb(order) {
			break
		}
	}
}

func (k Keeper) GetBestOrderPrice(ctx sdk.Context, marketId uint64, isBuy bool) (bestPrice sdk.Dec, found bool) {
	store := ctx.KVStore(k.storeKey)
	var iter sdk.Iterator
	if isBuy {
		iter = sdk.KVStoreReversePrefixIterator(
			store, types.GetOrderBookSideIteratorPrefix(marketId, true))
	} else {
		iter = sdk.KVStorePrefixIterator(
			store, types.GetOrderBookSideIteratorPrefix(marketId, false))
	}
	defer iter.Close()
	if iter.Valid() {
		bestPrice = types.ParsePriceFromOrderBookOrderIndexKey(iter.Key())
		return bestPrice, true
	}
	return bestPrice, false
}

func (k Keeper) IterateOrderBookSide(
	ctx sdk.Context, marketId uint64, isBuy bool, priceLimit *sdk.Dec,
	cb func(price sdk.Dec, orders []types.Order) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	var iter sdk.Iterator
	if isBuy {
		if priceLimit == nil {
			iter = sdk.KVStoreReversePrefixIterator(
				store, types.GetOrderBookSideIteratorPrefix(marketId, true))
		} else {
			iter = store.ReverseIterator(
				types.GetOrderBookSidePriceLimitIteratorPrefix(marketId, true, *priceLimit),
				sdk.PrefixEndBytes(
					types.GetOrderBookSideIteratorPrefix(marketId, true)))
		}
	} else {
		if priceLimit == nil {
			iter = sdk.KVStorePrefixIterator(
				store, types.GetOrderBookSideIteratorPrefix(marketId, false))
		} else {
			iter = store.Iterator(
				types.GetOrderBookSideIteratorPrefix(marketId, false),
				sdk.PrefixEndBytes(types.GetOrderBookSidePriceLimitIteratorPrefix(marketId, false, *priceLimit)))
		}
	}
	defer iter.Close()
	var (
		currentPrice sdk.Dec
		orders       []types.Order
	)
	for ; iter.Valid(); iter.Next() {
		orderId := types.ParseOrderIdFromOrderBookOrderIndexKey(iter.Key())
		order := k.MustGetOrder(ctx, orderId)
		if !currentPrice.IsNil() && !order.Price.Equal(currentPrice) {
			if cb(currentPrice, orders) {
				break
			}
			orders = []types.Order{order}
		} else {
			orders = append(orders, order)
		}
		currentPrice = order.Price
	}
	if len(orders) > 0 {
		// Ignore the return value since it's the last iteration.
		_ = cb(currentPrice, orders)
	}
}

func (k Keeper) DeleteOrderBookOrderIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(
		types.GetOrderBookOrderIndexKey(order.MarketId, order.IsBuy, order.Price, order.Id))
}

func (k Keeper) SetOrdersByOrdererIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetOrdersByOrdererIndexKey(order.MustGetOrdererAddress(), order.MarketId, order.Id), []byte{})
}

func (k Keeper) DeleteOrdersByOrdererIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(
		types.GetOrdersByOrdererIndexKey(order.MustGetOrdererAddress(), order.MarketId, order.Id))
}

func (k Keeper) GetNumMMOrders(ctx sdk.Context, ordererAddr sdk.AccAddress, marketId uint64) (num uint32, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetNumMMOrdersKey(ordererAddr, marketId))
	if bz == nil {
		return
	}
	return utils.BigEndianToUint32(bz), true
}

func (k Keeper) SetNumMMOrders(ctx sdk.Context, ordererAddr sdk.AccAddress, marketId uint64, num uint32) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetNumMMOrdersKey(ordererAddr, marketId), utils.Uint32ToBigEndian(num))
}

func (k Keeper) IterateAllNumMMOrders(ctx sdk.Context, cb func(ordererAddr sdk.AccAddress, marketId uint64, numMMOrders uint32) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.NumMMOrdersKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		ordererAddr, marketId := types.ParseNumMMOrdersKey(iter.Key())
		numMMOrders := utils.BigEndianToUint32(iter.Value())
		if cb(ordererAddr, marketId, numMMOrders) {
			break
		}
	}
}

func (k Keeper) DeleteNumMMOrders(ctx sdk.Context, ordererAddr sdk.AccAddress, marketId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetNumMMOrdersKey(ordererAddr, marketId))
}
