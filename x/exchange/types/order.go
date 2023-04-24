package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func NewSpotOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, qty, openQty, remainingDeposit sdk.Int) SpotOrder {
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

func NewTransientSpotOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, qty, openQty, remainingDeposit sdk.Int, isTemporary bool) TransientSpotOrder {
	return TransientSpotOrder{
		Order: SpotOrder{
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

func NewTransientSpotOrderFromSpotOrder(order SpotOrder) TransientSpotOrder {
	return TransientSpotOrder{
		Order:       order,
		Updated:     false,
		IsTemporary: false,
	}
}

func (order TransientSpotOrder) ExecutableQuantity() sdk.Int {
	if order.Order.IsBuy {
		return utils.MinInt(
			order.Order.OpenQuantity,
			order.Order.RemainingDeposit.ToDec().QuoTruncate(order.Order.Price).TruncateInt())
	}
	return order.Order.RemainingDeposit
}
