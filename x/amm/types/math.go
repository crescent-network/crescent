package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func LiquidityForAmount0(sqrtPriceA, sqrtPriceB sdk.Dec, amt0 sdk.Int) sdk.Dec {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	// TODO: possible precision loss?
	return sqrtPriceA.MulTruncate(sqrtPriceB).MulInt(amt0).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA))
}

func LiquidityForAmount1(sqrtPriceA, sqrtPriceB sdk.Dec, amt1 sdk.Int) sdk.Dec {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	return sdk.NewDecFromInt(amt1).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA))
}

func LiquidityForAmounts(currentSqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amt0, amt1 sdk.Int) sdk.Dec {
	//if sqrtPriceA.GT(sqrtPriceB) {
	//	sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	//}
	if currentSqrtPrice.LTE(sqrtPriceA) {
		return LiquidityForAmount0(sqrtPriceA, sqrtPriceB, amt0)
	} else if currentSqrtPrice.LT(sqrtPriceB) {
		liquidity0 := LiquidityForAmount0(currentSqrtPrice, sqrtPriceB, amt0)
		liquidity1 := LiquidityForAmount1(sqrtPriceA, currentSqrtPrice, amt1)
		return sdk.MinDec(liquidity0, liquidity1)
	}
	return LiquidityForAmount1(sqrtPriceA, sqrtPriceB, amt1)
}

func Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity sdk.Dec, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).Mul(liquidity)
	if roundUp {
		return intermediate.QuoRoundUp(sqrtPriceB).QuoRoundUp(sqrtPriceA).Ceil().TruncateInt()
	}
	return intermediate.QuoTruncate(sqrtPriceB).QuoTruncate(sqrtPriceA).TruncateInt()
}

func Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity sdk.Dec, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).Mul(liquidity)
	if roundUp {
		return intermediate.Ceil().TruncateInt()
	}
	return intermediate.TruncateInt()
}

func Amount0Delta(sqrtPriceA, sqrtPriceB, liquidity sdk.Dec) sdk.Int {
	if liquidity.IsNegative() {
		return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func Amount1Delta(sqrtPriceA, sqrtPriceB, liquidity sdk.Dec) sdk.Int {
	if liquidity.IsNegative() {
		return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity sdk.Dec, amt sdk.Int, add bool) sdk.Dec {
	numerator := liquidity
	if add {
		// TODO: check overflow
		return numerator.QuoRoundUp(numerator.QuoTruncate(sqrtPrice).Add(amt.ToDec()))
	}
	product := sqrtPrice.MulInt(amt)
	denominator := numerator.Sub(product)
	return numerator.Mul(sqrtPrice).QuoRoundUp(denominator)
}

func nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity sdk.Dec, amt sdk.Int, add bool) sdk.Dec {
	if add {
		quotient := amt.ToDec().QuoTruncate(liquidity)
		return sqrtPrice.Add(quotient)
	}
	quotient := amt.ToDec().QuoRoundUp(liquidity)
	return sqrtPrice.Sub(quotient)
}

func NextSqrtPriceFromOutput(sqrtPrice, liquidity sdk.Dec, amt sdk.Int, isBuy bool) sdk.Dec {
	if isBuy {
		return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amt, false)
	}
	return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amt, false)
}

func NextSqrtPriceFromInput(sqrtPrice, liquidity sdk.Dec, amt sdk.Int, isBuy bool) sdk.Dec {
	if isBuy {
		return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amt, true)
	}
	return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amt, true)
}
