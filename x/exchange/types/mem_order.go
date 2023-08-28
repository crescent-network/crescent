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
	typ              MemOrderType
	order            *Order      // nil for OrderSourceMemOrder
	source           OrderSource // nil for UserMemOrder
	ordererAddr      sdk.AccAddress
	isBuy            bool
	price            sdk.Dec
	qty              sdk.Dec
	openQty          sdk.Dec
	remainingDeposit sdk.Dec
	executedQty      sdk.Dec
	paid             sdk.Dec
	received         sdk.Dec
	fee              sdk.Dec
	isMatched        bool
	isMaker          *bool
}

func NewUserMemOrder(order Order) *MemOrder {
	return &MemOrder{
		typ:              UserMemOrder,
		order:            &order,
		ordererAddr:      order.MustGetOrdererAddress(),
		isBuy:            order.IsBuy,
		price:            order.Price,
		qty:              order.Quantity,
		openQty:          order.OpenQuantity,
		remainingDeposit: order.RemainingDeposit,
		executedQty:      utils.ZeroDec,
		paid:             utils.ZeroDec,
		received:         utils.ZeroDec,
		fee:              utils.ZeroDec,
	}
}

func NewOrderSourceMemOrder(
	ordererAddr sdk.AccAddress, isBuy bool, price, qty, openQty sdk.Dec, source OrderSource) *MemOrder {
	return &MemOrder{
		typ:              OrderSourceMemOrder,
		ordererAddr:      ordererAddr,
		isBuy:            isBuy,
		price:            price,
		qty:              qty,
		openQty:          openQty,
		remainingDeposit: DepositAmount(isBuy, price, openQty),
		executedQty:      utils.ZeroDec,
		paid:             utils.ZeroDec,
		received:         utils.ZeroDec,
		fee:              utils.ZeroDec,
		source:           source,
	}
}

func (order *MemOrder) String() string {
	isBuyStr := "buy"
	if !order.isBuy {
		isBuyStr = "sell"
	}
	return fmt.Sprintf(
		"{%s %s %s %s}",
		order.typ, isBuyStr, order.price, order.openQty)
}

func (order *MemOrder) Type() MemOrderType {
	return order.typ
}

func (order *MemOrder) Order() Order {
	return *order.order
}

func (order *MemOrder) Source() OrderSource {
	return order.source
}

func (order *MemOrder) OrdererAddress() sdk.AccAddress {
	return order.ordererAddr
}

func (order *MemOrder) IsBuy() bool {
	return order.isBuy
}

func (order *MemOrder) Price() sdk.Dec {
	return order.price
}

func (order *MemOrder) Quantity() sdk.Dec {
	return order.qty
}

func (order *MemOrder) OpenQuantity() sdk.Dec {
	return order.openQty
}

func (order *MemOrder) RemainingDeposit() sdk.Dec {
	return order.remainingDeposit
}

func (order *MemOrder) ExecutedQuantity() sdk.Dec {
	return order.executedQty
}

func (order *MemOrder) Paid() sdk.Dec {
	return order.paid
}

func (order *MemOrder) PaidWithoutFee() sdk.Dec {
	if order.fee.IsNegative() {
		return order.paid.Sub(order.fee)
	}
	return order.paid
}

func (order *MemOrder) Received() sdk.Dec {
	return order.received
}

func (order *MemOrder) Fee() sdk.Dec {
	return order.fee
}

func (order *MemOrder) IsMatched() bool {
	return order.isMatched
}

func (order *MemOrder) ExecutableQuantity() sdk.Dec {
	executableQty := order.openQty.Sub(order.executedQty)
	if order.isBuy {
		return sdk.MinDec(executableQty, order.remainingDeposit.QuoTruncate(order.price))
	}
	return executableQty
}

func (order *MemOrder) HasPriorityOver(other *MemOrder) bool {
	if !order.price.Equal(other.price) { // sanity check
		panic(fmt.Sprintf("orders with different price: %s != %s", order.price, other.price))
	}
	if !order.qty.Equal(other.qty) {
		return order.qty.GT(other.qty)
	}
	switch {
	case order.typ == UserMemOrder && other.typ == UserMemOrder:
		return order.order.Id < other.order.Id
	case order.typ == UserMemOrder && other.typ == OrderSourceMemOrder:
		return false
	case order.typ == OrderSourceMemOrder && other.typ == UserMemOrder:
		return true
	default:
		return order.source.Name() < other.source.Name() // lexicographical ordering
	}
}

