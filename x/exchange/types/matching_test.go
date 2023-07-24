package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestFillMemOrderBasic(t *testing.T) {
	market := types.NewMarket(1, "ucre", "uusd", utils.ParseDec("-0.0015"), utils.ParseDec("0.003"))
	ctx := types.NewMatchingContext(market, false)

	order := newUserMemOrder(1, true, utils.ParseDec("1.3"), sdk.NewDec(10_000000), sdk.NewDec(9_000000))
	ctx.FillMemOrder(order, sdk.NewDec(5_000000), utils.ParseDec("1.25"), true)

	require.True(t, order.IsMatched())
	require.True(sdk.DecEq(t, sdk.NewDec(6_250000), order.Paid()))
	require.True(sdk.DecEq(t, sdk.NewDec(5_000000), order.Received()))
	require.True(sdk.DecEq(t, sdk.NewDec(-9375), order.Fee()))

	order = newUserMemOrder(2, false, utils.ParseDec("1.2"), sdk.NewDec(10_000000), sdk.NewDec(9_000000))
	ctx.FillMemOrder(order, sdk.NewDec(5_000000), utils.ParseDec("1.25"), false)

	require.True(sdk.DecEq(t, sdk.NewDec(5_000000), order.Paid()))
	require.True(sdk.DecEq(t, sdk.NewDec(6_250000), order.Received()))
	require.True(sdk.DecEq(t, sdk.NewDec(18750), order.Fee()))
}
