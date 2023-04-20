package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreateTransientSpotOrder(
	ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, isTemporary bool) error {
	deposit := types.DepositAmount(isBuy, price, qty)
	if err := k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit)); err != nil {
		return err
	}
	orderId := k.GetNextOrderIdWithUpdate(ctx)
	order := types.NewTransientSpotOrder(
		orderId, ordererAddr, market.Id, isBuy, price, qty, qty, deposit, isTemporary)
	k.SetTransientSpotOrderBookOrder(ctx, order)
	return nil
}

func (k Keeper) TransientOrderBook(ctx sdk.Context, marketId string, minPrice, maxPrice sdk.Dec) (ob types.OrderBook, err error) {
	ctx, _ = ctx.CacheContext()
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}
	// TODO: do not use hardcoded quantity
	k.constructTransientSpotOrderBook(ctx, market, false, &maxPrice, sdk.NewIntWithDecimal(1, 40))
	k.constructTransientSpotOrderBook(ctx, market, true, &minPrice, sdk.NewIntWithDecimal(1, 40))
	makeCb := func(levels *[]types.OrderBookPriceLevel) func(order types.TransientSpotOrder) (stop bool) {
		return func(order types.TransientSpotOrder) (stop bool) {
			if len(*levels) > 0 {
				lastLevel := (*levels)[len(*levels)-1]
				if lastLevel.Price.Equal(order.Order.Price) {
					lastLevel.Quantity = lastLevel.Quantity.Add(order.Order.OpenQuantity)
					(*levels)[len(*levels)-1] = lastLevel
					return false
				}
			}
			*levels = append(*levels, types.OrderBookPriceLevel{
				Price:    order.Order.Price,
				Quantity: order.Order.OpenQuantity,
			})
			return false
		}
	}
	k.IterateTransientSpotOrderBookSide(ctx, marketId, false, makeCb(&ob.Sells))
	k.IterateTransientSpotOrderBookSide(ctx, marketId, true, makeCb(&ob.Buys))
	return ob, nil
}

func (k Keeper) constructTransientSpotOrderBook(ctx sdk.Context, market types.SpotMarket, isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) {
	remainingQty := qty
	k.IterateSpotOrderBookSide(ctx, market.Id, isBuy, func(order types.SpotOrder) (stop bool) {
		// If the order's price exceeds the price limit, break
		if priceLimit != nil &&
			((isBuy && order.Price.LT(*priceLimit)) ||
				(!isBuy && order.Price.GT(*priceLimit))) {
			return true
		}
		k.SetTransientSpotOrderBookOrder(ctx, types.NewTransientSpotOrderFromSpotOrder(order))
		remainingQty = remainingQty.Sub(order.OpenQuantity)
		return !remainingQty.IsPositive()
	})
	for _, source := range k.spotOrderSources {
		source.RequestTransientSpotOrders(ctx, market, isBuy, priceLimit, qty)
	}
}

func (k Keeper) settleTransientSpotOrderBook(ctx sdk.Context, market types.SpotMarket) {
	k.IterateTransientSpotOrderBook(ctx, market.Id, func(order types.TransientSpotOrder) (stop bool) {
		// Should refund deposit
		if order.IsTemporary || (order.Updated && order.Order.OpenQuantity.IsZero()) {
			if err := k.ReleaseCoin(
				ctx, market, sdk.MustAccAddressFromBech32(order.Order.Orderer),
				market.DepositCoin(order.Order.IsBuy, order.Order.RemainingDeposit)); err != nil {
				panic(err)
			}
		}
		if !order.IsTemporary && order.Updated {
			if order.Order.OpenQuantity.IsZero() {
				k.DeleteSpotOrder(ctx, order.Order)
				k.DeleteSpotOrderBookOrder(ctx, order.Order)
			} else {
				k.SetSpotOrder(ctx, order.Order)
			}
		}
		k.DeleteTransientSpotOrderBookOrder(ctx, order)
		return false
	})
}
