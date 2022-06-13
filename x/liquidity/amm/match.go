package amm

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PriceDirection specifies estimated price direction within this batch.
type PriceDirection int

const (
	PriceStaying PriceDirection = iota + 1
	PriceIncreasing
	PriceDecreasing
)

func (dir PriceDirection) String() string {
	switch dir {
	case PriceStaying:
		return "PriceStaying"
	case PriceIncreasing:
		return "PriceIncreasing"
	case PriceDecreasing:
		return "PriceDecreasing"
	default:
		return fmt.Sprintf("PriceDirection(%d)", dir)
	}
}

// MatchRecord holds a single match record.
type MatchRecord struct {
	Amount sdk.Int
	Price  sdk.Dec
}

// FillOrder fills the order by given amount and price.
func FillOrder(order Order, amt sdk.Int, price sdk.Dec) (quoteCoinDiff sdk.Int) {
	if amt.GT(order.GetOpenAmount()) {
		panic(fmt.Errorf("cannot match more than open amount; %s > %s", amt, order.GetOpenAmount()))
	}
	var paid, received sdk.Int
	switch order.GetDirection() {
	case Buy:
		paid = price.MulInt(amt).Ceil().TruncateInt()
		received = amt
		quoteCoinDiff = paid
	case Sell:
		paid = amt
		received = price.MulInt(amt).TruncateInt()
		quoteCoinDiff = received.Neg()
	}
	order.SetPaidOfferCoinAmount(order.GetPaidOfferCoinAmount().Add(paid))
	order.SetReceivedDemandCoinAmount(order.GetReceivedDemandCoinAmount().Add(received))
	order.SetOpenAmount(order.GetOpenAmount().Sub(amt))
	order.AddMatchRecord(MatchRecord{
		Amount: amt,
		Price:  price,
	})
	return
}

// FulfillOrder fills the order by its remaining open amount at given price.
func FulfillOrder(order Order, price sdk.Dec) (quoteCoinDiff sdk.Int) {
	quoteCoinDiff = sdk.ZeroInt()
	if order.GetOpenAmount().IsPositive() {
		quoteCoinDiff = quoteCoinDiff.Add(FillOrder(order, order.GetOpenAmount(), price))
	}
	return
}

// FulfillOrders fills multiple orders by their remaining open amount
// at given price.
func FulfillOrders(orders []Order, price sdk.Dec) (quoteCoinDiff sdk.Int) {
	quoteCoinDiff = sdk.ZeroInt()
	for _, order := range orders {
		quoteCoinDiff = quoteCoinDiff.Add(FulfillOrder(order, price))
	}
	return
}

func FindMatchPrice(ov OrderView, tickPrec int) (matchPrice sdk.Dec, found bool) {
	highestBuyPrice, found := ov.HighestBuyPrice()
	if !found {
		return sdk.Dec{}, false
	}
	lowestSellPrice, found := ov.LowestSellPrice()
	if !found {
		return sdk.Dec{}, false
	}
	if highestBuyPrice.LT(lowestSellPrice) {
		return sdk.Dec{}, false
	}

	prec := TickPrecision(tickPrec)
	lowestTickIdx := prec.TickToIndex(prec.LowestTick())
	highestTickIdx := prec.TickToIndex(prec.HighestTick())
	var i, j int
	i, found = findFirstTrueCondition(lowestTickIdx, highestTickIdx, func(i int) bool {
		return ov.BuyAmountOver(prec.TickFromIndex(i+1), true).LTE(ov.SellAmountUnder(prec.TickFromIndex(i), true))
	})
	if !found {
		return sdk.Dec{}, false
	}
	j, found = findFirstTrueCondition(highestTickIdx, lowestTickIdx, func(i int) bool {
		return ov.BuyAmountOver(prec.TickFromIndex(i), true).GTE(ov.SellAmountUnder(prec.TickFromIndex(i-1), true))
	})
	if !found {
		return sdk.Dec{}, false
	}
	midTick := TickFromIndex(i, tickPrec).Add(TickFromIndex(j, tickPrec)).QuoInt64(2)
	return RoundPrice(midTick, tickPrec), true
}

