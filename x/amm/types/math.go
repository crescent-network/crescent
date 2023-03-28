package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func LiquidityForAmount0(sqrtPriceA, sqrtPriceB sdk.Dec, amt0 sdk.Int) sdk.Int {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	// TODO: possible precision loss?
	return sqrtPriceA.MulTruncate(sqrtPriceB).MulInt(amt0).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmount1(sqrtPriceA, sqrtPriceB sdk.Dec, amt1 sdk.Int) sdk.Int {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	return sdk.NewDecFromInt(amt1).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmounts(currentSqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amt0, amt1 sdk.Int) sdk.Int {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	if currentSqrtPrice.LTE(sqrtPriceA) {
		return LiquidityForAmount0(sqrtPriceA, sqrtPriceB, amt0)
	} else if sqrtPriceA.LT(sqrtPriceB) {
		liquidity0 := LiquidityForAmount0(currentSqrtPrice, sqrtPriceB, amt0)
		liquidity1 := LiquidityForAmount1(sqrtPriceA, currentSqrtPrice, amt1)
		return sdk.MinInt(liquidity0, liquidity1)
	}
	return LiquidityForAmount1(sqrtPriceA, sqrtPriceB, amt1)
}
