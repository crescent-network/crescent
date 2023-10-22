package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

type MemOrderType int

const (
	UserMemOrder MemOrderType = iota + 1
	OrderSourceMemOrder
)

func (typ MemOrderType) String() string {
	switch typ {
	case UserMemOrder:
		return "UserMemOrder"
	case OrderSourceMemOrder:
		return "OrderSourceMemOrder"
	default:
		return fmt.Sprintf("MemOrderType(%d)", typ)
	}
}

type MemOrder struct {
	Type MemOrderType

	Order  *Order      // nil for OrderSourceMemOrder
	Source OrderSource // nil for UserMemOrder

	OrdererAddress   sdk.AccAddress
	IsBuy            bool
	Price            sdk.Dec
	Quantity         sdk.Int
	OpenQuantity     sdk.Int
	RemainingDeposit sdk.Int

	MatchState
}

func NewUserMemOrder(order Order) *MemOrder {
	remainingDeposit := order.RemainingDeposit.ToDec()
	return &MemOrder{
		Type:             UserMemOrder,
		Order:            &order,
		OrdererAddress:   order.MustGetOrdererAddress(),
		IsBuy:            order.IsBuy,
		Price:            order.Price,
		Quantity:         order.Quantity,
		OpenQuantity:     order.OpenQuantity,
		RemainingDeposit: order.RemainingDeposit,
		MatchState:       NewMatchState(&remainingDeposit),
	}
}

func NewOrderSourceMemOrder(
	ordererAddr sdk.AccAddress, isBuy bool, price sdk.Dec, qty, openQty, remainingDeposit sdk.Int, source OrderSource) *MemOrder {
	remainingDepositDec := remainingDeposit.ToDec()
	return &MemOrder{
		Type:             OrderSourceMemOrder,
		OrdererAddress:   ordererAddr,
		Source:           source,
		IsBuy:            isBuy,
		Price:            price,
		Quantity:         qty,
		OpenQuantity:     openQty,
		RemainingDeposit: remainingDeposit,
		MatchState:       NewMatchState(&remainingDepositDec),
	}
}

func (order *MemOrder) String() string {
	isBuyStr := "buy"
	if !order.IsBuy {
		isBuyStr = "sell"
	}
	return fmt.Sprintf(
		"{%s %s %s %s/%s}",
		order.Type, isBuyStr, order.Price, order.MatchState.executedQty, order.OpenQuantity)
}

func (order *MemOrder) ExecutableQuantity() sdk.Int {
	executableQty := order.OpenQuantity.Sub(order.executedQty)
	if order.IsBuy {
		return utils.MinInt(
			executableQty,
			order.remainingDeposit.QuoTruncate(order.Price).TruncateInt())
	}
	return utils.MinInt(executableQty, order.remainingDeposit.TruncateInt())
}