// MatchAtSinglePrice matches all matchable orders(buy orders with higher(or equal) price
// than the price and sell orders with lower(or equal) price than the price)
// at the price.
func (ob *OrderBook) MatchAtSinglePrice(matchPrice sdk.Dec) (quoteCoinDiff sdk.Int, matched bool) {
	// TODO: use OrderBookView to optimize?
	buySums := make([]sdk.Int, 0, len(ob.buys.ticks))
	for i, buyTick := range ob.buys.ticks {
		if buyTick.price.LT(matchPrice) {
			break
		}
		sum := TotalOpenAmount(buyTick.orders())
		if i > 0 {
			sum = buySums[i-1].Add(sum)
		}
		buySums = append(buySums, sum)
	}
	if len(buySums) == 0 {
		return sdk.Int{}, false
	}
	sellSums := make([]sdk.Int, 0, len(ob.sells.ticks))
	for i, sellTick := range ob.sells.ticks {
		if sellTick.price.GT(matchPrice) {
			break
		}
		sum := TotalOpenAmount(sellTick.orders())
		if i > 0 {
			sum = sellSums[i-1].Add(sum)
		}
		sellSums = append(sellSums, sum)
	}
	if len(sellSums) == 0 {
		return sdk.Int{}, false
	}
	matchAmt := sdk.MinInt(buySums[len(buySums)-1], sellSums[len(sellSums)-1])
	bi := sort.Search(len(buySums), func(i int) bool {
		return buySums[i].GTE(matchAmt)
	})
	si := sort.Search(len(sellSums), func(i int) bool {
		return sellSums[i].GTE(matchAmt)
	})
	quoteCoinDiff = sdk.ZeroInt()
	distributeAmtToTicks := func(ticks []*orderBookTick, sums []sdk.Int, lastIdx int) {
		for _, tick := range ticks[:lastIdx] {
			quoteCoinDiff = quoteCoinDiff.Add(FulfillOrders(tick.orders(), matchPrice))
		}
		var remainingAmt sdk.Int
		if lastIdx == 0 {
			remainingAmt = matchAmt
		} else {
			remainingAmt = matchAmt.Sub(sums[lastIdx-1])
		}
		quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToTick(ticks[lastIdx], remainingAmt, matchPrice))
	}
	distributeAmtToTicks(ob.buys.ticks, buySums, bi)
	distributeAmtToTicks(ob.sells.ticks, sellSums, si)
	return quoteCoinDiff, true
}

// PriceDirection returns the estimated price direction within this batch
// considering the last price.
func (ob *OrderBook) PriceDirection(lastPrice sdk.Dec) PriceDirection {
	// TODO: use OrderBookView
	buyAmtOverLastPrice := sdk.ZeroInt()
	buyAmtAtLastPrice := sdk.ZeroInt()
	for _, tick := range ob.buys.ticks {
		if tick.price.LT(lastPrice) {
			break
		}
		amt := TotalOpenAmount(tick.orders())
		if tick.price.Equal(lastPrice) {
			buyAmtAtLastPrice = amt
			break
		}
		buyAmtOverLastPrice = buyAmtOverLastPrice.Add(amt)
	}
	sellAmtUnderLastPrice := sdk.ZeroInt()
	sellAmtAtLastPrice := sdk.ZeroInt()
	for _, tick := range ob.sells.ticks {
		if tick.price.GT(lastPrice) {
			break
		}
		amt := TotalOpenAmount(tick.orders())
		if tick.price.Equal(lastPrice) {
			sellAmtAtLastPrice = amt
			break
		}
		sellAmtUnderLastPrice = sellAmtUnderLastPrice.Add(amt)
	}
	switch {
	case buyAmtOverLastPrice.GT(sellAmtAtLastPrice.Add(sellAmtUnderLastPrice)):
		return PriceIncreasing
	case sellAmtUnderLastPrice.GT(buyAmtAtLastPrice.Add(buyAmtOverLastPrice)):
		return PriceDecreasing
	default:
		return PriceStaying
	}
}

// Match matches orders sequentially, starting from buy orders with the highest price
// and sell orders with the lowest price.
// The matching continues until there's no more matchable orders.
func (ob *OrderBook) Match(lastPrice sdk.Dec) (matchPrice sdk.Dec, quoteCoinDiff sdk.Int, matched bool) {
	if len(ob.buys.ticks) == 0 || len(ob.sells.ticks) == 0 {
		return sdk.Dec{}, sdk.Int{}, false
	}
	matchPrice = lastPrice
	dir := ob.PriceDirection(lastPrice)
	quoteCoinDiff, matched = ob.MatchAtSinglePrice(lastPrice)
	if dir == PriceStaying {
		return matchPrice, quoteCoinDiff, matched
	}
	if !matched {
		quoteCoinDiff = sdk.ZeroInt()
	}
	bi, si := 0, 0
	for bi < len(ob.buys.ticks) && si < len(ob.sells.ticks) && ob.buys.ticks[bi].price.GTE(ob.sells.ticks[si].price) {
		buyTick := ob.buys.ticks[bi]
		sellTick := ob.sells.ticks[si]
		switch dir {
		case PriceIncreasing:
			matchPrice = sellTick.price
		case PriceDecreasing:
			matchPrice = buyTick.price
		}
		buyTickOpenAmt := TotalOpenAmount(buyTick.orders())
		sellTickOpenAmt := TotalOpenAmount(sellTick.orders())
		if buyTickOpenAmt.LTE(sellTickOpenAmt) {
			quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToTick(buyTick, buyTickOpenAmt, matchPrice))
			bi++
		} else {
			quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToTick(buyTick, sellTickOpenAmt, matchPrice))
		}
		if sellTickOpenAmt.LTE(buyTickOpenAmt) {
			quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToTick(sellTick, sellTickOpenAmt, matchPrice))
			si++
		} else {
			quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToTick(sellTick, buyTickOpenAmt, matchPrice))
		}
		matched = true
	}
	return
}

