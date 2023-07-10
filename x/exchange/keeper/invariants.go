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
			if order.MsgHeight == ctx.BlockHeight() {
				return false
			}
			if _, err := k.CancelOrder(ctx, order.MustGetOrdererAddress(), order.Id); err != nil {
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
			if order.ExecutableQuantity(order.Price).IsZero() {
				msg += fmt.Sprintf("\torder %d should have been deleted since it has no executable quantity\n", order.Id)
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
			var (
				bestBuyOrder, bestSellOrder           types.Order
				foundBestBuyOrder, foundBestSellOrder bool
			)
			k.IterateOrderBookSideByMarket(ctx, market.Id, false, false, func(order types.Order) (stop bool) {
				bestSellOrder = order
				foundBestSellOrder = true
				return true
			})
			k.IterateOrderBookSideByMarket(ctx, market.Id, true, false, func(order types.Order) (stop bool) {
				bestBuyOrder = order
				foundBestBuyOrder = true
				return true
			})
			if foundBestSellOrder && foundBestBuyOrder {
				if bestSellOrder.Price.LTE(bestBuyOrder.Price) {
					msg += fmt.Sprintf(
						"\tmarket %d has crossed order book: sell price %s <= buy price %s\n",
						market.Id, bestSellOrder.Price, bestBuyOrder.Price)
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
			if found := k.LookupOrderBookOrderIndex(ctx, order.MarketId, order.IsBuy, order.Price, order.Id); !found {
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
