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
		types.DefaultFees, types.DefaultOrderQuantityLimits, types.DefaultOrderQuoteLimits)
	ctx := types.NewMatchingContext(market, false)

	order := newUserMemOrder(1, true, utils.ParseDec("1.0015"), sdk.NewInt(10000), sdk.NewInt(9000))
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), false)
	res := order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5508), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5483), res.Received)
	utils.AssertEqual(t, sdk.NewInt(17), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeeReceived)

	order = newUserMemOrder(2, false, utils.ParseDec("1.001"), sdk.NewInt(10000), sdk.NewInt(9000))
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), true)
	res = order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5500), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5498), res.Received)
	utils.AssertEqual(t, sdk.NewInt(9), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeeReceived)

	// We don't need to set actual OrderSource in this test.
	order = newOrderSourceMemOrder(true, utils.ParseDec("1.0015"), sdk.NewInt(10000), nil)
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), true)
	res = order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5499), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5500), res.Received)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(8), res.FeeReceived)

	order = newOrderSourceMemOrder(false, utils.ParseDec("1.001"), sdk.NewInt(10000), nil)
	ctx.FillOrder(order, sdk.NewInt(5500), utils.ParseDec("1.0013"), true)
	res = order.Result()
	utils.AssertEqual(t, sdk.NewInt(5500), res.ExecutedQuantity)
	utils.AssertEqual(t, sdk.NewInt(5492), res.Paid)
	utils.AssertEqual(t, sdk.NewInt(5507), res.Received)
	utils.AssertEqual(t, sdk.NewInt(0), res.FeePaid)
	utils.AssertEqual(t, sdk.NewInt(8), res.FeeReceived)
}

func TestRunSinglePriceAuctionEdgecase(t *testing.T) {
	market := types.NewMarket(
		1, "ucre", "uusd",
		types.DefaultFees, types.DefaultOrderQuantityLimits, types.DefaultOrderQuoteLimits)
	ctx := types.NewMatchingContext(market, false)

	buyObs := types.NewMemOrderBookSide(true)
	buyObs.AddOrder(
		newUserMemOrder(1, true, utils.ParseDec("100"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	buyObs.AddOrder(
		newUserMemOrder(2, true, utils.ParseDec("110"), sdk.NewInt(50_000000), sdk.NewInt(50_000000)))
	sellObs := types.NewMemOrderBookSide(false)
	sellObs.AddOrder(
		newUserMemOrder(3, false, utils.ParseDec("99"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	sellObs.AddOrder(
		newUserMemOrder(4, false, utils.ParseDec("98"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	sellObs.AddOrder(
		newUserMemOrder(5, false, utils.ParseDec("97"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	sellObs.AddOrder(
		newUserMemOrder(6, false, utils.ParseDec("96"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	sellObs.AddOrder(
		newUserMemOrder(7, false, utils.ParseDec("95"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	sellObs.AddOrder(
		newUserMemOrder(8, false, utils.ParseDec("94"), sdk.NewInt(10_000000), sdk.NewInt(10_000000)))
	//    | 110 | #####
	//    | 100 | #
	//  # |  99 |
	//  # |  98 |
	//  # |  97 |
	//  # |  96 |
	//  # |  95 |
	//  # |  94 |
	matchPrice, matched := ctx.RunSinglePriceAuction(buyObs, sellObs)
	require.True(t, matched)
	utils.AssertEqual(t, utils.ParseDec("99.5"), matchPrice)
}
