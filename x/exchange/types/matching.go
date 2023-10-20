package types

import (
	"fmt"

	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type MatchState struct {
	executedQty sdk.Int
	paid        sdk.Dec
	received    sdk.Dec
	feePaid     sdk.Dec
	feeReceived sdk.Dec
}

func NewMatchState() MatchState {
	return MatchState{
		executedQty: utils.ZeroInt,
		paid:        utils.ZeroDec,
		received:    utils.ZeroDec,
		feePaid:     utils.ZeroDec,
		feeReceived: utils.ZeroDec,
	}
}

func (ms *MatchState) IsMatched() bool {
	return ms.executedQty.IsPositive()
}

func (ms *MatchState) Fill(isBuy bool, qty sdk.Int, price sdk.Dec, feeRate sdk.Dec) (executedQuote, paid, received, feePaid, feeReceived sdk.Dec) {
	ms.executedQty = ms.executedQty.Add(qty)
	executedQuote = price.MulInt(qty)
	if isBuy {
		paid = executedQuote
		received = qty.ToDec()
	} else {
		paid = qty.ToDec()
		received = executedQuote
	}
	if feeRate.IsPositive() {
		feePaid = feeRate.Mul(received)
		received = received.Sub(feePaid)
	} else {
		feePaid = utils.ZeroDec
	}
	if feeRate.IsNegative() {
		feeReceived = feeRate.Neg().MulTruncate(paid)
		paid = paid.Sub(feeReceived)
	} else {
		feeReceived = utils.ZeroDec
	}
	ms.paid = ms.paid.Add(paid)
	ms.received = ms.received.Add(received)
	ms.feePaid = ms.feePaid.Add(feePaid)
	ms.feeReceived = ms.feeReceived.Add(feeReceived)
	return
}

func (ms *MatchState) Result() MatchResult {
	paid := ms.paid.Ceil().TruncateInt()
	feePaid := ms.feePaid.Ceil().TruncateInt()
	feeReceived := ms.paid.Add(ms.feeReceived).Ceil().TruncateInt().Sub(paid)
	received := ms.received.Add(ms.feePaid).TruncateInt().Sub(feePaid)
	if received.IsNegative() {
		received = ms.received.Add(ms.feePaid).Ceil().TruncateInt().Sub(feePaid)
	}
	return MatchResult{
		ExecutedQuantity: ms.executedQty,
		Paid:             paid,
		Received:         received,
		FeePaid:          feePaid,
		FeeReceived:      feeReceived,
	}
}

type MatchResult struct {
	ExecutedQuantity sdk.Int
	Paid             sdk.Int // Real paid amount(FeeReceived excluded)
	Received         sdk.Int // Real received amount(FeePaid excluded)
	FeePaid          sdk.Int
	FeeReceived      sdk.Int
}

func (res MatchResult) IsMatched() bool {
	return res.ExecutedQuantity.IsPositive()
}

type ExecuteOrderResult struct {
	LastPrice        sdk.Dec
	ExecutedQuantity sdk.Int
	Paid             sdk.Coin
	Received         sdk.Coin
	FeePaid          sdk.Coin
	FeeReceived      sdk.Coin
}

func (res ExecuteOrderResult) IsMatched() bool {
	return res.ExecutedQuantity.IsPositive()
}

type MatchingContext struct {
	baseDenom, quoteDenom string

	makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec
}

func NewMatchingContext(market Market, halveFees bool) *MatchingContext {
	makerFeeRate := market.Fees.MakerFeeRate
	takerFeeRate := market.Fees.TakerFeeRate
	if halveFees {
		makerFeeRate = makerFeeRate.QuoInt64(2)
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	return &MatchingContext{
		baseDenom:           market.BaseDenom,
		quoteDenom:          market.QuoteDenom,
		makerFeeRate:        makerFeeRate,
		takerFeeRate:        takerFeeRate,
		orderSourceFeeRatio: market.Fees.OrderSourceFeeRatio,
	}
}

func (ctx *MatchingContext) FeeRate(orderType MemOrderType, isMaker bool) (feeRate sdk.Dec) {
	if orderType == UserMemOrder {
		if isMaker {
			feeRate = ctx.makerFeeRate
		} else {
			feeRate = ctx.takerFeeRate
		}
	} else if orderType == OrderSourceMemOrder { // order source order
		if isMaker {
			feeRate = ctx.takerFeeRate.Neg().Mul(ctx.orderSourceFeeRatio)
		} else {
			feeRate = utils.ZeroDec
		}
	} else { // sanity check
		panic("invalid order type")
	}
	return feeRate
}

func (ctx *MatchingContext) FillOrder(order *MemOrder, qty sdk.Int, price sdk.Dec, isMaker bool) {
	executableQty := order.ExecutableQuantity()
	if qty.GT(executableQty) { // sanity check
		panic("open quantity is less than quantity")
	}
	if order.IsBuy && price.GT(order.Price) { // sanity check
		panic(fmt.Sprintf("matching price is higher than the order price: %s > %s", price, order.Price))
	} else if !order.IsBuy && price.LT(order.Price) { // sanity check
		panic(fmt.Sprintf("matching price is lower than the order price: %s < %s", price, order.Price))
	}
	feeRate := ctx.FeeRate(order.Type, isMaker)
	order.Fill(order.IsBuy, qty, price, feeRate)
}

func (ctx *MatchingContext) FillOrders(orders []*MemOrder, qty sdk.Int, price sdk.Dec, isMaker bool) {
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
		executedQty := executableQty.Mul(qty).Quo(totalExecutableQty)
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
			executedQty := utils.MinInt(remainingQty, order.ExecutableQuantity())
			if executedQty.IsPositive() {
				ctx.FillOrder(order, executedQty, price, isMaker)
				remainingQty = remainingQty.Sub(executedQty)
			}
		}
	}
}

