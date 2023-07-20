package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (market Market) MatchOrderBookLevels(
	levelA *TempOrderBookLevel, isMakerA bool, levelB *TempOrderBookLevel, isMakerB bool, price sdk.Dec) (execQty sdk.Dec, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.Orders, price)
	executableQtyB := TotalExecutableQuantity(levelB.Orders, price)
	execQty = sdk.MinDec(executableQtyA, executableQtyB)
	fullA = execQty.Equal(executableQtyA)
	fullB = execQty.Equal(executableQtyB)
	market.FillTempOrderBookLevel(levelA, execQty, price, isMakerA, false)
	market.FillTempOrderBookLevel(levelB, execQty, price, isMakerB, false)
	return
}

func (market Market) FillTempOrderBookLevel(
	level *TempOrderBookLevel, qty, price sdk.Dec, isMaker, halveFees bool) {
	executableQty := TotalExecutableQuantity(level.Orders, price)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	} else if executableQty.Equal(qty) { // full matches
		market.FillTempOrders(level.Orders, qty, price, isMaker, halveFees)
	} else {
		groups := GroupTempOrdersByMsgHeight(level.Orders)
		totalExecQty := utils.ZeroDec
		for _, group := range groups {
			remainingQty := qty.Sub(totalExecQty)
			if remainingQty.IsZero() {
				break
			}
			// TODO: optimize duplicate TotalExecutableQuantity calls?
			execQty := sdk.MinDec(remainingQty, TotalExecutableQuantity(group.Orders, price))
			market.FillTempOrders(group.Orders, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (market Market) FillTempOrders(orders []*TempOrder, qty, price sdk.Dec, isMaker, halveFees bool) {
	totalExecutableQty := TotalExecutableQuantity(orders, price)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	if qty.LT(totalExecutableQty) { // partial matches
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].HasPriorityOver(orders[j])
		})
	}
	totalExecQty := utils.ZeroDec
	// First, distribute quantity evenly.
	for _, order := range orders {
		remainingQty := qty.Sub(totalExecQty)
		if remainingQty.IsZero() {
			break
		}
		executableQty := order.ExecutableQuantity(price)
		if executableQty.IsZero() {
			continue
		}
		execQty := sdk.MinDec(
			remainingQty,
			sdk.MinDec(
				executableQty,
				order.Quantity.MulTruncate(qty).QuoTruncate(totalExecutableQty)))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
	// Then, distribute remaining quantity based on priority.
	// TODO: sort?
	for _, order := range orders {
		remainingQty := qty.Sub(totalExecQty)
		if remainingQty.IsZero() {
			break
		}
		execQty := sdk.MinDec(remainingQty, order.ExecutableQuantity(price))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker, halveFees)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (market Market) FillTempOrder(order *TempOrder, qty, price sdk.Dec, isMaker, halveFees bool) {
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

type TempOrderBookSide struct {
	IsBuy  bool
	Levels []*TempOrderBookLevel
}

func NewTempOrderBookSide(isBuy bool) *TempOrderBookSide {
	return &TempOrderBookSide{IsBuy: isBuy}
}

func (side *TempOrderBookSide) AddOrder(order *TempOrder) {
	if order.IsBuy != side.IsBuy { // sanity check
		panic("inconsistent order isBuy")
	}
	i := sort.Search(len(side.Levels), func(i int) bool {
		if side.IsBuy {
			return side.Levels[i].Price.LTE(order.Price)
		}
		return side.Levels[i].Price.GTE(order.Price)
	})
	if i < len(side.Levels) && side.Levels[i].Price.Equal(order.Price) {
		side.Levels[i].Orders = append(side.Levels[i].Orders, order)
	} else {
		// Insert a new level.
		newLevels := make([]*TempOrderBookLevel, len(side.Levels)+1)
		copy(newLevels[:i], side.Levels[:i])
		newLevels[i] = NewTempOrderBookLevel(order)
		copy(newLevels[i+1:], side.Levels[i:])
		side.Levels = newLevels
	}
}

type TempOrderBookLevel struct {
	IsBuy  bool
	Price  sdk.Dec
	Orders []*TempOrder
}

func NewTempOrderBookLevel(order *TempOrder) *TempOrderBookLevel {
	return &TempOrderBookLevel{order.IsBuy, order.Price, []*TempOrder{order}}
}

type TempOrder struct {
	Order
	Source           OrderSource
	IsUpdated        bool
	ExecutedQuantity sdk.Dec
	Paid             sdk.DecCoin
	Received         sdk.DecCoins
}

func NewTempOrder(order Order, market Market, source OrderSource) *TempOrder {
	var payDenom string
	if order.IsBuy {
		payDenom = market.QuoteDenom
	} else {
		payDenom = market.BaseDenom
	}
	return &TempOrder{
		Order:            order,
		Source:           source,
		IsUpdated:        false,
		ExecutedQuantity: utils.ZeroDec,
		Paid:             sdk.NewDecCoin(payDenom, utils.ZeroInt),
		Received:         nil,
	}
}

func (order *TempOrder) HasPriorityOver(other *TempOrder) bool {
	if !order.Price.Equal(other.Price) { // sanity check
		panic(fmt.Sprintf("orders with different price: %s != %s", order.Price, other.Price))
	}
	if !order.Quantity.Equal(other.Quantity) {
		return order.Quantity.GT(other.Quantity)
	}
	switch {
	case order.Source == nil && other.Source == nil: // both user orders
		return order.Id < other.Id
	case order.Source == nil && other.Source != nil: // only the first order is user order
		return true
	case order.Source != nil && other.Source == nil: // only the second order is user order
		return false
	default: // both orders from OrderSource
		return order.Source.Name() < other.Source.Name() // lexicographical ordering
	}
}

type TempOrderGroup struct {
	MsgHeight int64
	Orders    []*TempOrder
}

func GroupTempOrdersByMsgHeight(orders []*TempOrder) (groups []*TempOrderGroup) {
	groupByMsgHeight := map[int64]*TempOrderGroup{}
	for _, order := range orders {
		group, ok := groupByMsgHeight[order.MsgHeight]
		if !ok {
			i := sort.Search(len(groups), func(i int) bool {
				if order.MsgHeight == 0 {
					return groups[i].MsgHeight == 0
				}
				if groups[i].MsgHeight == 0 {
					return true
				}
				return order.MsgHeight <= groups[i].MsgHeight
			})
			group = &TempOrderGroup{MsgHeight: order.MsgHeight}
			groupByMsgHeight[order.MsgHeight] = group

			newGroups := make([]*TempOrderGroup, len(groups)+1)
			copy(newGroups[:i], groups[:i])
			newGroups[i] = group
			copy(newGroups[i+1:], groups[i:])
			groups = newGroups
		}
		group.Orders = append(group.Orders, order)
	}
	return
}

func TotalExecutableQuantity(orders []*TempOrder, price sdk.Dec) sdk.Dec {
	qty := utils.ZeroDec
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity(price))
	}
	return qty
}
