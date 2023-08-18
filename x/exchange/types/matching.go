package types

import (
	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type MatchingContext struct {
	baseDenom, quoteDenom      string
	makerFeeRate, takerFeeRate sdk.Dec
}

func NewMatchingContext(market Market, halveFees bool) *MatchingContext {
	makerFeeRate := market.MakerFeeRate
	takerFeeRate := market.TakerFeeRate
	if halveFees {
		makerFeeRate = makerFeeRate.QuoInt64(2)
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	return &MatchingContext{
		baseDenom:    market.BaseDenom,
		quoteDenom:   market.QuoteDenom,
		makerFeeRate: makerFeeRate,
		takerFeeRate: takerFeeRate,
	}
}

func (ctx *MatchingContext) feeRate(isMaker bool) sdk.Dec {
	if isMaker {
		return ctx.makerFeeRate
	}
	return ctx.takerFeeRate
}

func (ctx *MatchingContext) FillOrder(order *MemOrder, qty, price sdk.Dec, isMaker bool) {
	executableQty := order.ExecutableQuantity()
	if qty.GT(executableQty) { // sanity check
		panic("open quantity is less than quantity")
	}
	if order.isMaker != nil && isMaker != *order.isMaker { // sanity check
		panic("an order's isMaker must be consistent under one matching context")
	}
	_, pays, receives, fee := ctx.fillOrder(order.typ, order.isBuy, qty, price, isMaker)
	order.paid = order.paid.Add(pays)
	order.remainingDeposit = order.remainingDeposit.Sub(pays)
	order.received = order.received.Add(receives)
	order.fee = order.fee.Add(fee)
	order.executedQty = order.executedQty.Add(qty)
	order.isMatched = true
	order.isMaker = &isMaker
}

func (ctx *MatchingContext) fillOrder(orderType MemOrderType, isBuy bool, qty, price sdk.Dec, isMaker bool) (executedQuote, pays, receives, fee sdk.Dec) {
	executedQuote = QuoteAmount(isBuy, price, qty)
	if isBuy {
		pays = executedQuote
		receives = qty
	} else {
		pays = qty
		receives = executedQuote
	}
	negativeMakerFeeRate := ctx.makerFeeRate.IsNegative()
	if isMaker && negativeMakerFeeRate {
		fee = ctx.makerFeeRate.MulTruncate(pays) // is negative
		pays = pays.Add(fee)
	} else if orderType == UserMemOrder {
		receives, fee = DeductFee(receives, ctx.feeRate(isMaker))
	} else {
		fee = utils.ZeroDec
	}
	return
}

func (ctx *MatchingContext) FillOrders(orders []*MemOrder, qty, price sdk.Dec, isMaker bool) {
	totalExecutableQty := TotalExecutableQuantity(orders)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	if len(orders) == 1 { // there's only one order
		ctx.FillOrder(orders[0], qty, price, isMaker)
		return
	}
	// First, distribute quantity evenly.
	remainingQty := qty
	for _, order := range orders {
		executableQty := order.ExecutableQuantity()
		if executableQty.IsZero() { // sanity check
			panic("executable quantity is zero")
		}
		executedQty := executableQty.MulTruncate(qty).QuoTruncate(totalExecutableQty)
		if executedQty.IsPositive() {
			ctx.FillOrder(order, executedQty, price, isMaker)
			remainingQty = remainingQty.Sub(executedQty)
		}
	}
	// Then, distribute remaining quantity based on priority.
	if remainingQty.IsPositive() {
		slices.SortFunc(orders, func(a, b *MemOrder) bool {
			return a.HasPriorityOver(b)
		})
		for _, order := range orders {
			if remainingQty.IsZero() {
				break
			}
			executedQty := sdk.MinDec(remainingQty, order.ExecutableQuantity())
			if executedQty.IsPositive() {
				ctx.FillOrder(order, executedQty, price, isMaker)
				remainingQty = remainingQty.Sub(executedQty)
			}
		}
	}
}

func (ctx *MatchingContext) FillOrderBookPriceLevel(level *MemOrderBookPriceLevel, qty, price sdk.Dec, isMaker bool) {
	executableQty := TotalExecutableQuantity(level.orders)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	} else if executableQty.Equal(qty) { // full matches
		ctx.FillOrders(level.orders, qty, price, isMaker)
	} else {
		groups := GroupMemOrdersByMsgHeight(level.orders)
		totalExecQty := utils.ZeroDec
		for _, group := range groups {
			remainingQty := qty.Sub(totalExecQty)
			if remainingQty.IsZero() {
				break
			}
			// TODO: optimize duplicate TotalExecutableQuantity calls?
			executableQty = TotalExecutableQuantity(group.orders)
			executedQty := sdk.MinDec(remainingQty, executableQty)
			ctx.FillOrders(group.orders, executedQty, price, isMaker)
			totalExecQty = totalExecQty.Add(executedQty)
		}
	}
}

func (ctx *MatchingContext) MatchOrderBookPriceLevels(
	levelA *MemOrderBookPriceLevel, isLevelAMaker bool,
	levelB *MemOrderBookPriceLevel, isLevelBMaker bool, price sdk.Dec) (executedQty sdk.Dec, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.orders)
	executableQtyB := TotalExecutableQuantity(levelB.orders)
	executedQty = sdk.MinDec(executableQtyA, executableQtyB)
	fullA = executedQty.Equal(executableQtyA)
	fullB = executedQty.Equal(executableQtyB)
	ctx.FillOrderBookPriceLevel(levelA, executedQty, price, isLevelAMaker)
	ctx.FillOrderBookPriceLevel(levelB, executedQty, price, isLevelBMaker)
	return
}

