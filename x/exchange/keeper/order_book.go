package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

//func (k Keeper) TransientOrderBook(ctx sdk.Context, marketId uint64, minPrice, maxPrice sdk.Dec) (ob types.OrderBook, err error) {
//	ctx, _ = ctx.CacheContext()
//	market, found := k.GetMarket(ctx, marketId)
//	if !found {
//		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
//		return
//	}
//	k.ConstructTempOrderBookSide(ctx, market, false, &maxPrice, nil, nil)
//	k.ConstructTempOrderBookSide(ctx, market, true, &minPrice, nil, nil)
//	makeCb := func(levels *[]types.OrderBookPriceLevel) func(order types.TransientOrder) (stop bool) {
//		return func(order types.TransientOrder) (stop bool) {
//			if len(*levels) > 0 {
//				lastLevel := (*levels)[len(*levels)-1]
//				if lastLevel.Price.Equal(order.Order.Price) {
//					lastLevel.Quantity = lastLevel.Quantity.Add(order.Order.OpenQuantity)
//					(*levels)[len(*levels)-1] = lastLevel
//					return false
//				}
//			}
//			*levels = append(*levels, types.OrderBookPriceLevel{
//				Price:    order.Order.Price,
//				Quantity: order.Order.OpenQuantity,
//			})
//			return false
//		}
//	}
//	k.IterateTransientOrderBookSide(ctx, marketId, false, makeCb(&ob.Sells))
//	k.IterateTransientOrderBookSide(ctx, marketId, true, makeCb(&ob.Buys))
//	return ob, nil
//}

func (k Keeper) ConstructTempOrderBookSide(
	ctx sdk.Context, market types.Market, isBuy bool,
	priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int) *types.TempOrderBookSide {
	accQty := utils.ZeroInt
	accQuote := utils.ZeroInt
	// TODO: adjust price limit
	// TODO: optimize gas for transient store
	obs := types.NewTempOrderBookSide(isBuy)
	k.IterateOrderBookSide(ctx, market.Id, isBuy, func(order types.Order) (stop bool) {
		if priceLimit != nil &&
			((isBuy && order.Price.LT(*priceLimit)) ||
				(!isBuy && order.Price.GT(*priceLimit))) {
			return true
		}
		if qtyLimit != nil && !qtyLimit.Sub(accQty).IsPositive() {
			return true
		}
		if quoteLimit != nil && !quoteLimit.Sub(accQuote).IsPositive() {
			return true
		}
		obs.AddOrder(market.NewTempOrder(order, nil))
		accQty = accQty.Add(order.OpenQuantity)
		accQuote = accQuote.Add(types.QuoteAmount(!isBuy, order.Price, order.OpenQuantity))
		return false
	})
	for _, name := range k.sourceNames {
		source := k.sources[name]
		source.GenerateOrders(ctx, market, func(ordererAddr sdk.AccAddress, price sdk.Dec, qty sdk.Int) error {
			order, err := k.CreateOrder(ctx, market, ordererAddr, isBuy, price, qty, qty, true)
			if err != nil {
				return err
			}
			obs.AddOrder(market.NewTempOrder(order, source))
			return nil
		}, types.GenerateOrdersOptions{
			IsBuy:         isBuy,
			PriceLimit:    priceLimit,
			QuantityLimit: qtyLimit,
			QuoteLimit:    quoteLimit,
		})
	}
	return obs
}

func (k Keeper) ApplyTempOrderBookSideChanges(ctx sdk.Context, market types.Market, obs *types.TempOrderBookSide) error {
	for _, level := range obs.Levels {
		for _, order := range level.Orders {
			ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
			if order.IsUpdated {
				if err := k.ReleaseCoins(ctx, market, ordererAddr, order.Received, true); err != nil {
					return err
				}
				if order.Source == nil {
					// Update user orders
					if order.ExecutableQuantity(order.Price).IsZero() {
						k.DeleteOrder(ctx, order.Order)
						k.DeleteOrderBookOrder(ctx, order.Order)
					} else {
						k.SetOrder(ctx, order.Order)
					}
				}
			}
			// Should refund deposit
			if order.Source != nil || (order.IsUpdated && order.ExecutableQuantity(order.Price).IsZero()) {
				if order.RemainingDeposit.IsPositive() {
					if err := k.ReleaseCoin(
						ctx, market, ordererAddr,
						market.DepositCoin(order.IsBuy, order.RemainingDeposit), true); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
