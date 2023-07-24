package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type MemOrder struct {
	Order
	Source           OrderSource
	IsUpdated        bool
	ExecutedQuantity sdk.Dec
	Paid             sdk.DecCoin
	Received         sdk.DecCoins
}

func NewMemOrder(order Order, market Market, source OrderSource) *MemOrder {
	var payDenom string
	if order.IsBuy {
		payDenom = market.QuoteDenom
	} else {
		payDenom = market.BaseDenom
	}
	return &MemOrder{
		Order:            order,
		Source:           source,
		IsUpdated:        false,
		ExecutedQuantity: utils.ZeroDec,
		Paid:             sdk.NewDecCoin(payDenom, utils.ZeroInt),
		Received:         nil,
	}
}

func (order *MemOrder) HasPriorityOver(other *MemOrder) bool {
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

type MemOrderBookPriceLevel struct {
	IsBuy  bool
	Price  sdk.Dec
	Orders []*MemOrder
}

func NewMemOrderBookPriceLevel(order *MemOrder) *MemOrderBookPriceLevel {
	return &MemOrderBookPriceLevel{order.IsBuy, order.Price, []*MemOrder{order}}
}

type MemOrderBookSide struct {
	IsBuy  bool
	Levels []*MemOrderBookPriceLevel
}

func NewMemOrderBookSide(isBuy bool) *MemOrderBookSide {
	return &MemOrderBookSide{IsBuy: isBuy}
}

func (side *MemOrderBookSide) AddOrder(order *MemOrder) {
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
		newLevels := make([]*MemOrderBookPriceLevel, len(side.Levels)+1)
		copy(newLevels[:i], side.Levels[:i])
		newLevels[i] = NewMemOrderBookPriceLevel(order)
		copy(newLevels[i+1:], side.Levels[i:])
		side.Levels = newLevels
	}
}

type MemOrderGroup struct {
	MsgHeight int64
	Orders    []*MemOrder
}

func GroupMemOrdersByMsgHeight(orders []*MemOrder) (groups []*MemOrderGroup) {
	groupByMsgHeight := map[int64]*MemOrderGroup{}
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
			group = &MemOrderGroup{MsgHeight: order.MsgHeight}
			groupByMsgHeight[order.MsgHeight] = group

			newGroups := make([]*MemOrderGroup, len(groups)+1)
			copy(newGroups[:i], groups[:i])
			newGroups[i] = group
			copy(newGroups[i+1:], groups[i:])
			groups = newGroups
		}
		group.Orders = append(group.Orders, order)
	}
	return
}

func TotalExecutableQuantity(orders []*MemOrder, price sdk.Dec) sdk.Dec {
	qty := utils.ZeroDec
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity(price))
	}
	return qty
}
