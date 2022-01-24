package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PriceDirection int

const (
	PriceIncreasing PriceDirection = iota + 1
	PriceDecreasing
)

type MatchEngine struct {
	BuyOrderSource  OrderSource
	SellOrderSource OrderSource
	TickPrecision   int // price tick precision
}

func NewMatchEngine(buys, sells OrderSource, prec int) *MatchEngine {
	return &MatchEngine{
		BuyOrderSource:  buys,
		SellOrderSource: sells,
		TickPrecision:   prec,
	}
}

func (engine *MatchEngine) Matchable() bool {
	highestBuyPrice, found := engine.BuyOrderSource.HighestTick()
	if !found {
		return false
	}
	return engine.SellOrderSource.AmountLTE(highestBuyPrice).IsPositive()
}

func (engine *MatchEngine) EstimatedPriceDirection(lastPrice sdk.Dec) PriceDirection {
	buyAmount := engine.BuyOrderSource.AmountGTE(lastPrice)
	sellAmount := engine.SellOrderSource.AmountLTE(lastPrice)
	if buyAmount.ToDec().GTE(lastPrice.MulInt(sellAmount)) {
		return PriceIncreasing
	}
	return PriceDecreasing
}

func (engine *MatchEngine) FindMatchPrice(lastPrice sdk.Dec) sdk.Dec {
	dir := engine.EstimatedPriceDirection(lastPrice)
	tickSource := MergeOrderSources(engine.BuyOrderSource, engine.SellOrderSource) // temporary order source just for ticks

	buysCache := map[int]sdk.Int{}
	buyAmountGTE := func(i int) sdk.Int {
		ba, ok := buysCache[i]
		if !ok {
			ba = engine.BuyOrderSource.AmountGTE(TickFromIndex(i, engine.TickPrecision))
			buysCache[i] = ba
		}
		return ba
	}
	sellsCache := map[int]sdk.Int{}
	sellAmountLTE := func(i int) sdk.Int {
		sa, ok := sellsCache[i]
		if !ok {
			sa = engine.SellOrderSource.AmountLTE(TickFromIndex(i, engine.TickPrecision))
			sellsCache[i] = sa
		}
		return sa
	}

	// TODO: lastPrice could be a price not fit in ticks.
	currentPrice := lastPrice
	for {
		i := TickToIndex(currentPrice, engine.TickPrecision)

		if buyAmountGTE(i+1).LTE(sellAmountLTE(i)) && buyAmountGTE(i).GTE(sellAmountLTE(i-1)) {
			return currentPrice
		}

		var nextPrice sdk.Dec
		var found bool
		switch dir {
		case PriceIncreasing:
			if buyAmountGTE(i + 1).IsZero() {
				return currentPrice
			}
			nextPrice, found = tickSource.UpTick(currentPrice)
		case PriceDecreasing:
			if sellAmountLTE(i - 1).IsZero() {
				return currentPrice
			}
			nextPrice, found = tickSource.DownTick(currentPrice)
		}
		if !found {
			return currentPrice
		}
		currentPrice = nextPrice
	}
}

// TODO: no need to return the order book
func (engine *MatchEngine) Match(lastPrice sdk.Dec) (orderBook *OrderBook, matchPrice sdk.Dec, quoteCoinDustAmt sdk.Int, matched bool) {
	if !engine.Matchable() {
		return
	}

	matchPrice = engine.FindMatchPrice(lastPrice)
	buyPrice, _ := engine.BuyOrderSource.HighestTick()
	sellPrice, _ := engine.SellOrderSource.LowestTick()

	var buyOrders, sellOrders Orders

	orderBook = NewOrderBook(engine.TickPrecision)

	for buyPrice.GTE(matchPrice) {
		orders := engine.BuyOrderSource.Orders(buyPrice)
		orderBook.AddOrders(orders...)
		buyOrders = append(buyOrders, orders...)
		var found bool
		buyPrice, found = engine.BuyOrderSource.DownTickWithOrders(buyPrice)
		if !found {
			break
		}
	}

	for sellPrice.LTE(matchPrice) {
		orders := engine.SellOrderSource.Orders(sellPrice)
		orderBook.AddOrders(orders...)
		sellOrders = append(sellOrders, orders...)
		var found bool
		sellPrice, found = engine.SellOrderSource.UpTickWithOrders(sellPrice)
		if !found {
			break
		}
	}

	quoteCoinDustAmt, matched = MatchOrders(buyOrders, sellOrders, matchPrice)

	return
}

