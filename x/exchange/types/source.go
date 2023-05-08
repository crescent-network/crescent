package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TemporaryOrderSource interface {
	ModuleName() string
	GenerateOrders(ctx sdk.Context, market Market, cb TemporaryOrderCallback, opts TemporaryOrderOptions)
	AfterOrdersExecuted(ctx sdk.Context, market Market, ordererAddr sdk.AccAddress, results []TemporaryOrderResult)
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
