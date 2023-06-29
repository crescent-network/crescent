package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "can-cancel-order", CanCancelOrderInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-state", OrderStateInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-book", OrderBookInvariant(k))
	ir.RegisterRoute(types.ModuleName, "order-book-order", OrderBookOrderInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		res, broken = CanCancelOrderInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = OrderStateInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = OrderBookInvariant(k)(ctx)
		if broken {
			return
		}
		return OrderBookOrderInvariant(k)(ctx)
	}
}

func CanCancelOrderInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		ctx, _ = ctx.CacheContext()
		msg := ""
		cnt := 0
		k.IterateAllOrders(ctx, func(order types.Order) (stop bool) {
			_, _, err := k.CancelOrder(ctx, order.MustGetOrdererAddress(), order.Id)
			if err != nil {
				msg += fmt.Sprintf("\tcannot cancel order %d: %v\n", order.Id, err)
				cnt++
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "can cancel order",
			fmt.Sprintf(
				"found %d order(s) that cannot be cancelled\n%s",
				cnt, msg)), broken
	}
}

func OrderStateInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		k.IterateAllOrders(ctx, func(order types.Order) (stop bool) {
			if !order.Deadline.After(ctx.BlockTime()) {
				msg += fmt.Sprintf("\torder %d should have been expired at %s\n", order.Id, order.Deadline)
				cnt++
			}
			if order.OpenQuantity.IsZero() {
				msg += fmt.Sprintf("\torder %d should have been deleted since it's fulfilled\n", order.Id)
				cnt++
			}
			if order.RemainingDeposit.IsZero() {
				msg += fmt.Sprintf("\torder %d should have been deleted since it has no remaining deposit\n", order.Id)
				cnt++
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "order state",
			fmt.Sprintf("found %d wrong order state(s)\n%s", cnt, msg)), broken
	}
}

func OrderBookInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
			var bestBuyPrice, bestSellPrice sdk.Dec
			k.IterateOrderBookSide(ctx, market.Id, false, false, func(order types.Order) (stop bool) {
				bestSellPrice = order.Price
				return true
			})
			k.IterateOrderBookSide(ctx, market.Id, true, false, func(order types.Order) (stop bool) {
				bestBuyPrice = order.Price
				return true
			})
			if !bestSellPrice.IsNil() && !bestBuyPrice.IsNil() {
				if bestSellPrice.LTE(bestBuyPrice) {
					msg += fmt.Sprintf(
						"\tmarket %d has crossed order book: sell price %s <= buy price %s\n",
						market.Id, bestSellPrice, bestBuyPrice)
					cnt++
				}
			}
			marketState := k.MustGetMarketState(ctx, market.Id)
			if marketState.LastPrice != nil {
				if !bestSellPrice.IsNil() && bestSellPrice.LT(*marketState.LastPrice) {
					msg += fmt.Sprintf(
						"\tmarket %d has sell order under the last price: %s < %s\n",
						market.Id, bestSellPrice, marketState.LastPrice)
					cnt++
				}
				if !bestBuyPrice.IsNil() && bestBuyPrice.GT(*marketState.LastPrice) {
					msg += fmt.Sprintf(
						"\tmarket %d has buy order above the last price: %s > %s\n",
						market.Id, bestBuyPrice, marketState.LastPrice)
					cnt++
				}
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "order book",
			fmt.Sprintf("found %d wrong order book state(s)\n%s", cnt, msg)), broken
	}
}

func OrderBookOrderInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		k.IterateAllOrders(ctx, func(order types.Order) (stop bool) {
			if found := k.LookupOrderBookOrder(ctx, order.MarketId, order.IsBuy, order.Price, order.Id); !found {
				msg += fmt.Sprintf("\torder %d not found in order book\n", order.Id)
				cnt++
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "order book order",
			fmt.Sprintf("found %d order(s) that are not found in order book\n%s", cnt, msg)), broken
	}
}
