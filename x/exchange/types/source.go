package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderSource interface {
	Name() string
	GenerateOrders(ctx sdk.Context, market Market, createOrder CreateOrderFunc, opts GenerateOrdersOptions)
	AfterOrdersExecuted(ctx sdk.Context, market Market, ordererAddr sdk.AccAddress, results []*MemOrder)
}

type CreateOrderFunc func(ordererAddr sdk.AccAddress, price, qty sdk.Dec) error

type GenerateOrdersOptions struct {
	IsBuy             bool
	PriceLimit        *sdk.Dec
	QuantityLimit     *sdk.Dec
	QuoteLimit        *sdk.Dec
	MaxNumPriceLevels int
}
