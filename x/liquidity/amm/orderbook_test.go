package amm_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
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
}