func (ctx *MatchingContext) FillOrderBookPriceLevel(level *MemOrderBookPriceLevel, qty sdk.Int, price sdk.Dec, isMaker bool) {
	executableQty := TotalExecutableQuantity(level.Orders)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	} else if executableQty.Equal(qty) { // full matches
		ctx.FillOrders(level.Orders, qty, price, isMaker)
	} else {
		groups := GroupMemOrdersByMsgHeight(level.Orders)
		totalExecQty := utils.ZeroInt
		for _, group := range groups {
			remainingQty := qty.Sub(totalExecQty)
			if remainingQty.IsZero() {
				break
			}
			// TODO: optimize duplicate TotalExecutableQuantity calls?
			executableQty = TotalExecutableQuantity(group.orders)
			executedQty := utils.MinInt(remainingQty, executableQty)
			ctx.FillOrders(group.orders, executedQty, price, isMaker)
			totalExecQty = totalExecQty.Add(executedQty)
		}
	}
}

func (ctx *MatchingContext) MatchOrderBookPriceLevels(
	levelA *MemOrderBookPriceLevel, isLevelAMaker bool,
	levelB *MemOrderBookPriceLevel, isLevelBMaker bool, price sdk.Dec) (executedQty sdk.Int, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.Orders)
	executableQtyB := TotalExecutableQuantity(levelB.Orders)
	executedQty = utils.MinInt(executableQtyA, executableQtyB)
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
	isBuy bool, obs *MemOrderBookSide, qtyLimit, quoteLimit *sdk.Int) (res MatchResult, full bool, lastPrice sdk.Dec) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	if isBuy != !obs.IsBuy {
		panic(fmt.Sprintf("%v != %v", isBuy, !obs.IsBuy))
	}
	matchState := NewMatchState()
	totalExecutedQuote := utils.ZeroDec
	for _, level := range obs.Levels {
		// Check limits
		var (
			remainingQty   sdk.Int
			remainingQuote sdk.Dec
		)
		if qtyLimit != nil {
			remainingQty = qtyLimit.Sub(matchState.executedQty)
		}
		if quoteLimit != nil { // only when isBuy == true
			remainingQuote = quoteLimit.ToDec().Sub(totalExecutedQuote)
			qty := remainingQuote.QuoTruncate(level.Price).TruncateInt()
			if remainingQty.IsNil() {
				remainingQty = qty
			} else {
				remainingQty = utils.MinInt(remainingQty, qty)
			}
		}
		if remainingQty.IsZero() {
			full = true
			break
		}

		executableQty := TotalExecutableQuantity(level.Orders)
		executedQty := utils.MinInt(executableQty, remainingQty)
		if executedQty.Equal(remainingQty) {
			full = true
		}

		matchPrice := level.Price
		ctx.FillOrderBookPriceLevel(level, executedQty, matchPrice, true)
		executedQuote, _, _, _, _ := matchState.Fill(isBuy, executedQty, matchPrice, ctx.FeeRate(UserMemOrder, false))
		totalExecutedQuote = totalExecutedQuote.Add(executedQuote)
		lastPrice = matchPrice
	}
	return matchState.Result(), full, lastPrice
}

