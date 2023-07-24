package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (market Market) FillMemOrder(order *MemOrder, qty, price sdk.Dec, isMaker, halveFees bool) {
	// TODO: refactor code
	if qty.GT(order.ExecutableQuantity(price)) { // sanity check
		panic("open quantity is less than quantity")
	}
	makerFeeRate, takerFeeRate := market.MakerFeeRate, market.TakerFeeRate
	if halveFees {
		makerFeeRate = makerFeeRate.QuoInt64(2)
		takerFeeRate = takerFeeRate.QuoInt64(2)
	}
	negativeMakerFeeRate := makerFeeRate.IsNegative()
	order.ExecutedQuantity = order.ExecutedQuantity.Add(qty)
	order.OpenQuantity = order.OpenQuantity.Sub(qty)
	if order.IsBuy {
		paid := QuoteAmount(true, price, qty)
		order.Paid.Amount = order.Paid.Amount.Add(paid)
		order.RemainingDeposit = order.RemainingDeposit.Sub(paid)
		if order.Source != nil || (isMaker && negativeMakerFeeRate) {
			order.Received = order.Received.Add(sdk.NewDecCoinFromDec(market.BaseDenom, qty))
		} else {
			if isMaker {
				order.Received = order.Received.Add(
					sdk.NewDecCoinFromDec(
						market.BaseDenom,
						utils.OneDec.Sub(makerFeeRate).MulTruncate(qty)))
			} else {
				order.Received = order.Received.Add(
					sdk.NewDecCoinFromDec(
						market.BaseDenom,
						utils.OneDec.Sub(takerFeeRate).MulTruncate(qty)))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.Received = order.Received.Add(
				sdk.NewDecCoinFromDec(
					market.QuoteDenom,
					makerFeeRate.Neg().MulTruncate(paid)))
		}
	} else {
		order.Paid.Amount = order.Paid.Amount.Add(qty)
		order.RemainingDeposit = order.RemainingDeposit.Sub(qty)
		quote := QuoteAmount(false, price, qty)
		if order.Source != nil || (isMaker && negativeMakerFeeRate) {
			order.Received = order.Received.Add(sdk.NewDecCoinFromDec(market.QuoteDenom, quote))
		} else {
			if isMaker {
				order.Received = order.Received.Add(
					sdk.NewDecCoinFromDec(
						market.QuoteDenom,
						utils.OneDec.Sub(makerFeeRate).MulTruncate(quote)))
			} else {
				order.Received = order.Received.Add(
					sdk.NewDecCoinFromDec(
						market.QuoteDenom,
						utils.OneDec.Sub(takerFeeRate).MulTruncate(quote)))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.Received = order.Received.Add(
				sdk.NewDecCoinFromDec(
					market.BaseDenom,
					makerFeeRate.Neg().MulTruncate(qty)))
		}
	}
	order.IsUpdated = true
}

func (market Market) FillMemOrders(orders []*MemOrder, qty, price sdk.Dec, isMaker, halveFees bool) {
	totalExecutableQty := TotalExecutableQuantity(orders, price)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	// First, distribute quantity evenly.
	remainingQty := qty
	for _, order := range orders {
		executableQty := order.ExecutableQuantity(price)
		if executableQty.IsZero() { // sanity check
			panic("executable quantity is zero")
		}
		executingQty := executableQty.MulTruncate(qty).QuoTruncate(totalExecutableQty)
		if executingQty.IsPositive() {
			market.FillMemOrder(order, executingQty, price, isMaker, halveFees)
			remainingQty = remainingQty.Sub(executingQty)
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
			executingQty := sdk.MinDec(remainingQty, order.ExecutableQuantity(price))
			if executingQty.IsPositive() {
				market.FillMemOrder(order, executingQty, price, isMaker, halveFees)
				remainingQty = remainingQty.Sub(executingQty)
			}
		}
	}
}

func (market Market) FillMemOrderBookPriceLevel(
	level *MemOrderBookPriceLevel, qty, price sdk.Dec, isMaker, halveFees bool) {
	executableQty := TotalExecutableQuantity(level.Orders, price)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	} else if executableQty.Equal(qty) { // full matches
		market.FillMemOrders(level.Orders, qty, price, isMaker, halveFees)
	} else {
		groups := GroupMemOrdersByMsgHeight(level.Orders)
		totalExecQty := utils.ZeroDec
		for _, group := range groups {
			remainingQty := qty.Sub(totalExecQty)
			if remainingQty.IsZero() {
				break
			}
			// TODO: optimize duplicate TotalExecutableQuantity calls?
			execQty := sdk.MinDec(remainingQty, TotalExecutableQuantity(group.Orders, price))
			market.FillMemOrders(group.Orders, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (market Market) MatchMemOrderBookPriceLevels(
	levelA *MemOrderBookPriceLevel, isLevelAMaker bool,
	levelB *MemOrderBookPriceLevel, isLevelBMaker bool, price sdk.Dec) (executedQty sdk.Dec, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.Orders, price)
	executableQtyB := TotalExecutableQuantity(levelB.Orders, price)
	executedQty = sdk.MinDec(executableQtyA, executableQtyB)
	fullA = executedQty.Equal(executableQtyA)
	fullB = executedQty.Equal(executableQtyB)
	market.FillMemOrderBookPriceLevel(levelA, executedQty, price, isLevelAMaker, false)
	market.FillMemOrderBookPriceLevel(levelB, executedQty, price, isLevelBMaker, false)
	return
}
