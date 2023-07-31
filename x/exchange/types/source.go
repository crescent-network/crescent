package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = MockOrderSource{}

type OrderSource interface {
	Name() string
	ConstructMemOrderBookSide(ctx sdk.Context, market Market, createOrder CreateOrderFunc, opts MemOrderBookSideOptions)
	AfterOrdersExecuted(ctx sdk.Context, market Market, ordererAddr sdk.AccAddress, results []*MemOrder)
}

type CreateOrderFunc func(ordererAddr sdk.AccAddress, price, qty sdk.Dec)

type MockOrderSource struct {
	name string
}

func NewMockOrderSource(name string) MockOrderSource {
	return MockOrderSource{name}
}

func (m MockOrderSource) Name() string {
	return m.name
}

func (MockOrderSource) ConstructMemOrderBookSide(sdk.Context, Market, CreateOrderFunc, MemOrderBookSideOptions) {
}

func (MockOrderSource) AfterOrdersExecuted(sdk.Context, Market, sdk.AccAddress, []*MemOrder) {
}
