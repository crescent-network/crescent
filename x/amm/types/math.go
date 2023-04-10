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

func amount0Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int, roundUp bool) sdk.Int {
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
	if roundUp {
		return intermediate.QuoRoundUp(sqrtPriceB).QuoRoundUp(sqrtPriceA).Ceil().TruncateInt()
	}
	return intermediate.QuoTruncate(sqrtPriceB).QuoTruncate(sqrtPriceA).TruncateInt()
}

func amount1Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int, roundUp bool) sdk.Int {
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
	if roundUp {
		return intermediate.Ceil().TruncateInt()
	}
	return intermediate.TruncateInt()
}

func Amount0Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return amount0Delta(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return amount0Delta(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func Amount1Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return amount1Delta(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return amount1Delta(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func NextSqrtPriceFromAmount0OutRoundingUp(sqrtPrice sdk.Dec, liquidity sdk.Int, amt sdk.Int) sdk.Dec {
	// amt == 0?
	numerator := liquidity.ToDec()
	product := sqrtPrice.MulInt(amt)
	denominator := numerator.Sub(product)
	return numerator.Mul(sqrtPrice).QuoRoundUp(denominator)
}

func NextSqrtPriceFromAmount1OutRoundingDown(sqrtPrice sdk.Dec, liquidity sdk.Int, amt sdk.Int) sdk.Dec {
	quotient := amt.ToDec().QuoInt(liquidity)
	return sqrtPrice.Sub(quotient)
}
