package amm

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrderBook is an order book.
type OrderBook struct {
	buys, sells *orderBookTicks
}

// NewOrderBook returns a new OrderBook.
func NewOrderBook(orders ...Order) *OrderBook {
	ob := &OrderBook{
		buys:  newOrderBookBuyTicks(),
		sells: newOrderBookSellTicks(),
	}
	ob.AddOrder(orders...)
	return ob
}

// AddOrder adds orders to the order book.
func (ob *OrderBook) AddOrder(orders ...Order) {
	for _, order := range orders {
		if MatchableAmount(order, order.GetPrice()).IsPositive() {
			switch order.GetDirection() {
			case Buy:
				ob.buys.addOrder(order)
			case Sell:
				ob.sells.addOrder(order)
			}
		}
	}
}

// Orders returns all orders in the order book.
func (ob *OrderBook) Orders() []Order {
	var orders []Order
	for _, tick := range append(ob.buys.ticks, ob.sells.ticks...) {
		orders = append(orders, tick.orders...)
	}
	return orders
}

// BuyOrdersAt returns buy orders at given price in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) BuyOrdersAt(price sdk.Dec) []Order {
	return ob.buys.ordersAt(price)
}

// SellOrdersAt returns sell orders at given price in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) SellOrdersAt(price sdk.Dec) []Order {
	return ob.sells.ordersAt(price)
}

func (ob *OrderBook) HighestPrice() (sdk.Dec, bool) {
	highestBuyPrice, _, foundBuy := ob.buys.highestPrice()
	highestSellPrice, _, foundSell := ob.sells.highestPrice()
	switch {
	case foundBuy && foundSell:
		return sdk.MaxDec(highestBuyPrice, highestSellPrice), true
	case foundBuy:
		return highestBuyPrice, true
	case foundSell:
		return highestSellPrice, true
	default:
		return sdk.Dec{}, false
	}
}

func (ob *OrderBook) LowestPrice() (sdk.Dec, bool) {
	lowestBuyPrice, _, foundBuy := ob.buys.lowestPrice()
	lowestSellPrice, _, foundSell := ob.sells.lowestPrice()
	switch {
	case foundBuy && foundSell:
		return sdk.MinDec(lowestBuyPrice, lowestSellPrice), true
	case foundBuy:
		return lowestBuyPrice, true
	case foundSell:
		return lowestSellPrice, true
	default:
		return sdk.Dec{}, false
	}
}

func (ob *OrderBook) stringRepresentation(prices []sdk.Dec) string {
	if len(prices) == 0 {
		return "<nil>"
	}
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].GT(prices[j])
	})
	var b strings.Builder
	b.WriteString("+--------sell--------+------------price-------------+--------buy---------+\n")
	for _, price := range prices {
		buyAmt := TotalMatchableAmount(ob.BuyOrdersAt(price), price)
		sellAmt := TotalMatchableAmount(ob.SellOrdersAt(price), price)
		_, _ = fmt.Fprintf(&b, "| %18s | %28s | %-18s |\n", sellAmt, price.String(), buyAmt)
	}
	b.WriteString("+--------------------+------------------------------+--------------------+")
	return b.String()
}

// FullString returns a full string representation of the order book.
// FullString includes all possible price ticks from the order book's
// highest price to the lowest price.
func (ob *OrderBook) FullString(tickPrec int) string {
	var prices []sdk.Dec
	highest, found := ob.HighestPrice()
	if !found {
		return "<nil>"
	}
	lowest, _ := ob.LowestPrice()
	for ; lowest.LTE(highest); lowest = UpTick(lowest, tickPrec) {
		prices = append(prices, lowest)
	}
	return ob.stringRepresentation(prices)
}

// String returns a compact string representation of the order book.
// String includes a tick only when there is at least one order on it.
func (ob *OrderBook) String() string {
	priceSet := map[string]sdk.Dec{}
	for _, tick := range append(ob.buys.ticks, ob.sells.ticks...) {
		priceSet[tick.price.String()] = tick.price
	}
	prices := make([]sdk.Dec, 0, len(priceSet))
	for _, price := range priceSet {
		prices = append(prices, price)
	}
	return ob.stringRepresentation(prices)
}

// orderBookTicks represents a list of orderBookTick.
// This type is used for both buy/sell sides of OrderBook.
type orderBookTicks struct {
	ticks           []*orderBookTick
	priceIncreasing bool
}

func newOrderBookBuyTicks() *orderBookTicks {
	return &orderBookTicks{
		priceIncreasing: false,
	}
}

func newOrderBookSellTicks() *orderBookTicks {
	return &orderBookTicks{
		priceIncreasing: true,
	}
}

func (ticks *orderBookTicks) findPrice(price sdk.Dec) (i int, exact bool) {
	i = sort.Search(len(ticks.ticks), func(i int) bool {
		if ticks.priceIncreasing {
			return ticks.ticks[i].price.GTE(price)
		} else {
			return ticks.ticks[i].price.LTE(price)
		}
	})
	if i < len(ticks.ticks) && ticks.ticks[i].price.Equal(price) {
		exact = true
	}
	return
}

func (ticks *orderBookTicks) addOrder(order Order) {
	i, exact := ticks.findPrice(order.GetPrice())
	if exact {
		ticks.ticks[i].addOrder(order)
	} else {
		if i < len(ticks.ticks) {
			// Insert a new order book tick at index i.
			ticks.ticks = append(ticks.ticks[:i], append([]*orderBookTick{newOrderBookTick(order)}, ticks.ticks[i:]...)...)
		} else {
			// Append a new order book tick at the end.
			ticks.ticks = append(ticks.ticks, newOrderBookTick(order))
		}
	}
}

func (ticks *orderBookTicks) ordersAt(price sdk.Dec) []Order {
	i, exact := ticks.findPrice(price)
	if !exact {
		return nil
	}
	return ticks.ticks[i].orders
}

func (ticks *orderBookTicks) highestPrice() (sdk.Dec, int, bool) {
	if len(ticks.ticks) == 0 {
		return sdk.Dec{}, 0, false
	}
	if ticks.priceIncreasing {
		return ticks.ticks[len(ticks.ticks)-1].price, len(ticks.ticks) - 1, true
	} else {
		return ticks.ticks[0].price, 0, true
	}
}

func (ticks *orderBookTicks) lowestPrice() (sdk.Dec, int, bool) {
	if len(ticks.ticks) == 0 {
		return sdk.Dec{}, 0, false
	}
	if ticks.priceIncreasing {
		return ticks.ticks[0].price, 0, true
	} else {
		return ticks.ticks[len(ticks.ticks)-1].price, len(ticks.ticks) - 1, true
	}
}

// orderBookTick represents a tick in OrderBook.
type orderBookTick struct {
	price  sdk.Dec
	orders []Order
}

func newOrderBookTick(order Order) *orderBookTick {
	return &orderBookTick{
		price:  order.GetPrice(),
		orders: []Order{order},
	}
}

func (tick *orderBookTick) addOrder(order Order) {
	tick.orders = append(tick.orders, order)
}
