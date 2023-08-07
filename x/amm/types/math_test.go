package types_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
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
			exchangetypes.PriceAtTick(types.MinTick), exchangetypes.PriceAtTick(types.MaxTick),
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
				utils.DecApproxSqrt(tc.currentPrice),
				utils.DecApproxSqrt(tc.priceA), utils.DecApproxSqrt(tc.priceB),
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
				utils.DecApproxSqrt(tc.priceA), utils.DecApproxSqrt(tc.priceB), tc.liquidity)
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
				utils.DecApproxSqrt(tc.priceA), utils.DecApproxSqrt(tc.priceB), tc.liquidity)
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
				utils.DecApproxSqrt(tc.price), tc.liquidity, tc.amt, tc.isBuy)
			require.Equal(t, tc.nextPrice, nextSqrtPrice.Power(2))
		})
	}
}

func TestCPMMAdjustment(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for i := 0; i < 100; i++ {
		seed := r.Int63()
		r := rand.New(rand.NewSource(seed))

		liquidity := utils.RandomInt(r, sdk.NewInt(10000), sdk.NewInt(10000000000))
		currentPrice := utils.RandomDec(r, utils.ParseDec("0.5"), utils.ParseDec("2"))
		currentSqrtPrice := utils.DecApproxSqrt(currentPrice)

		for _, tickSpacing := range []uint32{5, 10, 50} {
			for _, isBuy := range []bool{true, false} {
				orderPrice := types.AdjustPriceToTickSpacing(currentPrice, tickSpacing, !isBuy)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)

				var prevSqrtPrice, qty sdk.Dec
				if isBuy {
					qty = types.Amount1DeltaDec(prevSqrtPrice, orderSqrtPrice, liquidity).
						QuoTruncate(orderPrice)
				} else {
					qty = types.Amount0DeltaRoundingDec(prevSqrtPrice, orderSqrtPrice, liquidity, false)
				}

				executedRatio := utils.ParseDec("1") // starting from 100%
				step := utils.ParseDec("0.01")
				for ; executedRatio.IsPositive(); executedRatio = executedRatio.Sub(step) {
					executedQty := qty.Mul(executedRatio)
					var paid, received sdk.Dec
					if isBuy {
						paid = exchangetypes.QuoteAmount(true, orderPrice, executedQty)
						received = executedQty
					} else {
						paid = executedQty
						received = exchangetypes.QuoteAmount(false, orderPrice, executedQty)
					}

					var nextSqrtPrice sdk.Dec
					if executedRatio.Equal(sdk.OneDec()) {
						nextSqrtPrice = utils.DecApproxSqrt(orderPrice)
					} else {
						nextSqrtPrice = types.NextSqrtPriceFromOutput(
							currentSqrtPrice, liquidity, paid, isBuy)
					}

					var expectedReceived sdk.Dec
					if isBuy {
						expectedReceived = types.Amount0DeltaRoundingDec(
							currentSqrtPrice, nextSqrtPrice, liquidity, true)
					} else {
						expectedReceived = types.Amount1DeltaDec(
							currentSqrtPrice, nextSqrtPrice, liquidity)
					}

					receivedDiff := received.Sub(expectedReceived)
					require.True(t, !receivedDiff.IsNegative(), receivedDiff)
				}
			}
		}
	}
}
