package amm

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FindMatchPrice returns the best match price for given order sources.
// If there is no matchable orders, found will be false.
func FindMatchPrice(os OrderSource, tickPrec int) (matchPrice sdk.Dec, found bool) {
	highestBuyPrice, found := os.HighestBuyPrice()
	if !found {
		return sdk.Dec{}, false
	}
	lowestSellPrice, found := os.LowestSellPrice()
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
		return os.BuyAmountOver(prec.TickFromIndex(i + 1)).LTE(os.SellAmountUnder(prec.TickFromIndex(i)))
	})
	if !found {
		return sdk.Dec{}, false
	}
	j, found = findFirstTrueCondition(highestTickIdx, lowestTickIdx, func(i int) bool {
		return os.BuyAmountOver(prec.TickFromIndex(i)).GTE(os.SellAmountUnder(prec.TickFromIndex(i - 1)))
	})
	if !found {
		return sdk.Dec{}, false
	}
	midTick := TickFromIndex(i, tickPrec).Add(TickFromIndex(j, tickPrec)).QuoInt64(2)
	return RoundPrice(midTick, tickPrec), true
}

// findFirstTrueCondition uses the binary search to find the first index
// where f(i) is true, while searching in range [start, end].
// It assumes that f(j) == false where j < i and f(j) == true where j >= i.
// start can be greater than end.
func findFirstTrueCondition(start, end int, f func(i int) bool) (i int, found bool) {
	if start < end {
		i = start + sort.Search(end-start+1, func(i int) bool {
			return f(start + i)
		})
		if i > end {
			return 0, false
		}
		return i, true
	}
	i = start - sort.Search(start-end+1, func(i int) bool {
		return f(start - i)
	})
	if i < end {
		return 0, false
	}
	return i, true
}

// FindLastMatchableOrders returns the last matchable order indexes for
// each buy/sell side.
// lastBuyPartialMatchAmt and lastSellPartialMatchAmt are
// the amount of partially matched portion of the last orders.
// FindLastMatchableOrders drops(ignores) an order if the orderer
// receives zero demand coin after truncation when the order is either
// fully matched or partially matched.
func FindLastMatchableOrders(buyOrders, sellOrders []Order, matchPrice sdk.Dec) (lastBuyIdx, lastSellIdx int, lastBuyPartialMatchAmt, lastSellPartialMatchAmt sdk.Int, found bool) {
	if len(buyOrders) == 0 || len(sellOrders) == 0 {
		return 0, 0, sdk.Int{}, sdk.Int{}, false
	}
	type Side struct {
		orders          []Order
		totalOpenAmt    sdk.Int
		i               int
		partialMatchAmt sdk.Int
	}
	buySide := &Side{buyOrders, TotalOpenAmount(buyOrders), len(buyOrders) - 1, sdk.Int{}}
	sellSide := &Side{sellOrders, TotalOpenAmount(sellOrders), len(sellOrders) - 1, sdk.Int{}}
	sides := map[OrderDirection]*Side{
		Buy:  buySide,
		Sell: sellSide,
	}
	// Repeatedly check both buy/sell side to see if there is an order to drop.
	// If there is not, then the loop is finished.
	for {
		ok := true
		for _, dir := range []OrderDirection{Buy, Sell} {
			side := sides[dir]
			i := side.i
			order := side.orders[i]
			matchAmt := sdk.MinInt(buySide.totalOpenAmt, sellSide.totalOpenAmt)
			otherOrdersAmt := side.totalOpenAmt.Sub(order.GetOpenAmount())
			// side.partialMatchAmt can be negative at this moment, but
			// FindLastMatchableOrders won't return a negative amount because
			// the if-block below would set ok = false if otherOrdersAmt >= matchAmt
			// and the loop would be continued.
			side.partialMatchAmt = matchAmt.Sub(otherOrdersAmt)
			if otherOrdersAmt.GTE(matchAmt) ||
				(dir == Sell && matchPrice.MulInt(side.partialMatchAmt).TruncateInt().IsZero()) {
				if i == 0 { // There's no orders left, which means orders are not matchable.
					return 0, 0, sdk.Int{}, sdk.Int{}, false
				}
				side.totalOpenAmt = side.totalOpenAmt.Sub(order.GetOpenAmount())
				side.i--
				ok = false
			}
		}
		if ok {
			return buySide.i, sellSide.i, buySide.partialMatchAmt, sellSide.partialMatchAmt, true
		}
	}
}

// MatchOrders matches orders at given matchPrice if matchable.
// Note that MatchOrders modifies the orders in the parameters.
// quoteCoinDust is the difference between total paid quote coin and total
// received quote coin.
// quoteCoinDust can be positive because of the decimal truncation.
func MatchOrders(buyOrders, sellOrders []Order, matchPrice sdk.Dec) (quoteCoinDust sdk.Int, matched bool) {
	buyOrders = DropSmallOrders(buyOrders, matchPrice)
	sellOrders = DropSmallOrders(sellOrders, matchPrice)

	bi, si, pmb, pms, found := FindLastMatchableOrders(buyOrders, sellOrders, matchPrice)
	if !found {
		return sdk.Int{}, false
	}

	quoteCoinDust = sdk.ZeroInt()

	for i := 0; i <= bi; i++ {
		buyOrder := buyOrders[i]
		var receivedBaseCoinAmt sdk.Int
		if i < bi {
			receivedBaseCoinAmt = buyOrder.GetOpenAmount()
		} else {
			receivedBaseCoinAmt = pmb
		}
		paidQuoteCoinAmt := matchPrice.MulInt(receivedBaseCoinAmt).Ceil().TruncateInt()
		buyOrder.SetOpenAmount(buyOrder.GetOpenAmount().Sub(receivedBaseCoinAmt))
		buyOrder.DecrRemainingOfferCoin(paidQuoteCoinAmt)
		buyOrder.IncrReceivedDemandCoin(receivedBaseCoinAmt)
		buyOrder.SetMatched(true)
		quoteCoinDust = quoteCoinDust.Add(paidQuoteCoinAmt)
	}

	for i := 0; i <= si; i++ {
		sellOrder := sellOrders[i]
		var paidBaseCoinAmt sdk.Int
		if i < si {
			paidBaseCoinAmt = sellOrder.GetOpenAmount()
		} else {
			paidBaseCoinAmt = pms
		}
		receivedQuoteCoinAmt := matchPrice.MulInt(paidBaseCoinAmt).TruncateInt()
		sellOrder.SetOpenAmount(sellOrder.GetOpenAmount().Sub(paidBaseCoinAmt))
		sellOrder.DecrRemainingOfferCoin(paidBaseCoinAmt)
		sellOrder.IncrReceivedDemandCoin(receivedQuoteCoinAmt)
		sellOrder.SetMatched(true)
		quoteCoinDust = quoteCoinDust.Sub(receivedQuoteCoinAmt)
	}

	return quoteCoinDust, true
}

// DropSmallOrders returns filtered orders, where orders with too small amount
// are dropped.
func DropSmallOrders(orders []Order, matchPrice sdk.Dec) []Order {
	var res []Order
	for _, order := range orders {
		openAmt := order.GetOpenAmount()
		if openAmt.GTE(MinCoinAmount) && matchPrice.MulInt(openAmt).TruncateInt().GTE(MinCoinAmount) {
			res = append(res, order)
		}
	}
	return res
}
