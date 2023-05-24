package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (market Market) MatchOrderBookLevels(
	levelA *TempOrderBookLevel, isMakerA bool, levelB *TempOrderBookLevel, isMakerB bool, price sdk.Dec) (execQty sdk.Int, fullA, fullB bool) {
	executableQtyA := TotalExecutableQuantity(levelA.Orders, price)
	executableQtyB := TotalExecutableQuantity(levelB.Orders, price)
	execQty = utils.MinInt(executableQtyA, executableQtyB)
	fullA = execQty.Equal(executableQtyA)
	fullB = execQty.Equal(executableQtyB)
	market.FillTempOrderBookLevel(levelA, execQty, price, isMakerA)
	market.FillTempOrderBookLevel(levelB, execQty, price, isMakerB)
	return
}

func (market Market) FillTempOrderBookLevel(level *TempOrderBookLevel, qty sdk.Int, price sdk.Dec, isMaker bool) {
	executableQty := TotalExecutableQuantity(level.Orders, price)
	if executableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	if executableQty.LTE(qty) { // full matches
		market.FillTempOrders(level.Orders, qty, price, isMaker)
	} else {
		groups := GroupTempOrdersByMsgHeight(level.Orders)
		totalExecQty := utils.ZeroInt
		for _, group := range groups {
			remainingQty := qty.Sub(totalExecQty)
			if remainingQty.IsZero() {
				break
			}
			// TODO: optimize duplicate TotalExecutableQuantity calls?
			execQty := utils.MinInt(remainingQty, TotalExecutableQuantity(group.Orders, price))
			market.FillTempOrders(group.Orders, execQty, price, isMaker)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (market Market) FillTempOrders(orders []*TempOrder, qty sdk.Int, price sdk.Dec, isMaker bool) {
	totalExecutableQty := TotalExecutableQuantity(orders, price)
	if totalExecutableQty.LT(qty) { // sanity check
		panic("executable quantity is less than quantity")
	}
	if qty.LT(totalExecutableQty) { // partial matches
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].HasPriorityOver(orders[j])
		})
	}
	totalExecutableQtyDec := totalExecutableQty.ToDec()
	totalExecQty := utils.ZeroInt
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
		ratio := order.Quantity.ToDec().QuoTruncate(totalExecutableQtyDec)
		execQty := utils.MinInt(
			remainingQty,
			utils.MinInt(executableQty, ratio.MulInt(order.Quantity).TruncateInt()))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
	// Then, distribute remaining quantity based on priority.
	for _, order := range orders {
		remainingQty := qty.Sub(totalExecQty)
		if remainingQty.IsZero() {
			break
		}
		execQty := utils.MinInt(remainingQty, order.ExecutableQuantity(price))
		if execQty.IsPositive() {
			market.FillTempOrder(order, execQty, price, isMaker)
			totalExecQty = totalExecQty.Add(execQty)
		}
	}
}

func (market Market) FillTempOrder(order *TempOrder, qty sdk.Int, price sdk.Dec, isMaker bool) {
	if qty.GT(order.OpenQuantity) { // sanity check
		panic("open quantity is less than quantity")
	}
	negativeMakerFeeRate := market.MakerFeeRate.IsNegative()
	order.ExecutedQuantity = order.ExecutedQuantity.Add(qty)
	order.OpenQuantity = order.OpenQuantity.Sub(qty)
	if order.IsBuy {
		paid := QuoteAmount(true, price, qty)
		order.Paid = order.Paid.AddAmount(paid)
		order.RemainingDeposit = order.RemainingDeposit.Sub(paid)
		if order.Source != nil || (isMaker && negativeMakerFeeRate) {
			order.Received = order.Received.Add(sdk.NewCoin(market.BaseDenom, qty))
		} else {
			if isMaker {
				order.Received = order.Received.Add(
					sdk.NewCoin(
						market.BaseDenom,
						utils.OneDec.Sub(market.MakerFeeRate).MulInt(qty).TruncateInt()))
			} else {
				order.Received = order.Received.Add(
					sdk.NewCoin(
						market.BaseDenom,
						utils.OneDec.Sub(market.TakerFeeRate).MulInt(qty).TruncateInt()))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.Received = order.Received.Add(
				sdk.NewCoin(
					market.QuoteDenom,
					market.MakerFeeRate.Neg().MulInt(paid).TruncateInt()))
		}
	} else {
		order.Paid = order.Paid.AddAmount(qty)
		order.RemainingDeposit = order.RemainingDeposit.Sub(qty)
		quote := QuoteAmount(false, price, qty)
		if order.Source != nil || (isMaker && negativeMakerFeeRate) {
			order.Received = order.Received.Add(sdk.NewCoin(market.QuoteDenom, quote))
		} else {
			if isMaker {
				order.Received = order.Received.Add(
					sdk.NewCoin(
						market.QuoteDenom,
						utils.OneDec.Sub(market.MakerFeeRate).MulInt(quote).TruncateInt()))
			} else {
				order.Received = order.Received.Add(
					sdk.NewCoin(
						market.QuoteDenom,
						utils.OneDec.Sub(market.TakerFeeRate).MulInt(quote).TruncateInt()))
			}
		}
		if isMaker && negativeMakerFeeRate {
			order.Received = order.Received.Add(
				sdk.NewCoin(
					market.BaseDenom,
					market.MakerFeeRate.Neg().MulInt(qty).TruncateInt()))
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
	ExecutedQuantity sdk.Int
	Paid             sdk.Coin
	Received         sdk.Coins
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
		ExecutedQuantity: utils.ZeroInt,
		Paid:             sdk.NewCoin(payDenom, utils.ZeroInt),
		Received:         nil,
	}
}

func (order *TempOrder) HasPriorityOver(other *TempOrder) bool {
	if !order.Quantity.Equal(other.Quantity) {
		return order.Quantity.GT(other.Quantity)
	}
	if (order.Source == nil) != (other.Source == nil) {
		// If the orders aren't in the same type, give priority to user order.
		return order.Source == nil
	}
	return order.Id < other.Id
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

func TotalExecutableQuantity(orders []*TempOrder, price sdk.Dec) sdk.Int {
	qty := utils.ZeroInt
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity(price))
	}
	return qty
}
