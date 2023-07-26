package types_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestDeductFee(t *testing.T) {
	for i, tc := range []struct {
		amt, feeRate  sdk.Dec
		deducted, fee sdk.Dec
	}{
		{
			utils.ParseDec("123456789"), utils.ParseDec("0.003"),
			utils.ParseDec("123086418.633"), utils.ParseDec("370370.367"),
		},
		{
			utils.ParseDec("123456789"), utils.ParseDec("0.0015"),
			utils.ParseDec("123271603.8165"), utils.ParseDec("185185.1835"),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			deducted, fee := types.DeductFee(tc.amt, tc.feeRate)
			testutil.AssertEqual(t, tc.deducted, deducted)
			testutil.AssertEqual(t, tc.fee, fee)
		})
	}

	r := rand.New(rand.NewSource(1))
	for i := 0; i < 50; i++ {
		amt := utils.RandomDec(r, sdk.NewDec(10), sdk.NewDec(100000000))
		deducted, fee := types.DeductFee(amt, utils.ParseDec("0.003"))
		testutil.AssertEqual(t, amt, deducted.Add(fee))
	}
}

func TestPayReceiveDenoms(t *testing.T) {
	payDenom, receiveDenom := types.PayReceiveDenoms("ucre", "uusd", true)
	require.Equal(t, "uusd", payDenom)
	require.Equal(t, "ucre", receiveDenom)
	payDenom, receiveDenom = types.PayReceiveDenoms("ucre", "uusd", false)
	require.Equal(t, "ucre", payDenom)
	require.Equal(t, "uusd", receiveDenom)
}

func TestFillMemOrderBasic(t *testing.T) {
	market := types.NewMarket(1, "ucre", "uusd", utils.ParseDec("-0.0015"), utils.ParseDec("0.003"))
	ctx := types.NewMatchingContext(market, false)

	order := newUserMemOrder(1, true, utils.ParseDec("1.3"), sdk.NewDec(10_000000), sdk.NewDec(9_000000))
	ctx.FillOrder(order, sdk.NewDec(5_000000), utils.ParseDec("1.25"), true)

	require.True(t, order.IsMatched())
	testutil.AssertEqual(t, sdk.NewDec(6_240625), order.Paid())
	testutil.AssertEqual(t, sdk.NewDec(5_000000), order.Received())
	testutil.AssertEqual(t, sdk.NewDec(-9375), order.Fee())

	order = newUserMemOrder(2, false, utils.ParseDec("1.2"), sdk.NewDec(10_000000), sdk.NewDec(9_000000))
	ctx.FillOrder(order, sdk.NewDec(5_000000), utils.ParseDec("1.25"), false)

	testutil.AssertEqual(t, sdk.NewDec(5_000000), order.Paid())
	testutil.AssertEqual(t, sdk.NewDec(6_231250), order.Received())
	testutil.AssertEqual(t, sdk.NewDec(18750), order.Fee())
}
