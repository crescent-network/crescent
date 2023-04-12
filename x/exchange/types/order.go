package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewSpotOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price *sdk.Dec, qty, openQty, remainingDeposit sdk.Int) SpotOrder {
	return SpotOrder{
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

func (order SpotOrder) ExecutableQuantity() sdk.Int {
	if order.Price == nil {
		// Orders without price are market orders, thus executable quantity
		// cannot be determined. Simply return zero for market orders.
		return utils.ZeroInt
	}
	if order.IsBuy {
		return utils.MinInt(
			order.OpenQuantity,
			order.RemainingDeposit.ToDec().QuoTruncate(*order.Price).TruncateInt())
	}
	return order.RemainingDeposit
}
