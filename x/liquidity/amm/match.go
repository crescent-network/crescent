package amm

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PriceDirection int

const (
	PriceStaying PriceDirection = iota + 1
	PriceIncreasing
	PriceDecreasing
)

func InitialMatchPrice(os OrderSource, tickPrec int) (matchPrice sdk.Dec, dir PriceDirection, matchable bool) {
	highest, found := os.HighestBuyPrice()
	if !found {
		return sdk.Dec{}, 0, false
	}
	lowest, found := os.LowestSellPrice()
	if !found {
		return sdk.Dec{}, 0, false
	}
	if highest.LT(lowest) {
		return sdk.Dec{}, 0, false
	}

	midPrice := highest.Add(lowest).QuoInt64(2)
	buyAmt := os.BuyAmountOver(midPrice)
	sellAmt := os.SellAmountUnder(midPrice)
	switch {
	case buyAmt.GT(sellAmt):
		dir = PriceIncreasing
	case sellAmt.GT(buyAmt):
		dir = PriceDecreasing
	default:
		dir = PriceStaying
	}

	switch dir {
	case PriceStaying:
		matchPrice = RoundPrice(midPrice, tickPrec)
	case PriceIncreasing:
		matchPrice = DownTick(midPrice, tickPrec)
	case PriceDecreasing:
		matchPrice = UpTick(midPrice, tickPrec)
	}

	return matchPrice, dir, true
}

func FindMatchPrice(os OrderSource, tickPrec int) (matchPrice sdk.Dec, found bool) {
	initialMatchPrice, dir, matchable := InitialMatchPrice(os, tickPrec)
	if !matchable {
		return sdk.Dec{}, false
	}
	if dir == PriceStaying {
		return initialMatchPrice, true
	}

	buyAmtOver := func(i int) sdk.Int {
		return os.BuyAmountOver(TickFromIndex(i, tickPrec))
	}
	sellAmtUnder := func(i int) sdk.Int {
		return os.SellAmountUnder(TickFromIndex(i, tickPrec))
	}

	switch dir {
	case PriceIncreasing:
		start := TickToIndex(initialMatchPrice, tickPrec)
		end := TickToIndex(HighestTick(tickPrec), tickPrec)
		i := start + sort.Search(end-start+1, func(i int) bool {
			i += start
			bg := buyAmtOver(i + 1)
			return bg.IsZero() || (bg.LTE(sellAmtUnder(i)) && buyAmtOver(i).GT(sellAmtUnder(i-1)))
		})
		if i > end {
			i = end
		}
		return TickFromIndex(i, tickPrec), true
	default: // PriceDecreasing
		start := TickToIndex(initialMatchPrice, tickPrec)
		end := TickToIndex(LowestTick(tickPrec), tickPrec)
		i := start - sort.Search(start-end+1, func(i int) bool {
			i = start - i
			sl := sellAmtUnder(i - 1)
			return sl.IsZero() || (buyAmtOver(i+1).LTE(sellAmtUnder(i)) && buyAmtOver(i).GTE(sl))
		})
		if i < end {
			i = end
		}
		return TickFromIndex(i, tickPrec), true
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
