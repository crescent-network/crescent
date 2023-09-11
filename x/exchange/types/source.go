package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ OrderSource = MockOrderSource{}

type OrderSource interface {
	Name() string
	ConstructMemOrderBookSide(ctx sdk.Context, market Market, createOrder CreateOrderFunc, opts MemOrderBookSideOptions) error
	AfterOrdersExecuted(ctx sdk.Context, market Market, ordererAddr sdk.AccAddress, results []*MemOrder) error
}

type CreateOrderFunc func(ordererAddr sdk.AccAddress, price, qty, openQty sdk.Dec)

type MockOrderSource struct {
	name string
}

func NewMockOrderSource(name string) MockOrderSource {
	return MockOrderSource{name}
}

func (m MockOrderSource) Name() string {
	return m.name
}

func (MockOrderSource) ConstructMemOrderBookSide(sdk.Context, Market, CreateOrderFunc, MemOrderBookSideOptions) error {
	return nil
}

func (MockOrderSource) AfterOrdersExecuted(sdk.Context, Market, sdk.AccAddress, []*MemOrder) error {
	return nil
}
