package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/cremath"
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
			sdk.NewInt(274331),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012344"), utils.ParseDec("0.000000000000012346"),
			utils.ParseInt("1000000000000000000"), sdk.NewInt(10000),
			sdk.NewInt(2222116054455176),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012345"), utils.ParseDec("0.000000000000012346"),
			utils.ParseInt("1000000000000000000"), sdk.NewInt(0),
			sdk.NewInt(2743313825323908),
		},
		{
			utils.ParseDec("0.000000000000012345"),
			utils.ParseDec("0.000000000000012344"), utils.ParseDec("0.000000000000012345"),
			sdk.NewInt(0), sdk.NewInt(10000),
			sdk.NewInt(2222116054455176),
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
		price     sdk.Dec
		liquidity sdk.Int
		amt       sdk.Dec
		isBuy     bool
		nextPrice sdk.Dec
	}{
		{
			utils.ParseDec("1"), sdk.NewInt(1000000000), sdk.NewDec(100000), true,
			utils.ParseDec("0.999800010000000000"),
		},
		{
			utils.ParseDec("1"), sdk.NewInt(1000000000), sdk.NewDec(123456), true,
			utils.ParseDec("0.999753103241383936"),
		},
		{
			utils.ParseDec("1"), sdk.NewInt(1000000000), sdk.NewDec(123456), false,
			utils.ParseDec("1.000246957731679532"),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			nextSqrtPrice := types.NextSqrtPriceFromOutput(
				cremath.NewBigDecFromDec(tc.price).SqrtMut(), tc.liquidity, tc.amt, tc.isBuy)
			require.Equal(t, tc.nextPrice, nextSqrtPrice.Power(2))
		})
	}
}
