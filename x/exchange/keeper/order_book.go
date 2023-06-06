package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int, halveFees, simulate bool) (totalExecQty sdk.Int, totalPaid, totalReceived, totalFee sdk.Coin) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	var lastPrice sdk.Dec
	totalExecQty = utils.ZeroInt
	totalExecQuote := utils.ZeroInt
	totalPaid = sdk.NewCoin(market.PayDenom(isBuy), utils.ZeroInt)
	totalReceived = sdk.NewCoin(market.ReceiveDenom(isBuy), utils.ZeroInt)
	totalFee = sdk.NewCoin(market.ReceiveDenom(isBuy), utils.ZeroInt)
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

		market.FillTempOrderBookLevel(level, execQty, level.Price, true, halveFees)
		execQuote := types.QuoteAmount(isBuy, level.Price, execQty)
		var paid, received, fee sdk.Coin
		if isBuy {
			paid = sdk.NewCoin(market.QuoteDenom, execQuote)
			deductedQty, feeQty := market.DeductTakerFee(execQty, halveFees)
			received = sdk.NewCoin(market.BaseDenom, deductedQty)
			fee = sdk.NewCoin(market.BaseDenom, feeQty)
		} else {
			paid = sdk.NewCoin(market.BaseDenom, execQty)
			deductedQuote, feeQuote := market.DeductTakerFee(execQuote, halveFees)
			received = sdk.NewCoin(market.QuoteDenom, deductedQuote)
			fee = sdk.NewCoin(market.QuoteDenom, feeQuote)
		}
		if err := k.EscrowCoin(ctx, market, ordererAddr, paid, true); err != nil {
			panic(err)
		}
		if err := k.ReleaseCoin(ctx, market, ordererAddr, received, true); err != nil {
			panic(err)
		}
		totalExecQty = totalExecQty.Add(execQty)
		totalExecQuote = totalExecQuote.Add(execQuote)
		totalPaid = totalPaid.Add(paid)
		totalReceived = totalReceived.Add(received)
		totalFee = totalFee.Add(fee)
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

func (k Keeper) ConstructTempOrderBookSide(
	ctx sdk.Context, market types.Market, isBuy bool,
	priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int) *types.TempOrderBookSide {
	accQty := utils.ZeroInt
	accQuote := utils.ZeroInt
	// TODO: adjust price limit
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
			// orders from OrderSource don't have id - priority among them will
			// be determined the source's name.
			order, err := k.newOrder(
				ctx, 0, types.OrderTypeLimit, market, ordererAddr, isBuy, price,
				qty, qty, ctx.BlockTime(), true)
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
				payDenom, receiveDenom := market.PayDenom(order.IsBuy), market.ReceiveDenom(order.IsBuy)
				paid := order.Paid.SubAmount(order.Received.AmountOf(payDenom))
				received := sdk.NewCoin(receiveDenom, order.Received.AmountOf(receiveDenom))
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderFilled{
					MarketId:         market.Id,
					OrderId:          order.Id,
					Orderer:          order.Orderer,
					IsBuy:            order.IsBuy,
					Price:            order.Price,
					Quantity:         order.Quantity,
					OpenQuantity:     order.OpenQuantity,
					ExecutedQuantity: order.ExecutedQuantity,
					Paid:             paid,
					Received:         received,
				}); err != nil {
					return err
				}
				// Update user orders
				if order.ExecutableQuantity(order.Price).IsZero() {
					if _, err := k.cancelOrder(ctx, market, order.Order, true); err != nil {
						return err
					}
				} else {
					k.SetOrder(ctx, order.Order)
				}
			}
		}
		// Should refund deposit
		if order.Source != nil && order.RemainingDeposit.IsPositive() {
			if err := k.ReleaseCoin(
				ctx, market, ordererAddr,
				market.DepositCoin(order.IsBuy, order.RemainingDeposit), true); err != nil {
				return err
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
