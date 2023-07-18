package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit, qtyLimit, quoteLimit *sdk.Dec, halveFees, simulate bool) (res types.ExecuteOrderResult, err error) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	payDenom, receiveDenom := market.PayReceiveDenoms(isBuy)
	res = types.NewExecuteOrderResult(payDenom, receiveDenom)
	var lastPrice sdk.Dec
	escrow := types.NewEscrow(market.MustGetEscrowAddress())
	obs := k.ConstructTempOrderBookSide(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit, 0, escrow)
	for _, level := range obs.Levels {
		if priceLimit != nil &&
			((isBuy && level.Price.GT(*priceLimit)) ||
				(!isBuy && level.Price.LT(*priceLimit))) {
			break
		}
		if qtyLimit != nil && !qtyLimit.Sub(res.ExecutedQuantity).IsPositive() {
			break
		}
		if quoteLimit != nil && !quoteLimit.Sub(res.ExecutedQuote).IsPositive() {
			break
		}

		executableQty := types.TotalExecutableQuantity(level.Orders, level.Price)
		var remainingQty sdk.Dec
		if qtyLimit != nil {
			remainingQty = qtyLimit.Sub(res.ExecutedQuantity)
		}
		if quoteLimit != nil {
			qty := quoteLimit.Sub(res.ExecutedQuote).QuoTruncate(level.Price)
			if remainingQty.IsNil() {
				remainingQty = qty
			} else {
				remainingQty = sdk.MinDec(remainingQty, qty)
			}
		}
		execQty := executableQty
		if !remainingQty.IsNil() {
			execQty = sdk.MinDec(execQty, remainingQty)
			if execQty.Equal(remainingQty) {
				res.FullyExecuted = true
			}
		}

		market.FillTempOrderBookLevel(level, execQty, level.Price, true, halveFees)
		execQuote := types.QuoteAmount(isBuy, level.Price, execQty)
		var paid, received, fee sdk.DecCoin
		if isBuy {
			paid = sdk.NewDecCoinFromDec(market.QuoteDenom, execQuote)
			deductedQty, feeQty := market.DeductTakerFee(execQty, halveFees)
			received = sdk.NewDecCoinFromDec(market.BaseDenom, deductedQty)
			fee = sdk.NewDecCoinFromDec(market.BaseDenom, feeQty)
		} else {
			paid = sdk.NewDecCoinFromDec(market.BaseDenom, execQty)
			deductedQuote, feeQuote := market.DeductTakerFee(execQuote, halveFees)
			received = sdk.NewDecCoinFromDec(market.QuoteDenom, deductedQuote)
			fee = sdk.NewDecCoinFromDec(market.QuoteDenom, feeQuote)
		}
		if !simulate {
			escrow.Lock(ordererAddr, paid)
			escrow.Unlock(ordererAddr, received)
		}
		res.ExecutedQuantity = res.ExecutedQuantity.Add(execQty)
		res.ExecutedQuote = res.ExecutedQuote.Add(execQuote)
		res.Paid = res.Paid.Add(paid)
		res.Received = res.Received.Add(received)
		res.Fee = res.Fee.Add(fee)
		lastPrice = level.Price
	}
	res.Paid = sdk.NewDecCoinFromDec(res.Paid.Denom, res.Paid.Amount.Ceil())
	res.Received = sdk.NewDecCoinFromDec(res.Received.Denom, res.Received.Amount.TruncateDec())
	// TODO: calculate fee correctly
	if !simulate {
		var tempOrders []*types.TempOrder
		for _, level := range obs.Levels {
			tempOrders = append(tempOrders, level.Orders...)
		}
		if err = k.FinalizeMatching(ctx, market, tempOrders, escrow); err != nil {
			return
		}
		if !lastPrice.IsNil() {
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &lastPrice
			state.LastMatchingHeight = ctx.BlockHeight()
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	return
}

func (k Keeper) ConstructTempOrderBookSide(
	ctx sdk.Context, market types.Market, isBuy bool,
	priceLimit, qtyLimit, quoteLimit *sdk.Dec, maxNumPriceLevels int,
	escrow *types.Escrow) *types.TempOrderBookSide {
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	// TODO: adjust price limit
	obs := types.NewTempOrderBookSide(isBuy)
	numPriceLevels := 0
	k.IterateOrderBookSideByMarket(ctx, market.Id, isBuy, false, func(order types.Order) (stop bool) {
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
		if maxNumPriceLevels > 0 {
			numPriceLevels++
			if numPriceLevels >= maxNumPriceLevels {
				return true
			}
		}
		return false
	})
	for _, name := range k.sourceNames {
		source := k.sources[name]
		source.GenerateOrders(ctx, market, func(ordererAddr sdk.AccAddress, price, qty sdk.Dec) error {
			// orders from OrderSource don't have id - priority among them will
			// be determined the source's name.
			deposit := types.DepositAmount(isBuy, price, qty)
			order := types.NewOrder(
				0, types.OrderTypeLimit, ordererAddr, market.Id, isBuy, price, qty,
				0, qty, deposit, ctx.BlockTime())
			if escrow != nil {
				escrow.Lock(ordererAddr, sdk.NewDecCoinFromDec(market.DepositDenom(isBuy), deposit))
			}
			obs.AddOrder(types.NewTempOrder(order, market, source))
			return nil
		}, types.GenerateOrdersOptions{
			IsBuy:             isBuy,
			PriceLimit:        priceLimit,
			QuantityLimit:     qtyLimit,
			QuoteLimit:        quoteLimit,
			MaxNumPriceLevels: maxNumPriceLevels,
		})
	}
	if maxNumPriceLevels > 0 {
		bound := len(obs.Levels)
		if maxNumPriceLevels < bound {
			bound = maxNumPriceLevels
		}
		obs.Levels = obs.Levels[:bound]
	}
	return obs
}

func (k Keeper) FinalizeMatching(ctx sdk.Context, market types.Market, orders []*types.TempOrder, escrow *types.Escrow) error {
	if escrow == nil {
		escrow = types.NewEscrow(market.MustGetEscrowAddress())
	}
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

		ordererAddr := order.MustGetOrdererAddress()
		if order.IsUpdated {
			escrow.Unlock(ordererAddr, order.Received...)
			if order.Source == nil {
				payDenom, receiveDenom := market.PayReceiveDenoms(order.IsBuy)
				paid := sdk.NewDecCoinFromDec(
					order.Paid.Denom, order.Paid.Amount.Sub(order.Received.AmountOf(payDenom)))
				received := sdk.NewDecCoinFromDec(receiveDenom, order.Received.AmountOf(receiveDenom))
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
					if err := k.cancelOrder(ctx, market, order.Order, true); err != nil {
						return err
					}
				} else {
					k.SetOrder(ctx, order.Order)
				}
			}
		}
		// Should refund deposit
		if order.Source != nil && order.RemainingDeposit.IsPositive() {
			escrow.Unlock(ordererAddr, sdk.NewDecCoinFromDec(market.DepositDenom(order.IsBuy), order.RemainingDeposit))
		}
	}
	if err := k.ExecuteSendCoins(ctx); err != nil {
		return err
	}
	if err := escrow.Transact(ctx, k.bankKeeper); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := resultsBySourceName[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			// TODO: pass grouped results
			source.AfterOrdersExecuted(ctx, market, results)
			totalExecQty := utils.ZeroDec
			orderers, m := types.GroupTempOrderResultsByOrderer(results) // TODO: bit redundant
			for _, orderer := range orderers {
				var (
					isBuy                  bool
					payDenom, receiveDenom string
					totalPaid              sdk.DecCoin
					totalReceived          sdk.DecCoins
				)
				for _, res := range m[orderer] {
					totalExecQty = totalExecQty.Add(res.ExecutedQuantity)
					if totalPaid.Amount.IsNil() {
						isBuy = res.IsBuy
						totalPaid = res.Paid
						totalReceived = res.Received
					} else {
						if res.IsBuy != isBuy { // sanity check
							panic("inconsistent isBuy")
						}
						totalPaid = totalPaid.Add(res.Paid)
						totalReceived = totalReceived.Add(res.Received...)
					}
				}
				payDenom, receiveDenom = market.PayReceiveDenoms(isBuy)
				paid := sdk.NewDecCoinFromDec(totalPaid.Denom, totalPaid.Amount.Sub(totalReceived.AmountOf(payDenom)))
				received := sdk.NewDecCoinFromDec(receiveDenom, totalReceived.AmountOf(receiveDenom))
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderSourceOrdersFilled{
					MarketId:         market.Id,
					SourceName:       sourceName,
					Orderer:          orderer,
					IsBuy:            isBuy,
					ExecutedQuantity: totalExecQty,
					Paid:             paid,
					Received:         received,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