// DistributeOrderAmountToTick distributes the given order amount to the orders
// at the tick.
// Orders with higher priority(have lower batch id) get matched first,
// then the remaining amount is distributed to the remaining orders.
func DistributeOrderAmountToTick(tick *orderBookTick, amt sdk.Int, price sdk.Dec) (quoteCoinDiff sdk.Int) {
	remainingAmt := amt
	quoteCoinDiff = sdk.ZeroInt()
	for _, group := range tick.orderGroups {
		openAmt := TotalOpenAmount(group.orders)
		if openAmt.IsZero() {
			continue
		}
		if remainingAmt.GTE(openAmt) {
			quoteCoinDiff = quoteCoinDiff.Add(FulfillOrders(group.orders, price))
			remainingAmt = remainingAmt.Sub(openAmt)
		} else {
			quoteCoinDiff = quoteCoinDiff.Add(DistributeOrderAmountToOrders(group.orders, remainingAmt, price))
			remainingAmt = sdk.ZeroInt()
		}
		if remainingAmt.IsZero() {
			break
		}
	}
	return
}

// DistributeOrderAmountToOrders distributes the given order amount to the orders
// proportional to each order's amount.
// After distributing the amount based on each order's proportion,
// remaining amount due to the decimal truncation is distributed
// to the orders again, by priority.
// This time, the proportion is not considered and each order takes up
// the amount as much as possible.
func DistributeOrderAmountToOrders(orders []Order, amt sdk.Int, price sdk.Dec) (quoteCoinDiff sdk.Int) {
	totalAmt := TotalAmount(orders)
	totalMatchedAmt := sdk.ZeroInt()
	matchedAmtByOrder := map[Order]sdk.Int{}

	for _, order := range orders {
		if order.GetOpenAmount().IsZero() {
			continue
		}
		orderAmt := order.GetAmount().ToDec()
		proportion := orderAmt.QuoTruncate(totalAmt.ToDec())
		matchedAmt := sdk.MinInt(order.GetOpenAmount(), proportion.MulInt(amt).TruncateInt())
		if matchedAmt.IsPositive() {
			matchedAmtByOrder[order] = matchedAmt
			totalMatchedAmt = totalMatchedAmt.Add(matchedAmt)
		}
	}

	remainingAmt := amt.Sub(totalMatchedAmt)
	for _, order := range orders {
		if remainingAmt.IsZero() {
			break
		}
		prevMatchedAmt, ok := matchedAmtByOrder[order]
		if !ok { // TODO: is it possible?
			prevMatchedAmt = sdk.ZeroInt()
		}
		matchedAmt := sdk.MinInt(remainingAmt, order.GetOpenAmount().Sub(prevMatchedAmt))
		matchedAmtByOrder[order] = prevMatchedAmt.Add(matchedAmt)
		remainingAmt = remainingAmt.Sub(matchedAmt)
	}

	var matchedOrders, notMatchedOrders []Order
	for _, order := range orders {
		matchedAmt, ok := matchedAmtByOrder[order]
		if !ok {
			matchedAmt = sdk.ZeroInt()
		}
		if !matchedAmt.IsZero() && (order.GetDirection() == Buy || price.MulInt(matchedAmt).TruncateInt().IsPositive()) {
			matchedOrders = append(matchedOrders, order)
		} else {
			notMatchedOrders = append(notMatchedOrders, order)
		}
	}

	if len(notMatchedOrders) > 0 {
		if len(matchedOrders) == 0 {
			return DistributeOrderAmountToOrders(orders[:len(orders)-1], amt, price)
		} else {
			return DistributeOrderAmountToOrders(matchedOrders, amt, price)
		}
	}

	quoteCoinDiff = sdk.ZeroInt()
	for order, matchedAmt := range matchedAmtByOrder {
		quoteCoinDiff = quoteCoinDiff.Add(FillOrder(order, matchedAmt, price))
	}
	return
}
