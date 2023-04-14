package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func SqrtPriceAtTick(tick int32, prec int) sdk.Dec {
	return utils.DecApproxSqrt(exchangetypes.PriceAtTick(tick, prec))
}

func TickAtSqrtPrice(sqrtPrice sdk.Dec, prec int) int32 {
	return exchangetypes.TickAtPrice(sqrtPrice.Power(2), prec)
}

func NewTickInfo() TickInfo {
	return TickInfo{
		GrossLiquidity: utils.ZeroDec,
		NetLiquidity:   utils.ZeroDec,
	}
}