// ExecuteOrder simulates an order execution and returns the result.
// obs should be a valid MemOrderBookSide which the order will be executed against.
// The order will always be a taker.
func (ctx *MatchingContext) ExecuteOrder(
	obs *MemOrderBookSide, qtyLimit, quoteLimit *sdk.Dec) (res ExecuteOrderResult) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	isBuy := !obs.isBuy // the order's direction
	res = NewExecuteOrderResult(PayReceiveDenoms(ctx.baseDenom, ctx.quoteDenom, isBuy))
	for _, level := range obs.levels {
		// Check limits
		if qtyLimit != nil && !qtyLimit.Sub(res.ExecutedQuantity).IsPositive() {
			break
		}
		if quoteLimit != nil && !quoteLimit.Sub(res.ExecutedQuote).IsPositive() {
			break
		}

		executableQty := TotalExecutableQuantity(level.orders)
		var remainingQty sdk.Dec
		if qtyLimit != nil {
			remainingQty = qtyLimit.Sub(res.ExecutedQuantity)
		}
		if quoteLimit != nil {
			remainingQuote := quoteLimit.Sub(res.ExecutedQuote)
			if remainingQuote.LT(utils.OneDec) {
				res.FullyExecuted = true
				break
			}
			qty := remainingQuote.QuoTruncate(level.price)
			if remainingQty.IsNil() {
				remainingQty = qty
			} else {
				remainingQty = sdk.MinDec(remainingQty, qty)
			}
		}
		if remainingQty.LT(utils.OneDec) {
			res.FullyExecuted = true
			break
		}

		executedQty := sdk.MinDec(executableQty, remainingQty)
		if executedQty.Equal(remainingQty) {
			res.FullyExecuted = true // used in swap
		}

		matchPrice := level.price
		ctx.FillOrderBookPriceLevel(level, executedQty, matchPrice, true)
		executedQuote, pays, receives, fee := ctx.fillOrder(UserMemOrder, isBuy, executedQty, matchPrice, false)
		res.ExecutedQuantity = res.ExecutedQuantity.Add(executedQty)
		res.ExecutedQuote = res.ExecutedQuote.Add(executedQuote)
		res.Paid.Amount = res.Paid.Amount.Add(pays)
		res.Received.Amount = res.Received.Amount.Add(receives)
		res.Fee.Amount = res.Fee.Amount.Add(fee)
		res.LastPrice = matchPrice
	}
	return
}

