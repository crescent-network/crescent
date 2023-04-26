package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, qty, openQty, remainingDeposit sdk.Int) Order {
	return Order{
		Id:               orderId,
		Orderer:          ordererAddr.String(),
		MarketId:         marketId,
		IsBuy:            isBuy,
		Price:            price,
		Quantity:         qty,
		OpenQuantity:     openQty,
		RemainingDeposit: remainingDeposit,
	}
}

func (order Order) Validate() error {
	if order.Id == 0 {
		return fmt.Errorf("id must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(order.Orderer); err != nil {
		return fmt.Errorf("invalid orderer address: %w", err)
	}
	if order.MarketId == 0 {
		return fmt.Errorf("market id must not be 0")
	}
	if !order.Price.IsPositive() {
		return fmt.Errorf("price must be positive: %s", order.Price)
	}
	if !order.Quantity.IsPositive() {
		return fmt.Errorf("quantity must be positive: %s", order.Quantity)
	}
	if order.OpenQuantity.IsNegative() {
		return fmt.Errorf("open quantity must not be negative: %s", order.OpenQuantity)
	}
	if order.OpenQuantity.GT(order.Quantity) {
		return fmt.Errorf("open quantity must be smaller than quantity: %s > %s", order.OpenQuantity, order.Quantity)
	}
	if !order.RemainingDeposit.IsPositive() {
		return fmt.Errorf("remaining deposit must be positive: %s", order.RemainingDeposit)
	}
	return nil
}

func NewTransientOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, qty, openQty, remainingDeposit sdk.Int, isTemporary bool) TransientOrder {
	return TransientOrder{
		Order: Order{
			Id:               orderId,
			Orderer:          ordererAddr.String(),
			MarketId:         marketId,
			IsBuy:            isBuy,
			Price:            price,
			Quantity:         qty,
			OpenQuantity:     openQty,
			RemainingDeposit: remainingDeposit,
		},
		Updated:     false,
		IsTemporary: isTemporary,
	}
}

func NewTransientOrderFromOrder(order Order) TransientOrder {
	return TransientOrder{
		Order:       order,
		Updated:     false,
		IsTemporary: false,
	}
}

func (order TransientOrder) ExecutableQuantity() sdk.Int {
	if order.Order.IsBuy {
		return utils.MinInt(
			order.Order.OpenQuantity,
			order.Order.RemainingDeposit.ToDec().QuoTruncate(order.Order.Price).TruncateInt())
	}
	return utils.MinInt(order.Order.OpenQuantity, order.Order.RemainingDeposit)
}
