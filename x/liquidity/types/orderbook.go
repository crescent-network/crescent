package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ OrderSource = (*OrderBookTicks)(nil)
	_ OrderSource = (*MergedOrderSources)(nil)
)

type PriceDirection int

const (
	PriceIncreasing PriceDirection = iota + 1
	PriceDecreasing
)

type Order struct {
	Orderer         sdk.AccAddress
	Direction       SwapDirection
	Price           sdk.Dec
	OrderAmount     sdk.Int
	RemainingAmount sdk.Int
	ReceivedAmount  sdk.Int
}

func NewOrder(orderer sdk.AccAddress, dir SwapDirection, price sdk.Dec, amount sdk.Int) *Order {
	return &Order{
		Orderer:         orderer,
		Direction:       dir,
		Price:           price,
		OrderAmount:     amount,
		RemainingAmount: amount,
		ReceivedAmount:  sdk.ZeroInt(),
	}
}

// OrderSource defines a source of orders which can be an order book or
// a pool.
type OrderSource interface {
	AmountGTE(price sdk.Dec) sdk.Int
	AmountLTE(price sdk.Dec) sdk.Int
	Orders(price sdk.Dec) Orders
	UpTick(price sdk.Dec) (tick sdk.Dec, found bool)
	DownTick(price sdk.Dec) (tick sdk.Dec, found bool)
	UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool)
	DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool)
	HighestTick() (tick sdk.Dec, found bool)
	LowestTick() (tick sdk.Dec, found bool)
}

type Orders []*Order

func (orders Orders) RemainingAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.RemainingAmount)
	}
	return amount
}

type OrderBookTick struct {
	Price  sdk.Dec
	Orders Orders
}

func NewOrderBookTick(order *Order) *OrderBookTick {
	return &OrderBookTick{
		Price:  order.Price,
		Orders: Orders{order},
	}
}

type OrderBookTicks struct {
	Ticks         []*OrderBookTick
	TickPrecision int
}

func NewOrderBookTicks(prec int) *OrderBookTicks {
	return &OrderBookTicks{
		TickPrecision: prec,
	}
}

func (ticks *OrderBookTicks) FindPrice(price sdk.Dec) (i int, exact bool) {
	i = sort.Search(len(ticks.Ticks), func(i int) bool {
		return ticks.Ticks[i].Price.LTE(price)
	})
	if i < len(ticks.Ticks) && ticks.Ticks[i].Price.Equal(price) {
		exact = true
	}
	return
}

func (ticks *OrderBookTicks) AddOrder(order *Order) {
	i, exact := ticks.FindPrice(order.Price)
	if exact {
		ticks.Ticks[i].Orders = append(ticks.Ticks[i].Orders, order)
	} else {
		if i < len(ticks.Ticks) {
			// Insert a new order book tick at index i.
			ticks.Ticks = append(ticks.Ticks[:i], append([]*OrderBookTick{NewOrderBookTick(order)}, ticks.Ticks[i:]...)...)
		} else {
			// Append a new order group at the end.
			ticks.Ticks = append(ticks.Ticks, NewOrderBookTick(order))
		}
	}
}

func (ticks *OrderBookTicks) AddOrders(orders ...*Order) {
	for _, order := range orders {
		ticks.AddOrder(order)
	}
}

func (ticks *OrderBookTicks) AmountGTE(price sdk.Dec) sdk.Int {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	amount := sdk.ZeroInt()
	for ; i >= 0; i-- {
		amount = amount.Add(ticks.Ticks[i].Orders.RemainingAmount())
	}
	return amount
}

func (ticks OrderBookTicks) AmountLTE(price sdk.Dec) sdk.Int {
	i, _ := ticks.FindPrice(price)
	amount := sdk.ZeroInt()
	for ; i < len(ticks.Ticks); i++ {
		amount = amount.Add(ticks.Ticks[i].Orders.RemainingAmount())
	}
	return amount
}

func (ticks OrderBookTicks) Orders(price sdk.Dec) Orders {
	i, exact := ticks.FindPrice(price)
	if !exact {
		return nil
	}
	return ticks.Ticks[i].Orders
}

func (ticks OrderBookTicks) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	tick = UpTick(price, ticks.TickPrecision)
	found = true
	return
}

func (ticks OrderBookTicks) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	if i >= len(ticks.Ticks)-1 {
		return
	}
	tick = DownTick(price, ticks.TickPrecision)
	found = true
	return
}

func (ticks OrderBookTicks) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	return ticks.Ticks[i-1].Price, true
}

func (ticks OrderBookTicks) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	if i >= len(ticks.Ticks)-1 {
		return
	}
	return ticks.Ticks[i+1].Price, true
}

func (ticks OrderBookTicks) HighestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	return ticks.Ticks[0].Price, true
}

func (ticks OrderBookTicks) LowestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	return ticks.Ticks[len(ticks.Ticks)-1].Price, true
}

type OrderBook struct {
	BuyTicks      *OrderBookTicks
	SellTicks     *OrderBookTicks
	TickPrecision int
}

func NewOrderBook(prec int) *OrderBook {
	return &OrderBook{
		BuyTicks:      NewOrderBookTicks(prec),
		SellTicks:     NewOrderBookTicks(prec),
		TickPrecision: prec,
	}
}

func (ob *OrderBook) AddOrder(order *Order) {
	switch order.Direction {
	case SwapDirectionBuy:
		ob.BuyTicks.AddOrder(order)
	case SwapDirectionSell:
		ob.SellTicks.AddOrder(order)
	}
}

func (ob *OrderBook) AddOrders(orders ...*Order) {
	for _, order := range orders {
		ob.AddOrder(order)
	}
}

func (ob OrderBook) OrderSource(dir SwapDirection) OrderSource {
	switch dir {
	case SwapDirectionBuy:
		return ob.BuyTicks
	case SwapDirectionSell:
		return ob.SellTicks
	default:
		panic(fmt.Sprintf("unknown swap direction: %v", dir))
	}
}

func (ob OrderBook) String() string {
	os := MergeOrderSources(ob.BuyTicks, ob.SellTicks)
	price, found := os.HighestTick()
	if !found {
		return "<nil>"
	}
	lines := []string{
		"+-----buy------+----------price-----------+-----sell-----+",
	}
	for {
		lines = append(lines,
			fmt.Sprintf("| %12s | %24s | %-12s |",
				ob.BuyTicks.Orders(price).RemainingAmount(),
				price.String(),
				ob.SellTicks.Orders(price).RemainingAmount()))

		price, found = os.DownTickWithOrders(price)
		if !found {
			break
		}
	}
	lines = append(lines, "+--------------+--------------------------+--------------+")
	return strings.Join(lines, "\n")
}

type MergedOrderSources struct {
	Sources []OrderSource
}

func MergeOrderSources(sources ...OrderSource) OrderSource {
	return &MergedOrderSources{Sources: sources}
}

func (os *MergedOrderSources) AmountGTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.Sources {
		amt = amt.Add(source.AmountGTE(price))
	}
	return amt
}

func (os *MergedOrderSources) AmountLTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.Sources {
		amt = amt.Add(source.AmountLTE(price))
	}
	return amt
}

func (os *MergedOrderSources) Orders(price sdk.Dec) Orders {
	var orders Orders
	for _, source := range os.Sources {
		orders = append(orders, source.Orders(price)...)
	}
	return orders
}

func (os *MergedOrderSources) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.UpTick(price)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.DownTick(price)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.UpTickWithOrders(price)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.DownTickWithOrders(price)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) HighestTick() (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.HighestTick()
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) LowestTick() (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.LowestTick()
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}
