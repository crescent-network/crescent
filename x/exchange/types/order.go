package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSpotLimitOrder(
	orderId uint64, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty sdk.Int) SpotLimitOrder {
	return SpotLimitOrder{
		Id:           orderId,
		Orderer:      ordererAddr.String(),
		MarketId:     marketId,
		IsBuy:        isBuy,
		Price:        price,
		Quantity:     qty,
		OpenQuantity: qty,
	}
}
