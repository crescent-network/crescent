package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TemporaryOrderSource interface {
	Name() string
	GenerateOrders(ctx sdk.Context, market Market, cb TemporaryOrderCallback, opts TemporaryOrderOptions)
	AfterOrdersExecuted(ctx sdk.Context, market Market, results []TemporaryOrderResult)
}

type TemporaryOrderCallback func(ordererAddr sdk.AccAddress, price sdk.Dec, qty sdk.Int) error

type TemporaryOrderOptions struct {
	IsBuy         bool
	PriceLimit    *sdk.Dec
	QuantityLimit *sdk.Int
	QuoteLimit    *sdk.Int
}

type TemporaryOrderResult struct {
	Order            *Order
	ExecutedQuantity sdk.Int
	Paid             sdk.Coin
	Received         sdk.Coin
	Fee              sdk.Coin
}

func GroupTemporaryOrderResultsByOrderer(results []TemporaryOrderResult) (orderers []string, m map[string][]TemporaryOrderResult) {
	m = map[string][]TemporaryOrderResult{}
	for _, result := range results {
		if _, ok := m[result.Order.Orderer]; !ok {
			orderers = append(orderers, result.Order.Orderer)
		}
		m[result.Order.Orderer] = append(m[result.Order.Orderer], result)
	}
	return
}
