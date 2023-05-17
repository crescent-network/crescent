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

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int, simulate bool) (totalExecQty, totalExecQuote sdk.Int) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	var lastPrice sdk.Dec
	totalExecQty = utils.ZeroInt
	totalExecQuote = utils.ZeroInt
	obs := k.ConstructTempOrderBookSide(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit)
	for _, level := range obs.Levels {
		if priceLimit != nil &&
			((isBuy && level.Price.GT(*priceLimit)) ||
				(!isBuy && level.Price.LT(*priceLimit))) {
			break
		}
		if qtyLimit != nil && !qtyLimit.Sub(totalExecQty).IsPositive() {
			break
		}
		if quoteLimit != nil && !quoteLimit.Sub(totalExecQuote).IsPositive() {
			break
		}

		executableQty := types.TotalExecutableQuantity(level.Orders, level.Price)
		execQty := executableQty
		if qtyLimit != nil {
			execQty = utils.MinInt(execQty, qtyLimit.Sub(totalExecQty))
		}
		if quoteLimit != nil {
			execQty = utils.MinInt(
				execQty,
				quoteLimit.Sub(totalExecQuote).ToDec().QuoTruncate(level.Price).TruncateInt())
		}

		market.FillTempOrderBookLevel(level, execQty, level.Price, true)
		execQuote := types.QuoteAmount(isBuy, level.Price, execQty)
		// TODO: refactor code
		if isBuy {
			if err := k.EscrowCoin(ctx, market, ordererAddr, sdk.NewCoin(market.QuoteDenom, execQuote), true); err != nil {
				panic(err)
			}
			receive := utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQty).TruncateInt()
			if err := k.ReleaseCoin(ctx, market, ordererAddr, sdk.NewCoin(market.BaseDenom, receive), true); err != nil {
				panic(err)
			}
		} else {
			if err := k.EscrowCoin(ctx, market, ordererAddr, sdk.NewCoin(market.BaseDenom, execQty), true); err != nil {
				panic(err)
			}
			receive := utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQuote).TruncateInt()
			if err := k.ReleaseCoin(ctx, market, ordererAddr, sdk.NewCoin(market.QuoteDenom, receive), true); err != nil {
				panic(err)
			}
		}
		totalExecQty = totalExecQty.Add(execQty)
		totalExecQuote = totalExecQuote.Add(execQuote)
		lastPrice = level.Price
	}
	if !simulate {
		var tempOrders []*types.TempOrder
		for _, level := range obs.Levels {
			tempOrders = append(tempOrders, level.Orders...)
		}
		if err := k.FinalizeMatching(ctx, market, tempOrders); err != nil {
			panic(err)
		}
		if !lastPrice.IsNil() {
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &lastPrice
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	return
}

func (k Keeper) FinalizeMatching(ctx sdk.Context, market types.Market, orders []*types.TempOrder) error {
	var sourceNames []string
	resultsBySourceName := map[string][]types.TempOrder{}
	for _, order := range orders {
		if order.IsUpdated && order.Source != nil {
			sourceName := order.Source.Name()
			results, ok := resultsBySourceName[sourceName]
			if !ok {
				sourceNames = append(sourceNames, sourceName)
			}
			resultsBySourceName[sourceName] = append(results, *order)
		}

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
	if err := k.ExecuteSendCoins(ctx); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := resultsBySourceName[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			source.AfterOrdersExecuted(ctx, market, results)
		}
	}
	return nil
}

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
		obs.AddOrder(types.NewTempOrder(order, market, nil))
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
			obs.AddOrder(types.NewTempOrder(order, market, source))
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

func (k Keeper) ApplyTempOrderChanges(ctx sdk.Context, market types.Market, orders []*types.TempOrder) error {
	for _, order := range orders {
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
	return nil
}
