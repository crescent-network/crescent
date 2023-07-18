package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderSource interface {
	Name() string
	GenerateOrders(ctx sdk.Context, market Market, createOrder CreateOrderFunc, opts GenerateOrdersOptions)
	AfterOrdersExecuted(ctx sdk.Context, market Market, results []TempOrder)
}

type CreateOrderFunc func(ordererAddr sdk.AccAddress, price, qty sdk.Dec) error

type GenerateOrdersOptions struct {
	IsBuy             bool
	PriceLimit        *sdk.Dec
	QuantityLimit     *sdk.Dec
	QuoteLimit        *sdk.Dec
	MaxNumPriceLevels int
}

func GroupTempOrderResultsByOrderer(results []TempOrder) (orderers []string, m map[string][]TempOrder) {
	m = map[string][]TempOrder{}
	for _, result := range results {
		if _, ok := m[result.Orderer]; !ok {
			orderers = append(orderers, result.Orderer)
		}
		m[result.Orderer] = append(m[result.Orderer], result)
	}
	return
}
