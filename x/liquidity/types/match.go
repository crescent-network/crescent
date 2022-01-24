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

func (engine *MatchEngine) InitialMatchPrice() sdk.Dec {
	highestBuyPrice, found := engine.BuyOrderSource.HighestTick()
	if !found {
		panic("there is no buy orders")
	}
	lowestSellPrice, found := engine.SellOrderSource.LowestTick()
	if !found {
		panic("there is no sell orders")
	}
	return highestBuyPrice.Add(lowestSellPrice).QuoInt64(2)
}

func (engine *MatchEngine) FindMatchPrice() sdk.Dec {
	matchPrice := engine.InitialMatchPrice()
	dir := engine.EstimatedPriceDirection(matchPrice)

	matchPriceTick := PriceToTick(matchPrice, engine.TickPrecision)
	if dir == PriceIncreasing && !matchPrice.Equal(matchPriceTick) {
		matchPrice = UpTick(matchPriceTick, engine.TickPrecision)
	} else {
		matchPrice = matchPriceTick
	}

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

	for {
		i := TickToIndex(matchPrice, engine.TickPrecision)

		if buyAmountGTE(i+1).LTE(sellAmountLTE(i)) && buyAmountGTE(i).GTE(sellAmountLTE(i-1)) {
			return matchPrice
		}

		var nextPrice sdk.Dec
		var found bool
		switch dir {
		case PriceIncreasing:
			if buyAmountGTE(i + 1).IsZero() {
				return matchPrice
			}
			nextPrice, found = tickSource.UpTick(matchPrice)
		case PriceDecreasing:
			if sellAmountLTE(i - 1).IsZero() {
				return matchPrice
			}
			nextPrice, found = tickSource.DownTick(matchPrice)
		}
		if !found {
			return matchPrice
		}
		matchPrice = nextPrice
	}
}

// TODO: no need to return the order book
func (engine *MatchEngine) Match() (orderBook *OrderBook, matchPrice sdk.Dec, quoteCoinDustAmt sdk.Int, matched bool) {
	if !engine.Matchable() {
		return
	}

	matchPrice = engine.FindMatchPrice()
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
		{orders1, orders1.OpenAmount(), len(orders1) - 1, sdk.Int{}},
		{orders2, orders2.OpenAmount(), len(orders2) - 1, sdk.Int{}},
	}
	for {
		ok := true
		for _, side := range sides {
			i := side.i
			order := side.orders[i]
			matchAmt := sdk.MinInt(sides[0].totalAmt, sides[1].totalAmt)
			side.partialMatchAmt = matchAmt.Sub(side.totalAmt.Sub(order.GetOpenAmount()))
			if side.totalAmt.Sub(order.GetOpenAmount()).GT(matchAmt) ||
				matchPrice.MulInt(side.partialMatchAmt).TruncateInt().IsZero() {
				if i == 0 {
					return
				}
				side.totalAmt = side.totalAmt.Sub(order.GetOpenAmount())
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
			receivedBaseCoinAmt = buyOrder.GetOpenAmount()
		} else {
			receivedBaseCoinAmt = pmb
		}
		paidQuoteCoinAmt := matchPrice.MulInt(receivedBaseCoinAmt).Ceil().TruncateInt()
		buyOrder.SetOpenAmount(buyOrder.GetOpenAmount().Sub(receivedBaseCoinAmt))
		buyOrder.SetRemainingOfferCoinAmount(buyOrder.GetRemainingOfferCoinAmount().Sub(paidQuoteCoinAmt))
		buyOrder.SetReceivedAmount(receivedBaseCoinAmt)
		quoteCoinDustAmt = quoteCoinDustAmt.Add(paidQuoteCoinAmt)
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
		sellOrder.SetRemainingOfferCoinAmount(sellOrder.GetRemainingOfferCoinAmount().Sub(paidBaseCoinAmt))
		sellOrder.SetReceivedAmount(receivedQuoteCoinAmt)
		quoteCoinDustAmt = quoteCoinDustAmt.Sub(receivedQuoteCoinAmt)
	}

	return
}
