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

func (engine *MatchEngine) Match(lastPrice sdk.Dec) (orderBook *OrderBook, swapPrice sdk.Dec, matched bool) {
	if !engine.Matchable() {
		return
	}

	swapPrice = engine.FindMatchPrice(lastPrice)
	buyPrice, _ := engine.BuyOrderSource.HighestTick()
	sellPrice, _ := engine.SellOrderSource.LowestTick()

	orderBook = NewOrderBook(engine.TickPrecision)

	for {
		if buyPrice.LT(swapPrice) || sellPrice.GT(swapPrice) {
			break
		}

		buyOrders := orderBook.BuyTicks.Orders(buyPrice)
		if len(buyOrders) == 0 {
			orderBook.AddOrders(engine.BuyOrderSource.Orders(buyPrice)...)
			buyOrders = orderBook.BuyTicks.Orders(buyPrice)
		}
		sellOrders := orderBook.SellTicks.Orders(sellPrice)
		if len(sellOrders) == 0 {
			orderBook.AddOrders(engine.SellOrderSource.Orders(sellPrice)...)
			sellOrders = orderBook.SellTicks.Orders(sellPrice)
		}

		MatchOrders(buyOrders, sellOrders, swapPrice)
		matched = true

		if buyPrice.Equal(swapPrice) && sellPrice.Equal(swapPrice) {
			break
		}

		if buyOrders.OpenBaseCoinAmount().IsZero() {
			var found bool
			buyPrice, found = engine.BuyOrderSource.DownTickWithOrders(buyPrice)
			if !found {
				break
			}
		}
		if sellOrders.OpenBaseCoinAmount().IsZero() {
			var found bool
			sellPrice, found = engine.SellOrderSource.UpTickWithOrders(sellPrice)
			if !found {
				break
			}
		}
	}

	return
}

// MatchOrders matches two order groups at given price.
func MatchOrders(buyOrders, sellOrders Orders, price sdk.Dec) {
	buyAmount := buyOrders.OpenBaseCoinAmount()
	sellAmount := sellOrders.OpenBaseCoinAmount()

	if buyAmount.IsZero() || sellAmount.IsZero() {
		return
	}

	var smallerOrders, biggerOrders Orders
	var smallerAmount, biggerAmount sdk.Int
	if buyAmount.LTE(sellAmount) { // Note that we use LTE here.
		smallerOrders, biggerOrders = buyOrders, sellOrders
		smallerAmount, biggerAmount = buyAmount, sellAmount
	} else {
		smallerOrders, biggerOrders = sellOrders, buyOrders
		smallerAmount, biggerAmount = sellAmount, buyAmount
	}

	for _, order := range smallerOrders {
		proportion := order.GetOpenBaseCoinAmount().ToDec().QuoInt(smallerAmount)
		order.SetOpenBaseCoinAmount(sdk.ZeroInt())
		out := proportion.MulInt(smallerAmount)
		in := proportion.MulInt(biggerAmount).TruncateInt()
		order.SetReceivedAmount(order.GetReceivedAmount().Add(in))
	}

	for _, order := range biggerOrders {
		proportion := order.GetRemainingOfferCoinAmount().ToDec().QuoInt(biggerAmount)
		if matchAll {
			order.SetRemainingOfferCoinAmount(sdk.ZeroInt())
		} else {
			out := proportion.MulInt(smallerDemandAmount).TruncateInt()
			order.SetRemainingOfferCoinAmount(order.GetRemainingOfferCoinAmount().Sub(out))
		}
		in := proportion.MulInt(smallerAmount).TruncateInt()
		order.SetReceivedAmount(order.GetReceivedAmount().Add(in))
	}
}
