package amm

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ OrderView = (*OrderBookView)(nil)
	_ OrderView = Pool(nil)
	_ OrderView = MultipleOrderViews(nil)
)

type OrderView interface {
	HighestBuyPrice() (sdk.Dec, bool)
	LowestSellPrice() (sdk.Dec, bool)
	BuyAmountOver(price sdk.Dec, inclusive bool) sdk.Int
	SellAmountUnder(price sdk.Dec, inclusive bool) sdk.Int
}

type OrderBookView struct {
	buyAmtAccSums, sellAmtAccSums []amtAccSum
}

func (ob *OrderBook) MakeView() *OrderBookView {
	view := &OrderBookView{
		buyAmtAccSums:  make([]amtAccSum, len(ob.buys.ticks)),
		sellAmtAccSums: make([]amtAccSum, len(ob.sells.ticks)),
	}
	for i, tick := range ob.buys.ticks {
		var prevSum sdk.Int
		if i == 0 {
			prevSum = sdk.ZeroInt()
		} else {
			prevSum = view.buyAmtAccSums[i-1].sum
		}
		view.buyAmtAccSums[i] = amtAccSum{
			price: tick.price,
			sum:   prevSum.Add(TotalMatchableAmount(tick.orders(), tick.price)),
		}
	}
	for i, tick := range ob.sells.ticks {
		var prevSum sdk.Int
		if i == 0 {
			prevSum = sdk.ZeroInt()
		} else {
			prevSum = view.sellAmtAccSums[i-1].sum
		}
		view.sellAmtAccSums[i] = amtAccSum{
			price: tick.price,
			sum:   prevSum.Add(TotalMatchableAmount(tick.orders(), tick.price)),
		}
	}
	return view
}

func (view *OrderBookView) Match() {
	if len(view.buyAmtAccSums) == 0 || len(view.sellAmtAccSums) == 0 {
		return
	}
	buyIdx := sort.Search(len(view.buyAmtAccSums), func(i int) bool {
		return view.BuyAmountOver(view.buyAmtAccSums[i].price, true).GTE(
			view.SellAmountUnder(view.buyAmtAccSums[i].price, false))
	})
	if buyIdx >= len(view.buyAmtAccSums) { // not found
		buyIdx--
	}
	buyAmt := view.buyAmtAccSums[buyIdx].sum
	sellIdx := sort.Search(len(view.sellAmtAccSums), func(i int) bool {
		return view.SellAmountUnder(view.sellAmtAccSums[i].price, true).GTE(
			view.BuyAmountOver(view.sellAmtAccSums[i].price, false))
	})
	if sellIdx >= len(view.sellAmtAccSums) { // not found
		sellIdx--
	}
	sellAmt := view.sellAmtAccSums[sellIdx].sum
	matchAmt := sdk.MinInt(buyAmt, sellAmt)
	view.buyAmtAccSums = view.buyAmtAccSums[buyIdx:]
	if view.buyAmtAccSums[0].sum.Equal(matchAmt) {
		view.buyAmtAccSums = view.buyAmtAccSums[1:]
	}
	view.sellAmtAccSums = view.sellAmtAccSums[sellIdx:]
	if view.sellAmtAccSums[0].sum.Equal(matchAmt) {
		view.sellAmtAccSums = view.sellAmtAccSums[1:]
	}
	for i, accSum := range view.buyAmtAccSums {
		view.buyAmtAccSums[i].sum = accSum.sum.Sub(matchAmt)
	}
	for i, accSum := range view.sellAmtAccSums {
		view.sellAmtAccSums[i].sum = accSum.sum.Sub(matchAmt)
	}
}

func (view *OrderBookView) HighestBuyPrice() (sdk.Dec, bool) {
	if len(view.buyAmtAccSums) == 0 {
		return sdk.Dec{}, false
	}
	return view.buyAmtAccSums[0].price, true
}

func (view *OrderBookView) LowestSellPrice() (sdk.Dec, bool) {
	if len(view.sellAmtAccSums) == 0 {
		return sdk.Dec{}, false
	}
	return view.sellAmtAccSums[0].price, true
}

func (view *OrderBookView) BuyAmountOver(price sdk.Dec, inclusive bool) sdk.Int {
	i := sort.Search(len(view.buyAmtAccSums), func(i int) bool {
		if inclusive {
			return view.buyAmtAccSums[i].price.LT(price)
		} else {
			return view.buyAmtAccSums[i].price.LTE(price)
		}
	})
	if i == 0 {
		return sdk.ZeroInt()
	}
	return view.buyAmtAccSums[i-1].sum
}

func (view *OrderBookView) SellAmountUnder(price sdk.Dec, inclusive bool) sdk.Int {
	i := sort.Search(len(view.sellAmtAccSums), func(i int) bool {
		if inclusive {
			return view.sellAmtAccSums[i].price.GT(price)
		} else {
			return view.sellAmtAccSums[i].price.GTE(price)
		}
	})
	if i == 0 {
		return sdk.ZeroInt()
	}
	return view.sellAmtAccSums[i-1].sum
}

type amtAccSum struct {
	price sdk.Dec
	sum   sdk.Int
}

type MultipleOrderViews []OrderView

func MergeOrderViews(views ...OrderView) MultipleOrderViews {
	return MultipleOrderViews(views)
}

func (views MultipleOrderViews) HighestBuyPrice() (price sdk.Dec, found bool) {
	for _, view := range views {
		p, f := view.HighestBuyPrice()
		if f && (price.IsNil() || p.GT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (views MultipleOrderViews) LowestSellPrice() (price sdk.Dec, found bool) {
	for _, view := range views {
		p, f := view.LowestSellPrice()
		if f && (price.IsNil() || p.LT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (views MultipleOrderViews) BuyAmountOver(price sdk.Dec, inclusive bool) sdk.Int {
	amt := sdk.ZeroInt()
	for _, view := range views {
		amt = amt.Add(view.BuyAmountOver(price, inclusive))
	}
	return amt
}

func (views MultipleOrderViews) SellAmountUnder(price sdk.Dec, inclusive bool) sdk.Int {
	amt := sdk.ZeroInt()
	for _, view := range views {
		amt = amt.Add(view.SellAmountUnder(price, inclusive))
	}
	return amt
}
