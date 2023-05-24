package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, qty sdk.Int, msgHeight int64,
	openQty, remainingDeposit sdk.Int, deadline time.Time) Order {
	return Order{
		Id:               orderId,
		Orderer:          ordererAddr.String(),
		MarketId:         marketId,
		IsBuy:            isBuy,
		Price:            price,
		Quantity:         qty,
		MsgHeight:        msgHeight,
		OpenQuantity:     openQty,
		RemainingDeposit: remainingDeposit,
		Deadline:         deadline,
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

func (order Order) ExecutableQuantity(price sdk.Dec) sdk.Int {
	if order.IsBuy {
		return utils.MinInt(
			order.OpenQuantity,
			order.RemainingDeposit.ToDec().QuoTruncate(price).TruncateInt())
	}
	return utils.MinInt(order.OpenQuantity, order.RemainingDeposit)
}
