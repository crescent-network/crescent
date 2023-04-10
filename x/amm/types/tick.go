package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func SqrtPriceAtTick(tick int32, prec int) (sdk.Dec, error) {
	price := exchangetypes.PriceAtTick(tick, prec)
	return price.ApproxSqrt()
}

func TickAtSqrtPrice(sqrtPrice sdk.Dec, prec int) int32 {
	return exchangetypes.TickAtPrice(sqrtPrice.Power(2), prec)
}

func NewTickInfo() TickInfo {
	return TickInfo{
		GrossLiquidity: sdk.ZeroInt(),
		NetLiquidity:   sdk.ZeroInt(),
	}
}
