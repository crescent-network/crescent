package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
)

func LiquidityForAmount0(sqrtPriceA, sqrtPriceB cremath.BigDec, amt0 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return sqrtPriceA.MulTruncate(sqrtPriceB).MulIntMut(amt0).
		QuoTruncateMut(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmount1(sqrtPriceA, sqrtPriceB cremath.BigDec, amt1 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return cremath.NewBigDecFromInt(amt1).QuoTruncateMut(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmounts(
	currentSqrtPrice, sqrtPriceA, sqrtPriceB cremath.BigDec, amt0, amt1 sdk.Int) sdk.Int {
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

func Amount0DeltaRounding(
	sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate.QuoRoundUpMut(sqrtPriceB).QuoRoundUpMut(sqrtPriceA).Ceil().TruncateInt()
	}
	return intermediate.QuoTruncateMut(sqrtPriceB).QuoTruncateMut(sqrtPriceA).TruncateInt()
}

func Amount1DeltaRounding(
	sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate.Ceil().TruncateInt()
	}
	return intermediate.TruncateInt()
}

func Amount0Delta(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func Amount1Delta(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int) sdk.Int {
	if liquidity.IsNegative() {
		return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity.Neg(), false).Neg()
	}
	return Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
}

func AmountsForLiquidity(
	currentSqrtPrice, sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int) (amt0, amt1 sdk.Int) {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	if currentSqrtPrice.LTE(sqrtPriceA) {
		amt0 = Amount0DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
		amt1 = utils.ZeroInt
	} else if currentSqrtPrice.LT(sqrtPriceB) {
		amt0 = Amount0DeltaRounding(currentSqrtPrice, sqrtPriceB, liquidity, true)
		amt1 = Amount1DeltaRounding(sqrtPriceA, currentSqrtPrice, liquidity, true)
	} else {
		amt0 = utils.ZeroInt
		amt1 = Amount1DeltaRounding(sqrtPriceA, sqrtPriceB, liquidity, true)
	}
	return
}

func nextSqrtPriceFromAmount0RoundingUp(
	sqrtPrice cremath.BigDec, liquidity, amt sdk.Int, add bool) cremath.BigDec {
	if amt.IsZero() {
		return sqrtPrice
	}
	numerator := cremath.NewBigDecFromInt(liquidity)
	product := sqrtPrice.Mul(cremath.NewBigDecFromInt(amt))
	if add {
		denominator := numerator.Add(product)
		return numerator.MulRoundUpMut(sqrtPrice).QuoRoundUpMut(denominator)
	}
	denominator := numerator.Sub(product)
	return numerator.MulRoundUpMut(sqrtPrice).QuoRoundUpMut(denominator)
}

func nextSqrtPriceFromAmount1RoundingDown(
	sqrtPrice cremath.BigDec, liquidity, amt sdk.Int, add bool) cremath.BigDec {
	if add {
		quotient := cremath.NewBigDecFromInt(amt).QuoTruncateMut(cremath.NewBigDecFromInt(liquidity))
		return sqrtPrice.Add(quotient)
	}
	quotient := cremath.NewBigDecFromInt(amt).QuoRoundUpMut(cremath.NewBigDecFromInt(liquidity))
	return sqrtPrice.Sub(quotient)
}

func NextSqrtPriceFromInput(
	sqrtPrice cremath.BigDec, liquidity, amtIn sdk.Int, isBuy bool) cremath.BigDec {
	if isBuy {
		return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amtIn, true)
	}
	return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amtIn, true)
}

func NextSqrtPriceFromOutput(
	sqrtPrice cremath.BigDec, liquidity, amtOut sdk.Int, isBuy bool) cremath.BigDec {
	if isBuy {
		return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amtOut, false)
	}
	return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amtOut, false)
}

func Amount0DeltaRoundingBigDec(
	sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) cremath.BigDec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate.QuoRoundUpMut(sqrtPriceB).QuoRoundUpMut(sqrtPriceA)
	}
	return intermediate.QuoTruncateMut(sqrtPriceB).QuoTruncateMut(sqrtPriceA)
}

func Amount1DeltaRoundingBigDec(
	sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) cremath.BigDec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate // XXX
	}
	return intermediate // XXX
}

func nextSqrtPriceFromAmount0RoundingUpBigDec(
	sqrtPrice cremath.BigDec, liquidity sdk.Int, amt cremath.BigDec, add bool) cremath.BigDec {
	if amt.IsZero() {
		return sqrtPrice
	}
	numerator := cremath.NewBigDecFromInt(liquidity)
	product := sqrtPrice.Mul(amt)
	if add {
		denominator := numerator.Add(product)
		return numerator.MulRoundUpMut(sqrtPrice).QuoRoundUpMut(denominator)
	}
	denominator := numerator.Sub(product)
	return numerator.MulRoundUpMut(sqrtPrice).QuoRoundUpMut(denominator)
}

func nextSqrtPriceFromAmount1RoundingDownBigDec(
	sqrtPrice cremath.BigDec, liquidity sdk.Int, amt cremath.BigDec, add bool) cremath.BigDec {
	if add {
		quotient := amt.QuoTruncateMut(cremath.NewBigDecFromInt(liquidity))
		return sqrtPrice.Add(quotient)
	}
	quotient := amt.QuoRoundUpMut(cremath.NewBigDecFromInt(liquidity))
	return sqrtPrice.Sub(quotient)
}

func NextSqrtPriceFromInputBigDec(
	sqrtPrice cremath.BigDec, liquidity sdk.Int, amtIn cremath.BigDec, isBuy bool) cremath.BigDec {
	if isBuy {
		return nextSqrtPriceFromAmount0RoundingUpBigDec(sqrtPrice, liquidity, amtIn, true)
	}
	return nextSqrtPriceFromAmount1RoundingDownBigDec(sqrtPrice, liquidity, amtIn, true)
}

func NextSqrtPriceFromOutputBigDec(
	sqrtPrice cremath.BigDec, liquidity sdk.Int, amtOut cremath.BigDec, isBuy bool) cremath.BigDec {
	if isBuy {
		return nextSqrtPriceFromAmount1RoundingDownBigDec(sqrtPrice, liquidity, amtOut, false)
	}
	return nextSqrtPriceFromAmount0RoundingUpBigDec(sqrtPrice, liquidity, amtOut, false)
}