func FindLastMatchableOrders(orders1, orders2 Orders, matchPrice sdk.Dec) (idx1, idx2 int, partialMatchAmt1, partialMatchAmt2 sdk.Int, found bool) {
	sides := []*struct {
		orders          Orders
		totalAmt        sdk.Int
		i               int
		partialMatchAmt sdk.Int
	}{
		{orders1, orders1.OpenBaseCoinAmount(), len(orders1) - 1, sdk.Int{}},
		{orders2, orders2.OpenBaseCoinAmount(), len(orders2) - 1, sdk.Int{}},
	}
	for {
		ok := true
		for _, side := range sides {
			i := side.i
			order := side.orders[i]
			matchAmt := sdk.MinInt(sides[0].totalAmt, sides[1].totalAmt)
			side.partialMatchAmt = matchAmt.Sub(side.totalAmt.Sub(order.GetOpenBaseCoinAmount()))
			if side.totalAmt.Sub(order.GetOpenBaseCoinAmount()).GT(matchAmt) ||
				matchPrice.MulInt(side.partialMatchAmt).TruncateInt().IsZero() {
				if i == 0 {
					return
				}
				side.totalAmt = side.totalAmt.Sub(order.GetOpenBaseCoinAmount())
				side.i--
				ok = false
			}
		}
		if ok {
			return sides[0].i, sides[1].i, sides[0].partialMatchAmt, sides[1].partialMatchAmt, true
		}
	}
}

func MatchOrders(buyOrders, sellOrders Orders, matchPrice sdk.Dec) (quoteCoinDustAmt sdk.Int, matched bool) {
	buyOrders.Sort(DescendingPrice)
	sellOrders.Sort(AscendingPrice)

	bi, si, pmb, pms, found := FindLastMatchableOrders(buyOrders, sellOrders, matchPrice)
	if !found {
		return
	}

	quoteCoinDustAmt = sdk.ZeroInt()

	for i := 0; i <= bi; i++ {
		buyOrder := buyOrders[i]
		var receivedBaseCoinAmt sdk.Int
		if i < bi {
			receivedBaseCoinAmt = buyOrder.GetOpenBaseCoinAmount()
		} else {
			receivedBaseCoinAmt = pmb
		}
		paidQuoteCoinAmt := matchPrice.MulInt(receivedBaseCoinAmt).Ceil().TruncateInt()
		buyOrder.SetOpenBaseCoinAmount(buyOrder.GetOpenBaseCoinAmount().Sub(receivedBaseCoinAmt))
		buyOrder.SetRemainingOfferCoinAmount(buyOrder.GetRemainingOfferCoinAmount().Sub(paidQuoteCoinAmt))
		buyOrder.SetReceivedAmount(receivedBaseCoinAmt)
		quoteCoinDustAmt = quoteCoinDustAmt.Add(paidQuoteCoinAmt)
	}

	for i := 0; i <= si; i++ {
		sellOrder := sellOrders[i]
		var paidBaseCoinAmt sdk.Int
		if i < si {
			paidBaseCoinAmt = sellOrder.GetOpenBaseCoinAmount()
		} else {
			paidBaseCoinAmt = pms
		}
		receivedQuoteCoinAmt := matchPrice.MulInt(paidBaseCoinAmt).TruncateInt()
		sellOrder.SetOpenBaseCoinAmount(sellOrder.GetOpenBaseCoinAmount().Sub(paidBaseCoinAmt))
		sellOrder.SetRemainingOfferCoinAmount(sellOrder.GetRemainingOfferCoinAmount().Sub(paidBaseCoinAmt))
		sellOrder.SetReceivedAmount(receivedQuoteCoinAmt)
		quoteCoinDustAmt = quoteCoinDustAmt.Sub(receivedQuoteCoinAmt)
	}

	return
}
