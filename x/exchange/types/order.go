package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewOrder(
	orderId uint64, typ OrderType, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price, qty sdk.Dec, msgHeight int64,
	openQty, remainingDeposit sdk.Dec, deadline time.Time) Order {
	return Order{
		Id:               orderId,
		Type:             typ,
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
	if order.Type != OrderTypeLimit && order.Type != OrderTypeMM {
		return fmt.Errorf("invalid order type: %v", order.Type)
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
	if _, valid := ValidateTickPrice(order.Price); !valid {
		return fmt.Errorf("invalid tick price: %s", order.Price)
	}
	if !order.Quantity.IsPositive() {
		return fmt.Errorf("quantity must be positive: %s", order.Quantity)
	}
	if order.OpenQuantity.IsNegative() {
		return fmt.Errorf("open quantity must not be negative: %s", order.OpenQuantity)
	}
	if order.OpenQuantity.GT(order.Quantity) {
		return fmt.Errorf("open quantity must not be greater than quantity: %s > %s", order.OpenQuantity, order.Quantity)
	}
	if !order.RemainingDeposit.IsPositive() {
		return fmt.Errorf("remaining deposit must be positive: %s", order.RemainingDeposit)
	}
	return nil
}

func (order Order) ExecutableQuantity() sdk.Dec {
	if order.IsBuy {
		return sdk.MinDec(
			order.OpenQuantity,
			order.RemainingDeposit.QuoTruncate(order.Price))
	}
	return sdk.MinDec(order.OpenQuantity, order.RemainingDeposit)
}

func (order Order) MustGetOrdererAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(order.Orderer)
}

type ExecuteOrderResult struct {
	LastPrice        sdk.Dec
	ExecutedQuantity sdk.Dec
	ExecutedQuote    sdk.Dec
	Paid             sdk.DecCoin
	Received         sdk.DecCoin
	Fee              sdk.DecCoin
	FullyExecuted    bool
}

func NewExecuteOrderResult(payDenom, receiveDenom string) ExecuteOrderResult {
	return ExecuteOrderResult{
		ExecutedQuantity: utils.ZeroDec,
		ExecutedQuote:    utils.ZeroDec,
		Paid:             sdk.NewDecCoin(payDenom, utils.ZeroInt),
		Received:         sdk.NewDecCoin(receiveDenom, utils.ZeroInt),
		Fee:              sdk.NewDecCoin(receiveDenom, utils.ZeroInt), // always taker
	}
}

func (res ExecuteOrderResult) Executed() bool {
	return !res.LastPrice.IsNil()
}