func (order *MemOrder) HasPriorityOver(other *MemOrder) bool {
	if !order.Price.Equal(other.Price) { // sanity check
		panic(fmt.Sprintf("orders with different price: %s != %s", order.Price, other.Price))
	}
	if order.IsBuy != other.IsBuy { // sanity check
		panic(fmt.Sprintf("orders with different direction(isBuy): %v != %v", order.IsBuy, other.IsBuy))
	}
	if !order.Quantity.Equal(other.Quantity) {
		return order.Quantity.GT(other.Quantity)
	}
	switch {
	case order.Type == UserMemOrder && other.Type == UserMemOrder:
		return order.Order.Id < other.Order.Id
	case order.Type == UserMemOrder && other.Type == OrderSourceMemOrder:
		return false
	case order.Type == OrderSourceMemOrder && other.Type == UserMemOrder:
		return true
	default:
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

func (level *MemOrderBookPriceLevel) AddOrder(order *MemOrder) {
	if order.IsBuy != level.IsBuy { // sanity check
		panic("wrong order direction")
	}
	level.Orders = append(level.Orders, order)
}

// MemOrderBookSideOptions is options passed when constructing MemOrderBookSide.
type MemOrderBookSideOptions struct {
	IsBuy             bool
	PriceLimit        *sdk.Dec
	QuantityLimit     *sdk.Int
	QuoteLimit        *sdk.Int
	MaxNumPriceLevels int
}

func (opts MemOrderBookSideOptions) ReachedLimit(price sdk.Dec, accQty sdk.Int, accQuote sdk.Dec, numPriceLevels int) (reached bool) {
	if opts.PriceLimit != nil &&
		((opts.IsBuy && price.LT(*opts.PriceLimit)) ||
			(!opts.IsBuy && price.GT(*opts.PriceLimit))) {
		return true
	}
	if opts.QuantityLimit != nil && !opts.QuantityLimit.Sub(accQty).IsPositive() {
		return true
	}
	if opts.QuoteLimit != nil && !opts.QuoteLimit.ToDec().Sub(accQuote).IsPositive() {
		return true
	}
	if opts.MaxNumPriceLevels > 0 && numPriceLevels >= opts.MaxNumPriceLevels {
		return true
	}
	return false
}

type MemOrderBookSide struct {
	IsBuy  bool
	Levels []*MemOrderBookPriceLevel
}

func NewMemOrderBookSide(isBuy bool) *MemOrderBookSide {
	return &MemOrderBookSide{IsBuy: isBuy}
}

func (obs *MemOrderBookSide) Orders() []*MemOrder {
	var orders []*MemOrder
	for _, level := range obs.Levels {
		orders = append(orders, level.Orders...)
	}
	return orders
}

func (obs *MemOrderBookSide) Limit(n int) {
	limit := len(obs.Levels)
	if n < limit {
		limit = n
	}
	obs.Levels = obs.Levels[:limit]
}

func (obs *MemOrderBookSide) AddOrder(order *MemOrder) {
	if order.IsBuy != obs.IsBuy { // sanity check
		panic("wrong order direction")
	}
	i := sort.Search(len(obs.Levels), func(i int) bool {
		if obs.IsBuy {
			return obs.Levels[i].Price.LTE(order.Price)
		}
		return obs.Levels[i].Price.GTE(order.Price)
	})
	if i < len(obs.Levels) && obs.Levels[i].Price.Equal(order.Price) {
		obs.Levels[i].AddOrder(order)
	} else {
		// Insert a new level.
		newLevels := make([]*MemOrderBookPriceLevel, len(obs.Levels)+1)
		copy(newLevels[:i], obs.Levels[:i])
		newLevels[i] = NewMemOrderBookPriceLevel(order)
		copy(newLevels[i+1:], obs.Levels[i:])
		obs.Levels = newLevels
	}
}

func (obs *MemOrderBookSide) String() string {
	var lines []string
	for _, level := range obs.Levels {
		qty := TotalExecutableQuantity(level.Orders)
		lines = append(lines, fmt.Sprintf("%s | %s", level.Price, qty))
	}
	return strings.Join(lines, "\n")
}

type MemOrderGroup struct {
	msgHeight int64
	orders    []*MemOrder
}

func (group *MemOrderGroup) MsgHeight() int64 {
	return group.msgHeight
}

func (group *MemOrderGroup) Orders() []*MemOrder {
	return group.orders
}

func GroupMemOrdersByMsgHeight(orders []*MemOrder) (groups []*MemOrderGroup) {
	var orderSourceOrders, userOrders []*MemOrder
	for _, order := range orders {
		if order.Type == UserMemOrder {
			userOrders = append(userOrders, order)
		} else {
			orderSourceOrders = append(orderSourceOrders, order)
		}
	}
	if len(orderSourceOrders) > 0 {
		groups = append(groups, &MemOrderGroup{msgHeight: -1, orders: orderSourceOrders})
	}
	groupByMsgHeight := map[int64]*MemOrderGroup{}
	for _, order := range userOrders {
		group, ok := groupByMsgHeight[order.Order.MsgHeight]
		if !ok {
			i := sort.Search(len(groups), func(i int) bool {
				return groups[i].msgHeight >= order.Order.MsgHeight
			})
			group = &MemOrderGroup{msgHeight: order.Order.MsgHeight}
			groupByMsgHeight[order.Order.MsgHeight] = group

			newGroups := make([]*MemOrderGroup, len(groups)+1)
			copy(newGroups[:i], groups[:i])
			newGroups[i] = group
			copy(newGroups[i+1:], groups[i:])
			groups = newGroups
		}
		group.orders = append(group.orders, order)
	}
	return
}

func GroupMemOrdersByOrderer(results []*MemOrder) (ordererAddrs []sdk.AccAddress, m map[string][]*MemOrder) {
	m = map[string][]*MemOrder{}
	for _, result := range results {
		orderer := result.OrdererAddress.String()
		if _, ok := m[orderer]; !ok {
			ordererAddrs = append(ordererAddrs, result.OrdererAddress)
		}
		m[orderer] = append(m[orderer], result)
	}
	return
}

func TotalExecutableQuantity(orders []*MemOrder) sdk.Int {
	qty := utils.ZeroInt
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity())
	}
	return qty
}

func TotalExecutableQuantityPrint(orders []*MemOrder) sdk.Int {
	qty := utils.ZeroInt
	fmt.Println("orders: ", orders)
	fmt.Println("qty: ", qty)
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity())
		fmt.Println("qty: ", qty)
	}
	fmt.Println("qty: ", qty)
	return qty
}
