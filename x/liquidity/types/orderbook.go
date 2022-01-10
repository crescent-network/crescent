package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = (*OrderBook)(nil)

type PriceDirection int

const (
	PriceIncreasing PriceDirection = iota + 1
	PriceDecreasing
)

type Orders []Order

func (orders Orders) RemainingAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.RemainingAmount())
	}
	return amount
}

type OrderBookTick struct {
	price sdk.Dec
	buys  Orders
	sells Orders
}

func NewOrderBookTick(order Order) *OrderBookTick {
	tick := &OrderBookTick{
		price: order.Price(),
	}
	switch order.Direction() {
	case SwapDirectionBuy:
		tick.buys = append(tick.buys, order)
	case SwapDirectionSell:
		tick.sells = append(tick.sells, order)
	}

	return tick
}

type OrderBook struct {
	ticks []*OrderBookTick
}

func (ob OrderBook) AddOrder(order Order) {
	var prices []sdk.Dec
	for _, tick := range ob.ticks {
		prices = append(prices, tick.price)
	}
	i := sort.Search(len(prices), func(i int) bool {
		return prices[i].LT(order.Price())
	})
	if i < len(prices) {
		if prices[i].Equal(order.Price()) {
			switch order.Direction() {
			case SwapDirectionBuy:
				ob.ticks[i].buys = append(ob.ticks[i].buys, order)
			case SwapDirectionSell:
				ob.ticks[i].sells = append(ob.ticks[i].sells, order)
			}
		} else {
			// Insert a new order group at index i.
			ob.ticks = append(ob.ticks[:i], append([]*OrderBookTick{NewOrderBookTick(order)}, ob.ticks[i:]...)...)
		}
	} else {
		// Append a new order group at the end.
		ob.ticks = append(ob.ticks, NewOrderBookTick(order))
	}
}

func (ob OrderBook) AddOrders(orders ...Order) {
	for _, order := range orders {
		ob.AddOrder(order)
	}
}

func (ob OrderBook) AmountGTE(price sdk.Dec) sdk.Int {
	return sdk.ZeroInt()
}

func (ob OrderBook) AmountLTE(price sdk.Dec) sdk.Int {
	return sdk.ZeroInt()
}

func (ob OrderBook) Orders(price sdk.Dec) []Order {
	return nil
}

func (ob OrderBook) UpTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool) {
	return sdk.ZeroDec(), false
}

func (ob OrderBook) DownTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool) {
	return sdk.ZeroDec(), false
}

func (ob OrderBook) String() string {
	lines := []string{
		"+-----buy------+----------price-----------+-----sell-----+",
	}
	for _, tick := range ob.ticks {
		lines = append(lines,
			fmt.Sprintf("| %12s | %24s | %-12s |",
				tick.buys.RemainingAmount(), tick.price.String(), tick.sells.RemainingAmount()))
	}
	lines = append(lines, "+--------------+--------------------------+--------------+")
	return strings.Join(lines, "\n")
}

//func (ob OrderBook) HighestPriceXToYItemIndex(start int) (idx int, found bool) {
//	for i := start; i < len(ob); i++ {
//		if len(ob[i].XToYOrders) > 0 {
//			idx = i
//			found = true
//			return
//		}
//	}
//	return
//}
//
//func (ob OrderBook) LowestPriceYToXItemIndex(start int) (idx int, found bool) {
//	for i := start; i >= 0; i-- {
//		if len(ob[i].YToXOrders) > 0 {
//			idx = i
//			found = true
//			return
//		}
//	}
//	return
//}
//

//// DemandingAmount returns total demanding amount of orders at given price.
//// Demanding amount is the amount of coins these orders want to receive.
//// Note that orders should have same SwapDirection, since
//// DemandingAmount doesn't rely on SwapDirection.
//// TODO: use sdk.Dec here?
//func (os Orders) DemandingAmount(price sdk.Dec) sdk.Int {
//	da := sdk.ZeroInt()
//	for _, order := range os {
//		switch order.Direction {
//		case SwapDirectionXToY:
//			da = da.Add(order.RemainingAmount.ToDec().QuoTruncate(price).TruncateInt())
//		case SwapDirectionYToX:
//			da = da.Add(order.RemainingAmount.ToDec().MulTruncate(price).TruncateInt())
//		}
//	}
//	return da
//}
//
//// MatchOrders matches two order groups at given price.
//func MatchOrders(a, b Orders, price sdk.Dec) {
//	amtA := a.RemainingAmount()
//	amtB := b.RemainingAmount()
//	daA := a.DemandingAmount(price)
//	daB := b.DemandingAmount(price)
//
//	var sos, bos Orders // smaller orders, bigger orders
//	// Remaining amount and demanding amount of smaller orders and bigger orders.
//	var sa, ba, sda sdk.Int
//	if amtA.LTE(daB) { // a is smaller than(or equal to) b
//		if daA.GT(amtB) { // sanity check TODO: remove
//			panic(fmt.Sprintf("%s > %s!", daA, amtB))
//		}
//		sos, bos = a, b
//		sa, ba = amtA, amtB
//		sda = daA
//	} else { // b is smaller than a
//		if daB.GT(amtA) { // sanity check TODO: remove
//			panic(fmt.Sprintf("%s > %s!", daB, amtA))
//		}
//		sos, bos = b, a
//		sa, ba = amtB, amtA
//		sda = daB
//	}
//
//	if sa.IsZero() || ba.IsZero() { // TODO: need more zero value checks?
//		return
//	}
//
//	for _, order := range sos {
//		proportion := order.RemainingAmount.ToDec().QuoTruncate(sa.ToDec()) // RemainingAmount / sa
//		order.RemainingAmount = sdk.ZeroInt()
//		var in sdk.Int
//		if sa.Equal(ba) {
//			in = ba.ToDec().MulTruncate(proportion).TruncateInt() // ba * proportion
//		} else {
//			in = sda.ToDec().MulTruncate(proportion).TruncateInt() // sda * proportion
//		}
//		order.ReceivedAmount = order.ReceivedAmount.Add(in)
//	}
//
//	for _, order := range bos {
//		proportion := order.RemainingAmount.ToDec().QuoTruncate(ba.ToDec()) // RemainingAmount / ba
//		if sa.Equal(ba) {
//			order.RemainingAmount = sdk.ZeroInt()
//		} else {
//			out := sda.ToDec().MulTruncate(proportion).TruncateInt() // sda * proportion
//			order.RemainingAmount = order.RemainingAmount.Sub(out)
//		}
//		in := sa.ToDec().MulTruncate(proportion).TruncateInt() // sa * proportion
//		order.ReceivedAmount = order.ReceivedAmount.Add(in)
//	}
//}

//// Order represents a swap order, which is made by a user or a pool.
//type Order struct {
//	Orderer         sdk.AccAddress
//	Direction       SwapDirection
//	Price           sdk.Dec
//	RemainingAmount sdk.Int
//	ReceivedAmount  sdk.Int
//}

type Order interface {
	Orderer() sdk.AccAddress
	Direction() SwapDirection
	Price() sdk.Dec
	RemainingAmount() sdk.Int
	SubRemainingAmount(amount sdk.Int)
	ReceivedAmount() sdk.Int
	AddReceivedAmount(amount sdk.Int)
}
