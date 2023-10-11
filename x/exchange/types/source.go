package types

import (
	"fmt"

	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

var _ OrderSource = MockOrderSource{}

type OrderSource interface {
	Name() string
	ConstructMemOrderBookSide(ctx sdk.Context, market Market, createOrder CreateOrderFunc, opts MemOrderBookSideOptions) error
	AfterOrdersExecuted(ctx sdk.Context, market Market, ordererAddr sdk.AccAddress, results []*MemOrder) error
}

type CreateOrderFunc func(ordererAddr sdk.AccAddress, price sdk.Dec, qty sdk.Int)

type MockOrderSource struct {
	name    string
	Address sdk.AccAddress
	Orders  []MockOrderSourceOrder
}

type MockOrderSourceOrder struct {
	IsBuy    bool
	Price    sdk.Dec
	Quantity sdk.Int
}

func NewMockOrderSourceOrder(
	isBuy bool, price sdk.Dec, qty sdk.Int) MockOrderSourceOrder {
	return MockOrderSourceOrder{
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
	}
}

func NewMockOrderSource(name string, orders ...MockOrderSourceOrder) MockOrderSource {
	return MockOrderSource{
		name:    name,
		Address: address.Module(ModuleName, []byte(fmt.Sprintf("MockOrderSource/%s", name))),
		Orders:  orders,
	}
}

func (os MockOrderSource) Name() string {
	return os.name
}

func (os MockOrderSource) ConstructMemOrderBookSide(
	_ sdk.Context, _ Market,
	createOrder CreateOrderFunc, opts MemOrderBookSideOptions) error {
	var orders []MockOrderSourceOrder
	for _, order := range os.Orders {
		if order.IsBuy == opts.IsBuy {
			orders = append(orders, order)
		}
	}
	slices.SortFunc(orders, func(a, b MockOrderSourceOrder) bool {
		if opts.IsBuy {
			return a.Price.GT(b.Price)
		}
		return a.Price.LT(b.Price)
	})
	accQty := utils.ZeroInt
	accQuote := utils.ZeroDec
	for _, order := range orders {
		if opts.ReachedLimit(order.Price, accQty, accQuote, 0) {
			break
		}
		createOrder(os.Address, order.Price, order.Quantity)
		accQty = accQty.Add(order.Quantity)
		accQuote = accQuote.Add(order.Price.MulInt(order.Quantity))
	}
	return nil
}

func (os MockOrderSource) AfterOrdersExecuted(
	sdk.Context, Market, sdk.AccAddress, []*MemOrder) error {
	// Do nothing
	return nil
}
