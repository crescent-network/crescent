package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSpotLimitOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty, depositAmt sdk.Int) SpotLimitOrder {
	return SpotLimitOrder{
		Id:            orderId,
		Orderer:       ordererAddr.String(),
		MarketId:      marketId,
		IsBuy:         isBuy,
		Price:         price,
		Quantity:      qty,
		OpenQuantity:  qty, // TODO: replace this field with ExecutableQuantity?
		DepositAmount: depositAmt,
	}
}

func (order SpotLimitOrder) ExecutableQuantity() sdk.Int {
	if order.IsBuy {
		return order.DepositAmount.ToDec().QuoTruncate(order.Price).TruncateInt()
	}
	return order.DepositAmount
}
