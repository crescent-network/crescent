package amm

import (
	"math"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

	buyAmtOver := func(i int) sdk.Int {
		return os.BuyAmountOver(TickFromIndex(i, tickPrec))
	}
	sellAmtUnder := func(i int) sdk.Int {
		return os.SellAmountUnder(TickFromIndex(i, tickPrec))
	}

	lowestTickIdx := TickToIndex(LowestTick(tickPrec), tickPrec)
	highestTickIdx := TickToIndex(HighestTick(tickPrec), tickPrec)
	i, found := findFirstTrueCondition(lowestTickIdx, highestTickIdx, func(i int) bool {
		return buyAmtOver(i + 1).LTE(sellAmtUnder(i))
	})
	if !found {
		panic("impossible case")
	}
	j, found := findFirstTrueCondition(highestTickIdx, lowestTickIdx, func(i int) bool {
		return buyAmtOver(i).GTE(sellAmtUnder(i - 1))
	})
	if !found {
		panic("impossible case")
	}

	if i == j {
		return TickFromIndex(i, tickPrec), true
	} else {
		if int(math.Abs(float64(i-j))) > 1 { // sanity check
			panic("impossible case")
		}
		return TickFromIndex(RoundTickIndex(i), tickPrec), true
	}
}

func findFirstTrueCondition(start, end int, cb func(i int) bool) (int, bool) {
	if start < end {
		i := start + sort.Search(end-start+1, func(i int) bool {
			return cb(start + i)
		})
		return i, i <= end
	} else {
		i := start - sort.Search(start-end+1, func(i int) bool {
			return cb(start - i)
		})
		return i, i >= end
	}
}

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
	for {
		ok := true
		for dir, side := range sides {
			i := side.i
			order := side.orders[i]
			matchAmt := sdk.MinInt(buySide.totalOpenAmt, sellSide.totalOpenAmt)
			side.partialMatchAmt = matchAmt.Sub(side.totalOpenAmt.Sub(order.GetOpenAmount()))
			if side.totalOpenAmt.Sub(order.GetOpenAmount()).GT(matchAmt) ||
				(dir == Sell && matchPrice.MulInt(side.partialMatchAmt).TruncateInt().IsZero()) {
				if i == 0 {
					return
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

func MatchOrders(buyOrders, sellOrders []Order, matchPrice sdk.Dec) (quoteCoinDust sdk.Int, matched bool) {
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
