package amm

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = (*OrderBook)(nil)

// OrderBook is an order book.
type OrderBook struct {
	buys, sells orderBookTicks
}

// NewOrderBook returns a new OrderBook.
func NewOrderBook(orders ...Order) *OrderBook {
	ob := &OrderBook{}
	ob.Add(orders...)
	return ob
}

// Add adds orders to the order book.
func (ob *OrderBook) Add(orders ...Order) {
	for _, order := range orders {
		switch order.GetDirection() {
		case Buy:
			ob.buys.add(order)
		case Sell:
			ob.sells.add(order)
		}
	}
}

// Orders returns all orders in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) Orders() []Order {
	return append(ob.BuyOrders(), ob.SellOrders()...)
}

// OrdersAt returns orders at given price in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) OrdersAt(price sdk.Dec) []Order {
	return append(ob.BuyOrdersAt(price), ob.SellOrdersAt(price)...)
}

// BuyOrders returns all buy orders in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) BuyOrders() []Order {
	return ob.buys.orders()
}

// SellOrders returns all sell orders in the order book.
// Note that the orders are not sorted.
func (ob *OrderBook) SellOrders() []Order {
	return ob.sells.orders()
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

// HighestBuyPrice returns the highest buy price in the order book.
func (ob *OrderBook) HighestBuyPrice() (sdk.Dec, bool) {
	price, _, found := ob.buys.highestPrice()
	return price, found
}

// LowestSellPrice returns the lowest sell price in the order book.
func (ob *OrderBook) LowestSellPrice() (sdk.Dec, bool) {
	price, _, found := ob.sells.lowestPrice()
	return price, found
}

// BuyAmountOver returns the amount of buy orders in the order book
// for price greater or equal than given price.
func (ob *OrderBook) BuyAmountOver(price sdk.Dec) sdk.Int {
	return ob.buys.amountOver(price)
}

// SellAmountUnder returns the amount of sell orders in the order book
// for price less or equal than given price.
func (ob *OrderBook) SellAmountUnder(price sdk.Dec) sdk.Int {
	return ob.sells.amountUnder(price)
}

// BuyOrdersOver returns buy orders in the order book for price greater
// or equal than given price.
// Note that the orders are not sorted.
func (ob *OrderBook) BuyOrdersOver(price sdk.Dec) []Order {
	return ob.buys.ordersOver(price)
}

// SellOrdersUnder returns sell orders in the order book for price less
// or equal than given price.
// Note that the orders are not sorted.
func (ob *OrderBook) SellOrdersUnder(price sdk.Dec) []Order {
	return ob.sells.ordersUnder(price)
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
	b.WriteString("+--------buy---------+------------price-------------+--------sell--------+\n")
	for _, price := range prices {
		buyAmt := TotalOpenAmount(ob.BuyOrdersAt(price))
		sellAmt := TotalOpenAmount(ob.SellOrdersAt(price))
		_, _ = fmt.Fprintf(&b, "| %18s | %28s | %-18s |\n", buyAmt, price.String(), sellAmt)
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
	var prices []sdk.Dec
	for _, tick := range append(ob.buys, ob.sells...) {
		prices = append(prices, tick.price)
	}
	return ob.stringRepresentation(prices)
}

// orderBookTicks represents a list of orderBookTick.
// This type is used for both buy/sell sides of OrderBook.
type orderBookTicks []*orderBookTick

func (ticks orderBookTicks) findPrice(price sdk.Dec) (i int, exact bool) {
	i = sort.Search(len(ticks), func(i int) bool {
		return ticks[i].price.LTE(price)
	})
	if i < len(ticks) && ticks[i].price.Equal(price) {
		exact = true
	}
	return
}

func (ticks *orderBookTicks) add(order Order) {
	i, exact := ticks.findPrice(order.GetPrice())
	if exact {
		(*ticks)[i].add(order)
	} else {
		if i < len(*ticks) {
			// Insert a new order book tick at index i.
			*ticks = append((*ticks)[:i], append([]*orderBookTick{newOrderBookTick(order)}, (*ticks)[i:]...)...)
		} else {
			// Append a new order book tick at the end.
			*ticks = append(*ticks, newOrderBookTick(order))
		}
	}
}

func (ticks orderBookTicks) orders() []Order {
	var orders []Order
	for _, tick := range ticks {
		orders = append(orders, tick.orders...)
	}
	return orders
}

func (ticks orderBookTicks) ordersAt(price sdk.Dec) []Order {
	i, exact := ticks.findPrice(price)
	if !exact {
		return nil
	}
	return ticks[i].orders
}

func (ticks orderBookTicks) highestPrice() (sdk.Dec, int, bool) {
	if len(ticks) == 0 {
		return sdk.Dec{}, 0, false
	}
	for i, tick := range ticks {
		if TotalOpenAmount(tick.orders).IsPositive() {
			return tick.price, i, true
		}
	}
	return sdk.Dec{}, 0, false
}

func (ticks orderBookTicks) lowestPrice() (sdk.Dec, int, bool) {
	if len(ticks) == 0 {
		return sdk.Dec{}, 0, false
	}
	for i := len(ticks) - 1; i >= 0; i-- {
		if TotalOpenAmount(ticks[i].orders).IsPositive() {
			return ticks[i].price, i, true
		}
	}
	return sdk.Dec{}, 0, false
}

func (ticks orderBookTicks) amountOver(price sdk.Dec) sdk.Int {
	i, exact := ticks.findPrice(price)
	if !exact {
		i--
	}
	amt := sdk.ZeroInt()
	for ; i >= 0; i-- {
		amt = amt.Add(TotalOpenAmount(ticks[i].orders))
	}
	return amt
}

func (ticks orderBookTicks) amountUnder(price sdk.Dec) sdk.Int {
	i, _ := ticks.findPrice(price)
	amt := sdk.ZeroInt()
	for ; i < len(ticks); i++ {
		amt = amt.Add(TotalOpenAmount(ticks[i].orders))
	}
	return amt
}

func (ticks orderBookTicks) ordersOver(price sdk.Dec) []Order {
	i, exact := ticks.findPrice(price)
	if !exact {
		i--
	}
	var orders []Order
	for ; i >= 0; i-- {
		orders = append(orders, ticks[i].orders...)
	}
	return orders
}

func (ticks orderBookTicks) ordersUnder(price sdk.Dec) []Order {
	i, _ := ticks.findPrice(price)
	var orders []Order
	for ; i < len(ticks); i++ {
		orders = append(orders, ticks[i].orders...)
	}
	return orders
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

func (tick *orderBookTick) add(order Order) {
	if !order.GetPrice().Equal(tick.price) {
		panic(fmt.Sprintf("order price %q != tick price %q", order.GetPrice(), tick.price))
	}
	if first := tick.orders[0]; first.GetDirection() != order.GetDirection() {
		panic(fmt.Sprintf("order direction %q != tick direction %q", order.GetDirection(), first.GetDirection()))
	}
	tick.orders = append(tick.orders, order)
}
