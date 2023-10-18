package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestLiquidityForAmount0(t *testing.T) {
	for i, tc := range []struct {
		// args
		sqrtPriceA, sqrtPriceB cremath.BigDec
		amt0                   sdk.Int
		// result
		liquidity sdk.Int
	}{
		{
			utils.ParseBigDec("1").SqrtMut(),
			utils.ParseBigDec("2").SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(341421356),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MaxPrice).SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(100_000000),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MinPrice).SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(10),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			liquidity := types.LiquidityForAmount0(tc.sqrtPriceA, tc.sqrtPriceB, tc.amt0)
			utils.AssertEqual(t, tc.liquidity, liquidity)
			// The order of prices doesn't matter.
			liquidity2 := types.LiquidityForAmount0(tc.sqrtPriceB, tc.sqrtPriceA, tc.amt0)
			utils.AssertEqual(t, liquidity, liquidity2)
		})
	}
}

func TestLiquidityForAmount1(t *testing.T) {
	for i, tc := range []struct {
		// args
		sqrtPriceA, sqrtPriceB cremath.BigDec
		amt1                   sdk.Int
		// result
		liquidity sdk.Int
	}{
		{
			utils.ParseBigDec("1").SqrtMut(),
			utils.ParseBigDec("2").SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(241421356),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MaxPrice).SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(0),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MaxPrice).SqrtMut(),
			sdk.NewInt(1_000000_000000_000000),
			sdk.NewInt(1000000),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MinPrice).SqrtMut(),
			sdk.NewInt(100_000000),
			sdk.NewInt(100_000010),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			liquidity := types.LiquidityForAmount1(tc.sqrtPriceA, tc.sqrtPriceB, tc.amt1)
			utils.AssertEqual(t, tc.liquidity, liquidity)
			// The order of prices doesn't matter.
			liquidity2 := types.LiquidityForAmount1(tc.sqrtPriceB, tc.sqrtPriceA, tc.amt1)
			utils.AssertEqual(t, liquidity, liquidity2)
		})
	}
}

func TestLiquidityForAmounts(t *testing.T) {
	for i, tc := range []struct {
		currentPrice, priceA, priceB sdk.Dec
		amt0, amt1                   sdk.Int
		expected                     sdk.Int
	}{
		{
			utils.ParseDec("1"), utils.ParseDec("0.9"), utils.ParseDec("1.1"),
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(1948683298),
		},
		{
			utils.ParseDec("1"), utils.ParseDec("0.9"), utils.ParseDec("1.1"),
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(1948683298),
		},
		{
			utils.ParseDec("1"), utils.ParseDec("0.5"), utils.ParseDec("2"),
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(341421356),
		},
		{
			utils.ParseDec("1"),
			types.MinPrice, types.MaxPrice,
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(100000000),
		},
		{
			utils.ParseDec("1"), utils.ParseDec("0.99999"), utils.ParseDec("1.0001"),
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(2000149998750),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012344"), utils.ParseDec("0.000000000000012346"),
			sdk.NewInt(100_000000), sdk.NewInt(100_000000),
			sdk.NewInt(274342),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012344"), utils.ParseDec("0.000000000000012346"),
			utils.ParseInt("1000000000000000000"), sdk.NewInt(10000),
			sdk.NewInt(2222116108121897),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012345"), utils.ParseDec("0.000000000000012346"),
			utils.ParseInt("1000000000000000000"), sdk.NewInt(0),
			sdk.NewInt(2743424551587600),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012344"), utils.ParseDec("0.000000000000012345"),
			sdk.NewInt(0), sdk.NewInt(10000),
			sdk.NewInt(2222116108121897),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			liquidity := types.LiquidityForAmounts(
				cremath.NewBigDecFromDec(tc.currentPrice).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				tc.amt0, tc.amt1)
			utils.AssertEqual(t, tc.expected, liquidity)
			// The order of prices doesn't matter.
			liquidity2 := types.LiquidityForAmounts(
				cremath.NewBigDecFromDec(tc.currentPrice).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				tc.amt0, tc.amt1)
			utils.AssertEqual(t, liquidity, liquidity2)
		})
	}
}

