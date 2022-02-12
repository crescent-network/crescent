package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = (*mergedOrderSource)(nil)

type OrderView interface {
	HighestBuyPrice() (price sdk.Dec, found bool)
	LowestSellPrice() (price sdk.Dec, found bool)
	BuyAmountOver(price sdk.Dec) sdk.Int   // Includes the price
	SellAmountUnder(price sdk.Dec) sdk.Int // Includes the price
}

type OrderSource interface {
	OrderView
	BuyOrdersOver(price sdk.Dec) []Order   // Includes the price
	SellOrdersUnder(price sdk.Dec) []Order // Includes the price
}

type mergedOrderSource struct {
	sources []OrderSource
}

func MergeOrderSources(sources ...OrderSource) OrderSource {
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
