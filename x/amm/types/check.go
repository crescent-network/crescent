package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// ValidateAddLiquidityResult validates the result of AddLiquidity.
// TODO: remove it after enough testing
func ValidateAddLiquidityResult(
	desiredAmt0, desiredAmt1, amt0, amt1 sdk.Int) {
	if amt0.GT(desiredAmt0) {
		panic(fmt.Errorf("amt0 %s must be smaller than desired amt0 %s", amt0, desiredAmt0))
	}
	if amt1.GT(desiredAmt1) {
		panic(fmt.Errorf("amt1 %s must be smaller than desired amt1 %s", amt1, desiredAmt1))
	}
}

func ValidatePoolPriceAfterMatching(isBuy bool, lastPrice, currentPrice, initialPrice sdk.Dec) {
	if isBuy { // price decreased
		if !(lastPrice.LTE(currentPrice) && currentPrice.LTE(initialPrice)) {
			panic(fmt.Errorf("must satisfy: %s <= %s <= %s", lastPrice, currentPrice, initialPrice))
		}
	} else {
		if !(initialPrice.LTE(currentPrice) && currentPrice.LTE(lastPrice)) {
			panic(fmt.Errorf("must satisfy: %s <= %s <= %s", initialPrice, currentPrice, lastPrice))
		}
	}
}

// ValidatePositionState validates position state.
// TODO: remove after enough testing
func ValidatePositionState(pool Pool, poolState PoolState, position Position, amt sdk.Coins) {
	lowerPrice := exchangetypes.PriceAtTick(position.LowerTick)
	upperPrice := exchangetypes.PriceAtTick(position.UpperTick)
	currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
	sqrtPriceLower := SqrtPriceAtTick(position.LowerTick)
	sqrtPriceUpper := SqrtPriceAtTick(position.UpperTick)
	liquidity := position.Liquidity.ToDec()
	amt0, amt1 := amt.AmountOf(pool.Denom0), amt.AmountOf(pool.Denom1)

	// Position asset check
	if poolState.CurrentPrice.LTE(lowerPrice) {
		if len(amt) != 1 || !amt1.IsZero() {
			panic(fmt.Errorf("there must be token0 only: %s", amt))
		}
	} else if poolState.CurrentPrice.GTE(upperPrice) {
		if len(amt) != 1 || !amt0.IsZero() {
			panic(fmt.Errorf("there must be token1 only: %s", amt))
		}
	} else {
		if len(amt) != 2 {
			panic(fmt.Errorf("there must be both token0 and token1: %s", amt))
		}
		if poolState.CurrentPrice.GTE(utils.DecApproxSqrt(lowerPrice.Mul(upperPrice))) {
			if t := poolState.CurrentPrice.MulInt(amt0.Sub(utils.OneInt)); !t.LTE(amt1.ToDec()) {
				panic(fmt.Errorf("must satisfy: %s*%s(=%s) <= %s", poolState.CurrentPrice, amt0, t, amt1))
			}
		} else {
			if t := poolState.CurrentPrice.MulInt(amt0.Add(utils.OneInt)); !t.GTE(amt1.ToDec()) {
				panic(fmt.Errorf("must satisfy: %s*%s(=%s) >= %s", poolState.CurrentPrice, amt0, t, amt1))
			}
		}
	}

	// Tokens check
	if poolState.CurrentPrice.LTE(lowerPrice) {
		t2 := amt0.Sub(Amount0Delta(sqrtPriceLower, sqrtPriceUpper, position.Liquidity))
		if t2.IsPositive() {
			panic(fmt.Errorf("must satisfy: %s <= 0", t2))
		}
	} else if poolState.CurrentPrice.LTE(upperPrice) {
		t2 := amt0.Sub(Amount0Delta(currentSqrtPrice, sqrtPriceUpper, position.Liquidity))
		if t2.IsPositive() {
			panic(fmt.Errorf("must satisfy: %s <= 0", t2))
		}
		t2 = amt1.Sub(Amount1Delta(sqrtPriceLower, currentSqrtPrice, position.Liquidity))
		if t2.IsPositive() {
			panic(fmt.Errorf("must satisfy: %s <= 0", t2))
		}

		// Pool price check
		if liquidity.GTE(sdk.NewDec(100000)) {
			threshold := utils.ParseDec("0.00003")
			t := utils.OneDec.Sub(
				liquidity.Mul(sqrtPriceUpper).Quo(
					currentSqrtPrice.Mul(amt0.ToDec().Mul(sqrtPriceUpper).Add(liquidity))))
			if !t.Abs().LT(threshold) {
				fmt.Printf("[amt0] L=%s sqrtLower=%s sqrtUpper=%s sqrtCurrent=%s amt0=%s amt1=%s\n",
					liquidity, sqrtPriceLower, sqrtPriceUpper, currentSqrtPrice, amt0, amt1)
				panic(fmt.Errorf("must satisfy: %s < %s", t.Abs(), threshold))
			}
			t = utils.OneDec.Sub(
				amt1.ToDec().Add(liquidity.Mul(sqrtPriceLower)).Quo(
					liquidity.Mul(currentSqrtPrice)))
			if !t.Abs().LT(threshold) {
				fmt.Printf("[amt1] L=%s sqrtLower=%s sqrtUpper=%s sqrtCurrent=%s amt0=%s amt1=%s\n",
					liquidity, sqrtPriceLower, sqrtPriceUpper, currentSqrtPrice, amt0, amt1)
				panic(fmt.Errorf("must satisfy: %s < %s", t.Abs(), threshold))
			}
		}
	} else {
		t2 := amt1.Sub(Amount1Delta(sqrtPriceLower, sqrtPriceUpper, position.Liquidity))
		if t2.IsPositive() {
			panic(fmt.Errorf("must satisfy: %s <= 0", t2))
		}
	}
}