func TestAmount0Delta(t *testing.T) {
	for i, tc := range []struct {
		priceA, priceB sdk.Dec
		liquidity      sdk.Int
		expected       sdk.Int
	}{
		{
			utils.ParseDec("1.1"), utils.ParseDec("1"),
			sdk.NewInt(1000000000),
			sdk.NewInt(46537411),
		},
		{
			utils.ParseDec("1.1"), utils.ParseDec("1"),
			sdk.NewInt(-1000000000),
			sdk.NewInt(-46537410),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			amt0 := types.Amount0Delta(
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				tc.liquidity)
			utils.AssertEqual(t, tc.expected, amt0)
		})
	}
}

func TestAmount1Delta(t *testing.T) {
	for i, tc := range []struct {
		priceA, priceB sdk.Dec
		liquidity      sdk.Int
		expected       sdk.Int
	}{
		{
			utils.ParseDec("1"), utils.ParseDec("1.1"),
			sdk.NewInt(1000000000),
			sdk.NewInt(48808849),
		},
		{
			utils.ParseDec("1"), utils.ParseDec("1.1"),
			sdk.NewInt(-1000000000),
			sdk.NewInt(-48808848),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			amt1 := types.Amount1Delta(
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				tc.liquidity)
			utils.AssertEqual(t, tc.expected, amt1)
		})
	}
}

func TestAmountsForLiquidity(t *testing.T) {
	for i, tc := range []struct {
		// args
		currentSqrtPrice       cremath.BigDec
		sqrtPriceA, sqrtPriceB cremath.BigDec
		liquidity              sdk.Int
		// result
		amt0, amt1 sdk.Int
	}{
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MinPrice).SqrtMut(),
			cremath.NewBigDecFromDec(types.MaxPrice).SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(100000000),
			sdk.NewInt(99999990),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MinPrice).SqrtMut(),
			utils.ParseBigDec("1").SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(0),
			sdk.NewInt(99999990),
		},
		{
			utils.ParseBigDec("1").SqrtMut(),
			utils.ParseBigDec("1").SqrtMut(),
			cremath.NewBigDecFromDec(types.MaxPrice).SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(100000000),
			sdk.NewInt(0),
		},
		{
			utils.ParseBigDec("100").SqrtMut(),
			utils.ParseBigDec("50").SqrtMut(),
			utils.ParseBigDec("90").SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(0),
			sdk.NewInt(241576517),
		},
		{
			utils.ParseBigDec("100").SqrtMut(),
			utils.ParseBigDec("130").SqrtMut(),
			utils.ParseBigDec("200").SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(1699513),
			sdk.NewInt(0),
		},
		{
			utils.ParseBigDec("100").SqrtMut(),
			utils.ParseBigDec("50").SqrtMut(),
			utils.ParseBigDec("200").SqrtMut(),
			sdk.NewInt(100000000),
			sdk.NewInt(2928933),
			sdk.NewInt(292893219),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			amt0, amt1 := types.AmountsForLiquidity(
				tc.currentSqrtPrice, tc.sqrtPriceA, tc.sqrtPriceB, tc.liquidity)
			utils.AssertEqual(t, tc.amt0, amt0)
			utils.AssertEqual(t, tc.amt1, amt1)
			// The order of prices doesn't matter.
			amt0_, amt1_ := types.AmountsForLiquidity(
				tc.currentSqrtPrice, tc.sqrtPriceB, tc.sqrtPriceA, tc.liquidity)
			utils.AssertEqual(t, amt0, amt0_)
			utils.AssertEqual(t, amt1, amt1_)
		})
	}
}

func TestNextSqrtPriceFromOutput(t *testing.T) {
	for i, tc := range []struct {
		currentSqrtPrice cremath.BigDec
		liquidity        sdk.Int
		amt              sdk.Int
		isBuy            bool
		nextSqrtPrice    cremath.BigDec
	}{
		{
			utils.ParseBigDec("1"), sdk.NewInt(1000000000), sdk.NewInt(100000), true,
			utils.ParseBigDec("0.9999"),
		},
		{
			utils.ParseBigDec("1"), sdk.NewInt(1000000000), sdk.NewInt(123456), true,
			utils.ParseBigDec("0.999876544"),
		},
		{
			utils.ParseBigDec("1"), sdk.NewInt(1000000000), sdk.NewInt(123456), false,
			utils.ParseBigDec("1.000123471243265808623669443734845730"),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			nextSqrtPrice := types.NextSqrtPriceFromOutput(
				tc.currentSqrtPrice, tc.liquidity, tc.amt, tc.isBuy)
			utils.AssertEqual(t, tc.nextSqrtPrice, nextSqrtPrice)
		})
	}
}
