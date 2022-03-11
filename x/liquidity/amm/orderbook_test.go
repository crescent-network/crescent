package amm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

func newOrder(dir amm.OrderDirection, price sdk.Dec, amt sdk.Int) *amm.BaseOrder {
	var offerCoinDenom, demandCoinDenom string
	switch dir {
	case amm.Buy:
		offerCoinDenom, demandCoinDenom = "denom2", "denom1"
	case amm.Sell:
		offerCoinDenom, demandCoinDenom = "denom1", "denom2"
	}
	return amm.NewBaseOrder(dir, price, amt, sdk.NewCoin(offerCoinDenom, amm.OfferCoinAmount(dir, price, amt)), demandCoinDenom)
}

func TestOrderBook(t *testing.T) {
	ob := amm.NewOrderBook(
		newOrder(amm.Buy, utils.ParseDec("10.01"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("10.00"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("9.999"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("9.999"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("9.998"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("9.998"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("9.997"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("9.996"), sdk.NewInt(10000)),
	)

	highest, found := ob.HighestBuyPrice()
	require.True(t, found)
	require.True(sdk.DecEq(t, utils.ParseDec("10.01"), highest))
	lowest, found := ob.LowestSellPrice()
	require.True(t, found)
	require.True(sdk.DecEq(t, utils.ParseDec("9.996"), lowest))

	for _, tc := range []struct {
		price                 sdk.Dec
		expectedBuyAmt        int64
		expectedSellAmt       int64
		expectedNumBuyOrders  int
		expectedNumSellOrders int
	}{
		{utils.ParseDec("10.02"), 0, 40000, 0, 4},
		{utils.ParseDec("10.01"), 10000, 40000, 1, 4},
		{utils.ParseDec("10.00"), 20000, 40000, 2, 4},
		{utils.ParseDec("9.999"), 30000, 40000, 3, 4},
		{utils.ParseDec("9.998"), 40000, 30000, 4, 3},
		{utils.ParseDec("9.997"), 40000, 20000, 4, 2},
		{utils.ParseDec("9.996"), 40000, 10000, 4, 1},
		{utils.ParseDec("9.995"), 40000, 0, 4, 0},
	} {
		t.Run("", func(t *testing.T) {
			buyAmt := ob.BuyAmountOver(tc.price)
			require.True(sdk.IntEq(t, sdk.NewInt(tc.expectedBuyAmt), buyAmt))
			sellAmt := ob.SellAmountUnder(tc.price)
			require.True(sdk.IntEq(t, sdk.NewInt(tc.expectedSellAmt), sellAmt))

			require.Len(t, ob.BuyOrdersOver(tc.price), tc.expectedNumBuyOrders)
			require.Len(t, ob.SellOrdersUnder(tc.price), tc.expectedNumSellOrders)
		})
	}
}
