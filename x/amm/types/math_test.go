package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

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
			require.Equal(t, tc.expected, liquidity)
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
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			amt0 := types.Amount0Delta(
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				tc.liquidity)
			require.Equal(t, tc.expected, amt0)
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
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			amt1 := types.Amount1Delta(
				cremath.NewBigDecFromDec(tc.priceA).SqrtMut(),
				cremath.NewBigDecFromDec(tc.priceB).SqrtMut(),
				tc.liquidity)
			require.Equal(t, tc.expected, amt1)
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
