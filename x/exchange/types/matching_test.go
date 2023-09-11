package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestPayReceiveDenoms(t *testing.T) {
	payDenom, receiveDenom := types.PayReceiveDenoms("ucre", "uusd", true)
	require.Equal(t, "uusd", payDenom)
	require.Equal(t, "ucre", receiveDenom)
	payDenom, receiveDenom = types.PayReceiveDenoms("ucre", "uusd", false)
	require.Equal(t, "ucre", payDenom)
	require.Equal(t, "uusd", receiveDenom)
}

func TestFillMemOrderBasic(t *testing.T) {
	market := types.NewMarket(
		1, "ucre", "uusd", utils.ParseDec("-0.0015"), utils.ParseDec("0.003"), utils.ParseDec("0.5"))
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
