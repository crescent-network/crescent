package types

import (
	"fmt"
	"strings"
)

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

func (ob *OrderBook) AddOrder(order Order) {
	switch order.GetDirection() {
	case SwapDirectionBuy:
		ob.BuyTicks.AddOrder(order)
	case SwapDirectionSell:
		ob.SellTicks.AddOrder(order)
	}
}

func (ob *OrderBook) AddOrders(orders ...Order) {
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

func (ob OrderBook) AllOrders() Orders {
	var orders Orders
	for _, ticks := range []*OrderBookTicks{ob.BuyTicks, ob.SellTicks} {
		orders = append(orders, ticks.AllOrders()...)
	}
	return orders
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
				ob.BuyTicks.Orders(price).OpenBaseCoinAmount(),
				price.String(),
				ob.SellTicks.Orders(price).OpenBaseCoinAmount()))

		price, found = os.DownTickWithOrders(price)
		if !found {
			break
		}
	}
	lines = append(lines, "+--------------+--------------------------+--------------+")
	return strings.Join(lines, "\n")
}