func (ctx *MatchingContext) RunSinglePriceAuction(buyObs, sellObs *MemOrderBookSide) (matchPrice sdk.Dec, matched bool) {
	buyLevelIdx, sellLevelIdx := 0, 0
	var buyLastPrice, sellLastPrice sdk.Dec
	buyExecutedQty, sellExecutedQty := utils.ZeroInt, utils.ZeroInt
	for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
		buyLevel := buyObs.Levels[buyLevelIdx]
		sellLevel := sellObs.Levels[sellLevelIdx]
		if buyLevel.Price.LT(sellLevel.Price) {
			break
		}
		buyExecutableQty := TotalExecutableQuantity(buyLevel.Orders).Sub(buyExecutedQty)
		sellExecutableQty := TotalExecutableQuantity(sellLevel.Orders).Sub(sellExecutedQty)
		executedQty := utils.MinInt(buyExecutableQty, sellExecutableQty)
		buyLastPrice = buyLevel.Price
		sellLastPrice = sellLevel.Price
		buyFull := executedQty.Equal(buyExecutableQty)
		sellFull := executedQty.Equal(sellExecutableQty)
		if buyFull {
			buyLevelIdx++
			buyExecutedQty = utils.ZeroInt
		} else {
			buyExecutedQty = buyExecutedQty.Add(executedQty)
		}
		if sellFull {
			sellLevelIdx++
			sellExecutedQty = utils.ZeroInt
		} else {
			sellExecutedQty = sellExecutedQty.Add(executedQty)
		}
	}
	if !buyLastPrice.IsNil() && !sellLastPrice.IsNil() {
		matchPrice = RoundPrice(buyLastPrice.Add(sellLastPrice).QuoInt64(2))
		buyLevelIdx, sellLevelIdx = 0, 0
		for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
			buyLevel := buyObs.Levels[buyLevelIdx]
			sellLevel := sellObs.Levels[sellLevelIdx]
			if buyLevel.Price.LT(matchPrice) || sellLevel.Price.GT(matchPrice) {
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
		matched = true
	}
	return
}

func (ctx *MatchingContext) BatchMatchOrderBookSides(buyObs, sellObs *MemOrderBookSide, lastPrice sdk.Dec) (newLastPrice sdk.Dec, matched bool) {
	newLastPrice = lastPrice // in case of there's no matching at all

	// Phase 1: Match orders with price below(or equal to) the last price and
	// price above(or equal to) the last price.
	// The execution price is the last price.
	buyLevelIdx, sellLevelIdx := 0, 0
	for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
		buyLevel := buyObs.Levels[buyLevelIdx]
		sellLevel := sellObs.Levels[sellLevelIdx]
		if buyLevel.Price.LT(lastPrice) || sellLevel.Price.GT(lastPrice) {
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
		matched = true
	}
	// If there's no more levels to match, return earlier.
	if (buyLevelIdx >= len(buyObs.Levels) || sellLevelIdx >= len(sellObs.Levels)) ||
		buyObs.Levels[buyLevelIdx].Price.LT(sellObs.Levels[sellLevelIdx].Price) {
		return lastPrice, matched
	}

	// Phase 2: Match orders in traditional exchange's manner.
	// The matching price is determined by the direction of price.

	// No sell orders with price below(or equal to) the last price,
	// thus the price will increase.
	isPriceIncreasing := sellObs.Levels[sellLevelIdx].Price.GT(lastPrice)
	for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
		buyLevel := buyObs.Levels[buyLevelIdx]
		sellLevel := sellObs.Levels[sellLevelIdx]
		if buyLevel.Price.LT(sellLevel.Price) {
			break
		}
		var matchPrice sdk.Dec
		if isPriceIncreasing {
			matchPrice = sellLevel.Price
		} else {
			matchPrice = buyLevel.Price
		}
		_, sellFull, buyFull := ctx.MatchOrderBookPriceLevels(sellLevel, isPriceIncreasing, buyLevel, !isPriceIncreasing, matchPrice)
		newLastPrice = matchPrice
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
		matched = true
	}
	return newLastPrice, matched
}

func PayReceiveDenoms(baseDenom, quoteDenom string, isBuy bool) (payDenom, receiveDenom string) {
	if isBuy {
		return quoteDenom, baseDenom
	}
	return baseDenom, quoteDenom
}
