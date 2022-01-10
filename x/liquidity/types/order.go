package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = (*mergedOrderSource)(nil)

// OrderSource defines a source of orders which can be an order book or
// a pool.
// TODO: omit prec parameter?
type OrderSource interface {
	AmountGTE(price sdk.Dec) sdk.Int
	AmountLTE(price sdk.Dec) sdk.Int
	Orders(price sdk.Dec) []Order
	UpTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool)
	DownTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool)
	HighestTick(prec int) (tick sdk.Dec, found bool)
	LowestTick(prec int) (tick sdk.Dec, found bool)
}

type mergedOrderSource struct {
	sources []OrderSource
}

func (mos *mergedOrderSource) AmountGTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range mos.sources {
		amt = amt.Add(source.AmountGTE(price))
	}
	return amt
}

func (mos *mergedOrderSource) AmountLTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range mos.sources {
		amt = amt.Add(source.AmountLTE(price))
	}
	return amt
}

func (mos *mergedOrderSource) Orders(price sdk.Dec) []Order {
	var os []Order
	for _, source := range mos.sources {
		os = append(os, source.Orders(price)...)
	}
	return os
}

func (mos *mergedOrderSource) UpTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool) {
	for _, source := range mos.sources {
		t, f := source.UpTick(price, prec)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (mos *mergedOrderSource) DownTick(price sdk.Dec, prec int) (tick sdk.Dec, found bool) {
	for _, source := range mos.sources {
		t, f := source.DownTick(price, prec)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (mos *mergedOrderSource) HighestTick(prec int) (tick sdk.Dec, found bool) {
	for _, source := range mos.sources {
		t, f := source.HighestTick(prec)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (mos *mergedOrderSource) LowestTick(prec int) (tick sdk.Dec, found bool) {
	for _, source := range mos.sources {
		t, f := source.LowestTick(prec)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func MergeOrderSources(sources ...OrderSource) OrderSource {
	return &mergedOrderSource{sources: sources}
}

type MatchEngine struct {
	buys  OrderSource
	sells OrderSource
	prec  int // price tick precision
}

func NewMatchEngine(buys, sells OrderSource, prec int) *MatchEngine {
	return &MatchEngine{
		buys:  buys,
		sells: sells,
		prec:  prec,
	}
}

func (eng *MatchEngine) EstimatedPriceDirection(lastPrice sdk.Dec) PriceDirection {
	if eng.buys.AmountGTE(lastPrice).ToDec().GTE(lastPrice.MulInt(eng.sells.AmountLTE(lastPrice))) {
		return PriceIncreasing
	}
	return PriceDecreasing
}

// SwapPrice assumes that the last price is fit in tick.
func (eng *MatchEngine) SwapPrice(lastPrice sdk.Dec) sdk.Dec {
	dir := eng.EstimatedPriceDirection(lastPrice)
	os := MergeOrderSources(eng.buys, eng.sells) // temporary order source just for ticks

	buysCache := map[int]sdk.Int{}
	buyAmountGTE := func(i int) sdk.Int {
		ba, ok := buysCache[i]
		if !ok {
			ba = eng.buys.AmountGTE(TickFromIndex(i, eng.prec))
			buysCache[i] = ba
		}
		return ba
	}
	sellsCache := map[int]sdk.Int{}
	sellAmountLTE := func(i int) sdk.Int {
		sa, ok := sellsCache[i]
		if !ok {
			sa = eng.sells.AmountLTE(TickFromIndex(i, eng.prec))
			sellsCache[i] = sa
		}
		return sa
	}

	currentPrice := lastPrice
	for {
		i := TickToIndex(currentPrice, eng.prec)
		ba := buyAmountGTE(i)
		sa := sellAmountLTE(i)
		hba := buyAmountGTE(i + 1)
		lsa := sellAmountLTE(i - 1)

		if currentPrice.MulInt(sa).TruncateInt().GTE(hba) && ba.GTE(currentPrice.MulInt(lsa).TruncateInt()) {
			return currentPrice
		}

		var nextPrice sdk.Dec
		var found bool
		switch dir {
		case PriceIncreasing:
			nextPrice, found = os.UpTick(currentPrice, eng.prec)
		case PriceDecreasing:
			nextPrice, found = os.DownTick(currentPrice, eng.prec)
		}
		if !found {
			return currentPrice
		}
		currentPrice = nextPrice
	}
}

func (eng *MatchEngine) Match(lastPrice sdk.Dec) {

}
