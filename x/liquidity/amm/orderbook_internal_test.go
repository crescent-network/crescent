package amm

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	squad "github.com/cosmosquad-labs/squad/types"
)

// Copied from orderbook_test.go
func newOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) *BaseOrder {
	var offerCoinDenom, demandCoinDenom string
	switch dir {
	case Buy:
		offerCoinDenom, demandCoinDenom = "denom2", "denom1"
	case Sell:
		offerCoinDenom, demandCoinDenom = "denom1", "denom2"
	}
	return NewBaseOrder(dir, price, amt, sdk.NewCoin(offerCoinDenom, OfferCoinAmount(dir, price, amt)), demandCoinDenom)
}

func TestOrderBookTicks_add(t *testing.T) {
	prices := []sdk.Dec{
		squad.ParseDec("1.0"),
		squad.ParseDec("1.1"),
		squad.ParseDec("1.05"),
		squad.ParseDec("1.1"),
		squad.ParseDec("1.2"),
		squad.ParseDec("0.9"),
		squad.ParseDec("0.9"),
	}
	var ticks orderBookTicks
	for _, price := range prices {
		ticks.add(newOrder(Buy, price, sdk.NewInt(10000)))
	}
	pricesSet := map[string]struct{}{}
	for _, price := range prices {
		pricesSet[price.String()] = struct{}{}
	}
	prices = nil
	for priceStr := range pricesSet {
		prices = append(prices, squad.ParseDec(priceStr))
	}
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].GT(prices[j])
	})
	for i, price := range prices {
		require.True(sdk.DecEq(t, price, ticks[i].price))
	}
}