type MemOrderBookPriceLevel struct {
	isBuy  bool
	price  sdk.Dec
	orders []*MemOrder
}

func NewMemOrderBookPriceLevel(order *MemOrder) *MemOrderBookPriceLevel {
	return &MemOrderBookPriceLevel{order.isBuy, order.price, []*MemOrder{order}}
}

func (level *MemOrderBookPriceLevel) Price() sdk.Dec {
	return level.price
}

func (level *MemOrderBookPriceLevel) Orders() []*MemOrder {
	return level.orders
}

func (level *MemOrderBookPriceLevel) AddOrder(order *MemOrder) {
	if order.isBuy != level.isBuy { // sanity check
		panic("wrong order direction")
	}
	level.orders = append(level.orders, order)
}

// MemOrderBookSideOptions is options passed when constructing MemOrderBookSide.
type MemOrderBookSideOptions struct {
	IsBuy             bool
	PriceLimit        *sdk.Dec
	QuantityLimit     *sdk.Dec
	QuoteLimit        *sdk.Dec
	MaxNumPriceLevels int
}

func (opts MemOrderBookSideOptions) ReachedLimit(price, accQty, accQuote sdk.Dec, numPriceLevels int) (reached bool) {
	if opts.PriceLimit != nil &&
		((opts.IsBuy && price.LT(*opts.PriceLimit)) ||
			(!opts.IsBuy && price.GT(*opts.PriceLimit))) {
		return true
	}
	if opts.QuantityLimit != nil && !opts.QuantityLimit.Sub(accQty).IsPositive() {
		return true
	}
	if opts.QuoteLimit != nil && !opts.QuoteLimit.Sub(accQuote).IsPositive() {
		return true
	}
	if opts.MaxNumPriceLevels > 0 && numPriceLevels >= opts.MaxNumPriceLevels {
		return true
	}
	return false
}

type MemOrderBookSide struct {
	isBuy  bool
	levels []*MemOrderBookPriceLevel
}

func NewMemOrderBookSide(isBuy bool) *MemOrderBookSide {
	return &MemOrderBookSide{isBuy: isBuy}
}

func (obs *MemOrderBookSide) Levels() []*MemOrderBookPriceLevel {
	return obs.levels
}

func (obs *MemOrderBookSide) Orders() []*MemOrder {
	var orders []*MemOrder
	for _, level := range obs.levels {
		orders = append(orders, level.orders...)
	}
	return orders
}

func (obs *MemOrderBookSide) Limit(n int) {
	limit := len(obs.Levels())
	if n < limit {
		limit = n
	}
	obs.levels = obs.levels[:limit]
}

func (side *MemOrderBookSide) AddOrder(order *MemOrder) {
	if order.isBuy != side.isBuy { // sanity check
		panic("wrong order direction")
	}
	i := sort.Search(len(side.levels), func(i int) bool {
		if side.isBuy {
			return side.levels[i].price.LTE(order.price)
		}
		return side.levels[i].price.GTE(order.price)
	})
	if i < len(side.levels) && side.levels[i].price.Equal(order.price) {
		side.levels[i].AddOrder(order)
	} else {
		// Insert a new level.
		newLevels := make([]*MemOrderBookPriceLevel, len(side.levels)+1)
		copy(newLevels[:i], side.levels[:i])
		newLevels[i] = NewMemOrderBookPriceLevel(order)
		copy(newLevels[i+1:], side.levels[i:])
		side.levels = newLevels
	}
}

func (side *MemOrderBookSide) String() string {
	var lines []string
	for _, level := range side.levels {
		qty := TotalExecutableQuantity(level.orders)
		lines = append(lines, fmt.Sprintf("%s | %s", level.price, qty))
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
		if order.typ == UserMemOrder {
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
		group, ok := groupByMsgHeight[order.order.MsgHeight]
		if !ok {
			i := sort.Search(len(groups), func(i int) bool {
				return groups[i].msgHeight >= order.order.MsgHeight
			})
			group = &MemOrderGroup{msgHeight: order.order.MsgHeight}
			groupByMsgHeight[order.order.MsgHeight] = group

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
		orderer := result.ordererAddr.String()
		if _, ok := m[orderer]; !ok {
			ordererAddrs = append(ordererAddrs, result.ordererAddr)
		}
		m[orderer] = append(m[orderer], result)
	}
	return
}

func TotalExecutableQuantity(orders []*MemOrder) sdk.Dec {
	qty := utils.ZeroDec
	for _, order := range orders {
		qty = qty.Add(order.ExecutableQuantity())
	}
	return qty
}
