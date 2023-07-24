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

func (ctx *MatchingContext) FillMemOrder(order *MemOrder, qty, price sdk.Dec, isMaker bool) {
	executableQty := order.ExecutableQuantity()
	if qty.GT(executableQty) { // sanity check
		panic("open quantity is less than quantity")
	}
	if order.isMaker != nil && isMaker != *order.isMaker { // sanity check
		panic("an order's isMaker must be consistent under one matching context")
	}
	var pays, receives, fee sdk.Dec
	if order.isBuy {
		pays = QuoteAmount(true, price, qty)
		receives = qty
	} else {
		pays = qty
		receives = QuoteAmount(false, price, qty)
	}
	negativeMakerFeeRate := ctx.makerFeeRate.IsNegative()
	if isMaker && negativeMakerFeeRate {
		fee = ctx.makerFeeRate.MulTruncate(pays) // is negative
	} else if order.typ == UserMemOrder {
		var receivesAfterFee sdk.Dec
		if isMaker {
			receivesAfterFee = utils.OneDec.Sub(ctx.makerFeeRate).MulTruncate(receives)
		} else {
			receivesAfterFee = utils.OneDec.Sub(ctx.takerFeeRate).MulTruncate(receives)
		}
		fee = receives.Sub(receivesAfterFee)
	}
	order.paid = order.paid.Add(pays)
	order.remainingDeposit = order.remainingDeposit.Sub(pays)
	order.received = order.received.Add(receives)
	order.fee = order.fee.Add(fee)
	order.executedQty = order.executedQty.Add(qty)
	order.isMatched = true
	order.isMaker = &isMaker
}

func (ctx *MatchingContext) FillMemOrders(orders []*MemOrder, qty, price sdk.Dec, isMaker bool) {
	totalExecutableQty := TotalExecutableQuantity(orders)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
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
			ctx.FillMemOrder(order, executedQty, price, isMaker)
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
				ctx.FillMemOrder(order, executedQty, price, isMaker)
				remainingQty = remainingQty.Sub(executedQty)
			}
		}
	}
}

func (ctx *MatchingContext) FillMemOrderBookPriceLevel(level *MemOrderBookPriceLevel, qty, price sdk.Dec, isMaker bool) {
	executableQty := TotalExecutableQuantity(level.orders)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	} else if executableQty.Equal(qty) { // full matches
		ctx.FillMemOrders(level.orders, qty, price, isMaker)
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
			ctx.FillMemOrders(group.orders, executedQty, price, isMaker)
			totalExecQty = totalExecQty.Add(executedQty)
		}
	}
}

func (ctx *MatchingContext) MatchMemOrderBookPriceLevels(
	levelA *MemOrderBookPriceLevel, isLevelAMaker bool,
	levelB *MemOrderBookPriceLevel, isLevelBMaker bool, price sdk.Dec) (executedQty sdk.Dec, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.orders)
	executableQtyB := TotalExecutableQuantity(levelB.orders)
	executedQty = sdk.MinDec(executableQtyA, executableQtyB)
	fullA = executedQty.Equal(executableQtyA)
	fullB = executedQty.Equal(executableQtyB)
	ctx.FillMemOrderBookPriceLevel(levelA, executedQty, price, isLevelAMaker)
	ctx.FillMemOrderBookPriceLevel(levelB, executedQty, price, isLevelBMaker)
	return
}
