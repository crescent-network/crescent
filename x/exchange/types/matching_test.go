package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
		1, "ucre", "uusd",
		utils.ParseDec("-0.0015"), utils.ParseDec("0.003"), utils.ParseDec("0.5"),
		types.DefaultDefaultMinOrderQuantity, types.DefaultDefaultMinOrderQuote,
		types.DefaultDefaultMaxOrderQuantity, types.DefaultDefaultMaxOrderQuote)
	ctx := types.NewMatchingContext(market, false)

	order := newUserMemOrder(1, true, utils.ParseDec("1.0015"), sdk.NewInt(10000), sdk.NewInt(9000))
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), false)
	res := order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5508), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5483), res.Received)
	utils.AssertEqual(t, sdk.NewInt(17), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeeReceived)

	order = newUserMemOrder(2, false, utils.ParseDec("1.0015"), sdk.NewInt(10000), sdk.NewInt(9000))
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), true)
	res = order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5492), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5507), res.Received)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(8), res.FeeReceived)
}
