package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func LiquidityForAmount0(sqrtPriceA, sqrtPriceB sdk.Dec, amt0 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return sqrtPriceA.MulTruncate(sqrtPriceB).MulInt(amt0).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmount1(sqrtPriceA, sqrtPriceB sdk.Dec, amt1 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return sdk.NewDecFromInt(amt1).QuoTruncate(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmounts(currentSqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amt0, amt1 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	if currentSqrtPrice.LTE(sqrtPriceA) {
		return LiquidityForAmount0(sqrtPriceA, sqrtPriceB, amt0)
	} else if currentSqrtPrice.LT(sqrtPriceB) {
		liquidity0 := LiquidityForAmount0(currentSqrtPrice, sqrtPriceB, amt0)
		liquidity1 := LiquidityForAmount1(sqrtPriceA, currentSqrtPrice, amt1)
		return utils.MinInt(liquidity0, liquidity1)
	}
	return LiquidityForAmount1(sqrtPriceA, sqrtPriceB, amt1)
}

func Amount0DeltaRounding(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
	if roundUp {
		return intermediate.QuoRoundUp(sqrtPriceB).QuoRoundUp(sqrtPriceA).Ceil().TruncateInt()
	}
	return intermediate.QuoTruncate(sqrtPriceB).QuoTruncate(sqrtPriceA).TruncateInt()
}

func Amount1DeltaRounding(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
	if roundUp {
		return intermediate.Ceil().TruncateInt()
	}
	return intermediate.TruncateInt()
}

func Amount0Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func Amount1Delta(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func nextSqrtPriceFromAmount0RoundingUp(sqrtPrice sdk.Dec, liquidity sdk.Int, amt sdk.Dec, add bool) sdk.Dec {
	numerator := liquidity.ToDec()
	if add {
		// TODO: check overflow
		return numerator.QuoRoundUp(numerator.QuoTruncate(sqrtPrice).Add(amt))
	}
	product := sqrtPrice.Mul(amt)
	denominator := numerator.Sub(product)
	return numerator.Mul(sqrtPrice).QuoRoundUp(denominator)
}

func nextSqrtPriceFromAmount1RoundingDown(sqrtPrice sdk.Dec, liquidity sdk.Int, amt sdk.Dec, add bool) sdk.Dec {
	if add {
		quotient := amt.QuoTruncate(liquidity.ToDec())
		return sqrtPrice.Add(quotient)
	}
	quotient := amt.QuoRoundUp(liquidity.ToDec())
	return sqrtPrice.Sub(quotient)
}

func NextSqrtPriceFromOutput(sqrtPrice sdk.Dec, liquidity sdk.Int, amt sdk.Dec, isBuy bool) sdk.Dec {
	if isBuy {
		return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amt, false)
	}
	return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amt, false)
}

func Amount0DeltaRoundingDec(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int, roundUp bool) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
	if roundUp {
		return intermediate.QuoRoundUp(sqrtPriceB).QuoRoundUp(sqrtPriceA)
	}
	return intermediate.QuoTruncate(sqrtPriceB).QuoTruncate(sqrtPriceA)
}

func Amount1DeltaDec(sqrtPriceA, sqrtPriceB sdk.Dec, liquidity sdk.Int) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return sqrtPriceB.Sub(sqrtPriceA).MulInt(liquidity)
}
