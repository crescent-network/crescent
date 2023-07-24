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
	mCtx := types.NewMatchingContext(market, halveFees)
	obs := k.ConstructMemOrderBookSide(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit, 0, escrow)
	for _, level := range obs.Levels() {
		if priceLimit != nil &&
			((isBuy && level.Price().GT(*priceLimit)) ||
				(!isBuy && level.Price().LT(*priceLimit))) {
			break
		}
		if qtyLimit != nil && !qtyLimit.Sub(res.ExecutedQuantity).IsPositive() {
			break
		}
		if quoteLimit != nil && !quoteLimit.Sub(res.ExecutedQuote).IsPositive() {
			break
		}

		executableQty := types.TotalExecutableQuantity(level.Orders())
		var remainingQty sdk.Dec
		if qtyLimit != nil {
			remainingQty = qtyLimit.Sub(res.ExecutedQuantity)
		}
		if quoteLimit != nil {
			qty := quoteLimit.Sub(res.ExecutedQuote).QuoTruncate(level.Price())
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

		mCtx.FillMemOrderBookPriceLevel(level, execQty, level.Price(), true)
		execQuote := types.QuoteAmount(isBuy, level.Price(), execQty)
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
		lastPrice = level.Price()
	}
	res.Paid = sdk.NewDecCoinFromDec(res.Paid.Denom, res.Paid.Amount.Ceil())
	res.Received = sdk.NewDecCoinFromDec(res.Received.Denom, res.Received.Amount.TruncateDec())
	// TODO: calculate fee correctly
	if !simulate {
		if err = k.FinalizeMatching(ctx, market, obs.Orders(), escrow); err != nil {
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

func (k Keeper) ConstructMemOrderBookSide(
	ctx sdk.Context, market types.Market, isBuy bool,
	priceLimit, qtyLimit, quoteLimit *sdk.Dec, maxNumPriceLevels int,
	escrow *types.Escrow) *types.MemOrderBookSide {
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	obs := types.NewMemOrderBookSide(isBuy)
	numPriceLevels := 0
	k.IterateOrderBookSide(ctx, market.Id, isBuy, priceLimit, func(price sdk.Dec, orders []types.Order) (stop bool) {
		if qtyLimit != nil && !qtyLimit.Sub(accQty).IsPositive() {
			return true
		}
		if quoteLimit != nil && !quoteLimit.Sub(accQuote).IsPositive() {
			return true
		}
		for _, order := range orders {
			obs.AddOrder(types.NewUserMemOrder(order))
			accQty = accQty.Add(order.OpenQuantity)
			accQuote = accQuote.Add(types.QuoteAmount(!isBuy, order.Price, order.OpenQuantity))
		}
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
			if escrow != nil {
				escrow.Lock(ordererAddr, sdk.NewDecCoinFromDec(market.DepositDenom(isBuy), deposit))
			}
			obs.AddOrder(types.NewOrderSourceMemOrder(ordererAddr, isBuy, price, qty, source))
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
		obs.Limit(maxNumPriceLevels)
	}
	return obs
}

func (k Keeper) FinalizeMatching(ctx sdk.Context, market types.Market, orders []*types.MemOrder, escrow *types.Escrow) error {
	if escrow == nil {
		escrow = types.NewEscrow(market.MustGetEscrowAddress())
	}
	var sourceNames []string
	ordersBySource := map[string][]*types.MemOrder{}
	for _, memOrder := range orders {
		if memOrder.IsMatched() && memOrder.Type() == types.OrderSourceMemOrder {
			sourceName := memOrder.Source().Name()
			results, ok := ordersBySource[sourceName]
			if !ok {
				sourceNames = append(sourceNames, sourceName)
			}
			ordersBySource[sourceName] = append(results, memOrder)
		}

		ordererAddr := memOrder.OrdererAddress()
		if memOrder.IsMatched() {
			payDenom, receiveDenom := market.PayReceiveDenoms(memOrder.IsBuy())
			received := sdk.NewDecCoinFromDec(receiveDenom, memOrder.Received())
			if memOrder.Fee().IsPositive() {
				received.Amount = received.Amount.Sub(memOrder.Fee())
			}
			escrow.Unlock(ordererAddr, received)
			if memOrder.Fee().IsNegative() {
				escrow.Unlock(ordererAddr, sdk.NewDecCoinFromDec(payDenom, memOrder.Fee().Neg()))
			}
			if memOrder.Type() == types.UserMemOrder {
				order := memOrder.Order()
				order.OpenQuantity = order.OpenQuantity.Sub(memOrder.ExecutedQuantity())
				order.RemainingDeposit = order.RemainingDeposit.Sub(memOrder.Paid())
				paid := sdk.NewDecCoinFromDec(payDenom, memOrder.Paid())
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderFilled{
					MarketId:         market.Id,
					OrderId:          order.Id,
					Orderer:          order.Orderer,
					IsBuy:            order.IsBuy,
					Price:            order.Price,
					Quantity:         order.Quantity,
					OpenQuantity:     order.OpenQuantity,
					ExecutedQuantity: memOrder.ExecutedQuantity(),
					Paid:             paid,
					Received:         received,
				}); err != nil {
					return err
				}
				// Update user orders
				executableQty := order.ExecutableQuantity()
				if order.IsBuy && executableQty.TruncateDec().IsZero() ||
					!order.IsBuy && executableQty.MulTruncate(order.Price).TruncateDec().IsZero() {
					if err := k.cancelOrder(ctx, market, order); err != nil {
						return err
					}
				} else {
					k.SetOrder(ctx, order)
				}
			}
		}
		// Should refund deposit
		if memOrder.Type() == types.OrderSourceMemOrder && memOrder.RemainingDeposit().IsPositive() {
			escrow.Unlock(
				ordererAddr,
				sdk.NewDecCoinFromDec(market.DepositDenom(memOrder.IsBuy()), memOrder.RemainingDeposit()))
		}
	}
	if err := escrow.Transact(ctx, k.bankKeeper); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := ordersBySource[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			// TODO: pass grouped results
			totalExecQty := utils.ZeroDec
			ordererAddrs, m := types.GroupMemOrdersByOrderer(results) // TODO: bit redundant
			for _, ordererAddr := range ordererAddrs {
				source.AfterOrdersExecuted(ctx, market, ordererAddr, results)
				var (
					isBuy                  bool
					payDenom, receiveDenom string
					totalPaid              sdk.Dec
					totalReceived          sdk.Dec
					totalFee               sdk.Dec
				)
				for _, order := range m[ordererAddr.String()] {
					totalExecQty = totalExecQty.Add(order.ExecutedQuantity())
					if totalPaid.IsNil() {
						isBuy = order.IsBuy()
						totalPaid = order.Paid()
						totalReceived = order.Received()
						totalFee = order.Fee()
					} else {
						if order.IsBuy() != isBuy { // sanity check
							panic("inconsistent isBuy")
						}
						totalPaid = totalPaid.Add(order.Paid())
						totalReceived = totalReceived.Add(order.Received())
						totalFee = totalFee.Add(order.Fee())
					}
				}
				payDenom, receiveDenom = market.PayReceiveDenoms(isBuy)
				paid := sdk.NewDecCoinFromDec(payDenom, totalPaid)
				if totalFee.IsNegative() {
					paid.Amount = paid.Amount.Add(totalFee)
				}
				received := sdk.NewDecCoinFromDec(receiveDenom, totalReceived)
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderSourceOrdersFilled{
					MarketId:         market.Id,
					SourceName:       sourceName,
					Orderer:          ordererAddr.String(),
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