func (ctx *MatchingContext) RunSinglePriceAuction(buyObs, sellObs *MemOrderBookSide) (matchPrice sdk.Dec) {
	buyLevelIdx, sellLevelIdx := 0, 0
	var buyLastPrice, sellLastPrice sdk.Dec
	for buyLevelIdx < len(buyObs.levels) && sellLevelIdx < len(sellObs.levels) {
		buyLevel := buyObs.levels[buyLevelIdx]
		sellLevel := sellObs.levels[sellLevelIdx]
		if buyLevel.price.LT(sellLevel.price) {
			break
		}
		buyExecutableQty := TotalExecutableQuantity(buyLevel.Orders())
		sellExecutableQty := TotalExecutableQuantity(sellLevel.Orders())
		execQty := sdk.MinDec(buyExecutableQty, sellExecutableQty)
		buyLastPrice = buyLevel.price
		sellLastPrice = sellLevel.price
		buyFull := execQty.Equal(buyExecutableQty)
		sellFull := execQty.Equal(sellExecutableQty)
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
	}
	if !buyLastPrice.IsNil() && !sellLastPrice.IsNil() {
		matchPrice = RoundPrice(buyLastPrice.Add(sellLastPrice).QuoInt64(2))
		buyLevelIdx, sellLevelIdx = 0, 0
		for buyLevelIdx < len(buyObs.levels) && sellLevelIdx < len(sellObs.levels) {
			buyLevel := buyObs.levels[buyLevelIdx]
			sellLevel := sellObs.levels[sellLevelIdx]
			if buyLevel.price.LT(matchPrice) || sellLevel.price.GT(matchPrice) {
				break
			}
			// Both sides are taker
			_, sellFull, buyFull := ctx.MatchOrderBookPriceLevels(sellLevel, false, buyLevel, false, matchPrice)
			if buyFull {
				buyLevelIdx++
			}
			if sellFull {
				sellLevelIdx++
			}
		}
	}
	return
}

func (ctx *MatchingContext) BatchMatchOrderBookSides(buyObs, sellObs *MemOrderBookSide, lastPrice sdk.Dec) (newLastPrice sdk.Dec) {
	// Phase 1: Match orders with price below(or equal to) the last price and
	// price above(or equal to) the last price.
	// The execution price is the last price.
	buyLevelIdx, sellLevelIdx := 0, 0
	for buyLevelIdx < len(buyObs.levels) && sellLevelIdx < len(sellObs.levels) {
		buyLevel := buyObs.levels[buyLevelIdx]
		sellLevel := sellObs.levels[sellLevelIdx]
		if buyLevel.price.LT(lastPrice) || sellLevel.price.GT(lastPrice) {
			break
		}
		// Both sides are taker
		_, sellFull, buyFull := ctx.MatchOrderBookPriceLevels(sellLevel, false, buyLevel, false, lastPrice)
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
	}
	// If there's no more levels to match, return earlier.
	if buyLevelIdx >= len(buyObs.levels) || sellLevelIdx >= len(sellObs.levels) {
		return lastPrice
	}

	// Phase 2: Match orders in traditional exchange's manner.
	// The matching price is determined by the direction of price.

	// No sell orders with price below(or equal to) the last price,
	// thus the price will increase.
	isPriceIncreasing := sellObs.levels[sellLevelIdx].price.GT(lastPrice)
	for buyLevelIdx < len(buyObs.levels) && sellLevelIdx < len(sellObs.levels) {
		buyLevel := buyObs.levels[buyLevelIdx]
		sellLevel := sellObs.levels[sellLevelIdx]
		if buyLevel.price.LT(sellLevel.price) {
			break
		}
		var matchPrice sdk.Dec
		if isPriceIncreasing {
			matchPrice = sellLevel.price
		} else {
			matchPrice = buyLevel.price
		}
		_, sellFull, buyFull := ctx.MatchOrderBookPriceLevels(sellLevel, isPriceIncreasing, buyLevel, !isPriceIncreasing, matchPrice)
		newLastPrice = matchPrice
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
	}
	return newLastPrice
}

func PayReceiveDenoms(baseDenom, quoteDenom string, isBuy bool) (payDenom, receiveDenom string) {
	if isBuy {
		return quoteDenom, baseDenom
	}
	return baseDenom, quoteDenom
}
