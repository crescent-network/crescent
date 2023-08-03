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
	lowerPrice, upperPrice, currentPrice sdk.Dec, desiredAmt0, desiredAmt1, liquidity, amt0, amt1 sdk.Int) {
	if amt0.GT(desiredAmt0) {
		panic(fmt.Errorf("amt0 %s must be smaller than desired amt0 %s", amt0, desiredAmt0))
	}
	if amt1.GT(desiredAmt1) {
		panic(fmt.Errorf("amt1 %s must be smaller than desired amt1 %s", amt1, desiredAmt1))
	}
	sqrtPriceLower := utils.DecApproxSqrt(lowerPrice)
	sqrtPriceUpper := utils.DecApproxSqrt(upperPrice)
	sqrtPriceCurrent := utils.DecApproxSqrt(currentPrice)

	t := liquidity.ToDec().Sub(
		amt0.ToDec().Mul(
			sqrtPriceUpper.Mul(sqrtPriceCurrent).Quo(sqrtPriceUpper.Sub(sqrtPriceCurrent))))
	if !t.IsNegative() {
		panic(fmt.Errorf("%s must be negtaive", t))
	}
	t = liquidity.ToDec().Sub(amt1.ToDec().Quo(sqrtPriceCurrent.Sub(sqrtPriceLower)))
	if !t.IsNegative() {
		panic(fmt.Errorf("%s must be negtaive", t))
	}
}

// ValidateRemoveLiquidityResult validates the result of RemoveLiquidity.
// TODO: remove after enough testing
func ValidateRemoveLiquidityResult(
	lowerPrice, upperPrice, currentPrice sdk.Dec, liquidity, amt0, amt1 sdk.Int) {
	sqrtPriceLower := utils.DecApproxSqrt(lowerPrice)
	sqrtPriceUpper := utils.DecApproxSqrt(upperPrice)
	sqrtPriceCurrent := utils.DecApproxSqrt(currentPrice)

	t := amt0.ToDec().Sub(
		liquidity.ToDec().Mul(
			sqrtPriceUpper.Sub(sqrtPriceCurrent).Quo(sqrtPriceUpper.Mul(sqrtPriceCurrent))))
	if !t.IsNegative() {
		panic(fmt.Errorf("%s must be negative", t))
	}
	t = amt1.ToDec().Sub(liquidity.ToDec().Mul(sqrtPriceCurrent.Sub(sqrtPriceLower)))
	if !t.IsNegative() {
		panic(fmt.Errorf("%s must be negative", t))
	}
}

func ValidatePoolPriceAfterMatching(isBuy bool, tickSpacing uint32, lastPrice, currentPrice sdk.Dec) {
	if isBuy { // price decreased
		priceAbove := exchangetypes.PriceAtTick(exchangetypes.TickAtPrice(lastPrice) + int32(tickSpacing))
		if !(lastPrice.LTE(currentPrice) && currentPrice.LTE(priceAbove)) {
			panic(fmt.Errorf("must satisfy: %s <= %s <= %s", lastPrice, currentPrice, priceAbove))
		}
	} else {
		priceBelow := exchangetypes.PriceAtTick(exchangetypes.TickAtPrice(lastPrice) - int32(tickSpacing))
		if !(priceBelow.LTE(currentPrice) && currentPrice.LTE(lastPrice)) {
			panic(fmt.Errorf("must satisfy: %s <= %s <= %s", priceBelow, currentPrice, lastPrice))
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
			if t := poolState.CurrentPrice.MulInt(amt0); !t.LTE(amt1.ToDec()) {
				panic(fmt.Errorf("must satisfy: %s*%s(=%s) <= %s", poolState.CurrentPrice, amt0, t, amt1))
			}
		} else {
			if t := poolState.CurrentPrice.MulInt(amt0); !t.GTE(amt1.ToDec()) {
				panic(fmt.Errorf("must satisfy: %s*%s(=%s) >= %s", poolState.CurrentPrice, amt0, t, amt1))
			}
		}
	}

	// Pool price check
	threshold := utils.ParseDec("0.00001")
	t := utils.OneDec.Sub(
		liquidity.Mul(sqrtPriceLower).Quo(
			currentSqrtPrice.Mul(amt0.ToDec().Mul(sqrtPriceUpper).Add(liquidity))))
	if !t.Abs().LT(threshold) {
		panic(fmt.Errorf("must satisfy: %s < %s", t.Abs(), threshold))
	}
	t = utils.OneDec.Sub(
		amt1.ToDec().Add(liquidity.Mul(sqrtPriceLower)).Quo(
			liquidity.Mul(currentSqrtPrice)))
	if !t.Abs().LT(threshold) {
		panic(fmt.Errorf("must satisfy: %s < %s", t.Abs(), threshold))
	}

	// Tokens check
	t = amt0.ToDec().Sub(liquidity.Mul(sqrtPriceUpper.Sub(currentSqrtPrice).Quo(sqrtPriceUpper.Mul(currentSqrtPrice))))
	if !t.IsPositive() {
		panic(fmt.Errorf("must satisfy: %s > 0", t))
	}
	t = amt1.ToDec().Sub(liquidity.Mul(currentSqrtPrice.Sub(sqrtPriceLower)))
	if !t.IsPositive() {
		panic(fmt.Errorf("must satisfy: %s > 0", t))
	}
}
