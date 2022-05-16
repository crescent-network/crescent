package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ OrderView = (*mergedOrderView)(nil)
	_ OrderSource = (*mergedOrderSource)(nil)
)

// OrderView is the interface which provides a view of orders.
type OrderView interface {
	HighestBuyPrice() (price sdk.Dec, found bool)
	LowestSellPrice() (price sdk.Dec, found bool)
	BuyAmountOver(price sdk.Dec) sdk.Int   // Includes the price
	SellAmountUnder(price sdk.Dec) sdk.Int // Includes the price
}

// OrderSource is the interface which provides a view of orders and also
// provides a way to extract orders from it.
type OrderSource interface {
	OrderView
	BuyOrdersOver(price sdk.Dec) []Order   // Includes the price
	SellOrdersUnder(price sdk.Dec) []Order // Includes the price
}

// mergedOrderView is a merged order view of multiple views.
type mergedOrderView struct {
	views []OrderView
}

func MergeOrderViews(views ...OrderView) *mergedOrderView {
	return &mergedOrderView{views: views}
}

func (ov *mergedOrderView) HighestBuyPrice() (price sdk.Dec, found bool) {
	for _, source := range ov.views {
		p, f := source.HighestBuyPrice()
		if f && (price.IsNil() || p.GT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (ov *mergedOrderView) LowestSellPrice() (price sdk.Dec, found bool) {
	for _, source := range ov.views {
		p, f := source.LowestSellPrice()
		if f && (price.IsNil() || p.LT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (ov *mergedOrderView) BuyAmountOver(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range ov.views {
		amt = amt.Add(source.BuyAmountOver(price))
	}
	return amt
}

func (ov *mergedOrderView) SellAmountUnder(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range ov.views {
		amt = amt.Add(source.SellAmountUnder(price))
	}
	return amt
}

// mergedOrderSource is a merged order source of multiple sources.
type mergedOrderSource struct {
	sources []OrderSource
}

// MergeOrderSources returns a merged order source of multiple sources.
func MergeOrderSources(sources ...OrderSource) *mergedOrderSource {
	return &mergedOrderSource{sources: sources}
}

func (os *mergedOrderSource) HighestBuyPrice() (price sdk.Dec, found bool) {
	for _, source := range os.sources {
		p, f := source.HighestBuyPrice()
		if f && (price.IsNil() || p.GT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (os *mergedOrderSource) LowestSellPrice() (price sdk.Dec, found bool) {
	for _, source := range os.sources {
		p, f := source.LowestSellPrice()
		if f && (price.IsNil() || p.LT(price)) {
			price = p
			found = true
		}
	}
	return
}

func (os *mergedOrderSource) BuyAmountOver(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.sources {
		amt = amt.Add(source.BuyAmountOver(price))
	}
	return amt
}

func (os *mergedOrderSource) SellAmountUnder(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.sources {
		amt = amt.Add(source.SellAmountUnder(price))
	}
	return amt
}

func (os *mergedOrderSource) BuyOrdersOver(price sdk.Dec) []Order {
	var orders []Order
	for _, source := range os.sources {
		orders = append(orders, source.BuyOrdersOver(price)...)
	}
	return orders
}

func (os *mergedOrderSource) SellOrdersUnder(price sdk.Dec) []Order {
	var orders []Order
	for _, source := range os.sources {
		orders = append(orders, source.SellOrdersUnder(price)...)
	}
	return orders
}
