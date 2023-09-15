package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
)

func LiquidityForAmount0(sqrtPriceA, sqrtPriceB cremath.BigDec, amt0 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return sqrtPriceA.MulTruncate(sqrtPriceB).MulIntMut(amt0).QuoTruncateMut(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmount1(sqrtPriceA, sqrtPriceB cremath.BigDec, amt1 sdk.Int) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	return cremath.NewBigDecFromInt(amt1).QuoTruncateMut(sqrtPriceB.Sub(sqrtPriceA)).TruncateInt()
}

func LiquidityForAmounts(currentSqrtPrice, sqrtPriceA, sqrtPriceB cremath.BigDec, amt0, amt1 sdk.Int) sdk.Int {
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

func Amount0DeltaRounding(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Int {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate.QuoRoundUpMut(sqrtPriceB).QuoRoundUpMut(sqrtPriceA).Ceil().TruncateInt()
	}
	return intermediate.QuoTruncateMut(sqrtPriceB).QuoTruncateMut(sqrtPriceA).TruncateInt()
}

func Amount1DeltaRounding(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Int {
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

func AmountsForLiquidity(currentSqrtPrice, sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int) (amt0, amt1 sdk.Int) {
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

func nextSqrtPriceFromAmount0RoundingUp(sqrtPrice cremath.BigDec, liquidity sdk.Int, amt sdk.Dec, add bool) cremath.BigDec {
	liquidityBigDec := cremath.NewBigDecFromInt(liquidity)
	amtBigDec := cremath.NewBigDecFromDec(amt)
	if add {
		// TODO: check overflow
		return liquidityBigDec.QuoRoundUpMut(liquidityBigDec.QuoTruncate(sqrtPrice).AddMut(amtBigDec))
	}
	product := sqrtPrice.Mul(amtBigDec)
	denominator := liquidityBigDec.Sub(product)
	return liquidityBigDec.MulMut(sqrtPrice).QuoRoundUpMut(denominator)
}

func nextSqrtPriceFromAmount1RoundingDown(sqrtPrice cremath.BigDec, liquidity sdk.Int, amt sdk.Dec, add bool) cremath.BigDec {
	liquidityBigDec := cremath.NewBigDecFromInt(liquidity)
	amtBigDec := cremath.NewBigDecFromDec(amt)
	if add {
		quotient := amtBigDec.QuoTruncateMut(liquidityBigDec)
		return sqrtPrice.Add(quotient)
	}
	quotient := amtBigDec.QuoRoundUpMut(liquidityBigDec)
	return sqrtPrice.Sub(quotient)
}

func NextSqrtPriceFromOutput(sqrtPrice cremath.BigDec, liquidity sdk.Int, amt sdk.Dec, isBuy bool) cremath.BigDec {
	if isBuy {
		return nextSqrtPriceFromAmount1RoundingDown(sqrtPrice, liquidity, amt, false)
	}
	return nextSqrtPriceFromAmount0RoundingUp(sqrtPrice, liquidity, amt, false)
}

func Amount0DeltaRoundingDec(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	intermediate := sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity)
	if roundUp {
		return intermediate.QuoRoundUp(sqrtPriceB).QuoRoundUp(sqrtPriceA).DecRoundUp()
	}
	return intermediate.QuoTruncate(sqrtPriceB).QuoTruncate(sqrtPriceA).Dec()
}

func Amount1DeltaRoundingDec(sqrtPriceA, sqrtPriceB cremath.BigDec, liquidity sdk.Int, roundUp bool) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	if roundUp {
		return sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity).DecRoundUp()
	}
	return sqrtPriceB.Sub(sqrtPriceA).MulIntMut(liquidity).Dec()
}
