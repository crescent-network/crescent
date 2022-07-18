package amm_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

func TestOrderBookView(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	for i := 0; i < 100; i++ {
		ob := amm.NewOrderBook()

		// Add 10 random orders for each buy and sell direction
		for j := 0; j < 10; j++ {
			price := utils.RandomDec(r, utils.ParseDec("0.5"), utils.ParseDec("2.0"))
			amt := utils.RandomInt(r, sdk.NewInt(1000), sdk.NewInt(10000))
			ob.AddOrder(newOrder(amm.Buy, price, amt))

			price = utils.RandomDec(r, utils.ParseDec("0.5"), utils.ParseDec("2.0"))
			amt = utils.RandomInt(r, sdk.NewInt(1000), sdk.NewInt(10000))
			ob.AddOrder(newOrder(amm.Sell, price, amt))
		}

		orders := ob.Orders()

		ov := ob.MakeView() // We do not call Match() here, to test amt view methods only

		// Test 100 random prices
		for j := 0; j < 100; j++ {
			price := utils.RandomDec(r, utils.ParseDec("0.4"), utils.ParseDec("2.1"))

			// buy amount over (inclusive)
			expected := sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Buy && order.GetPrice().GTE(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.BuyAmountOver(price, true)))

			// buy amount over (exclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Buy && order.GetPrice().GT(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.BuyAmountOver(price, false)))

			// buy amount under (inclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Buy && order.GetPrice().LTE(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.BuyAmountUnder(price, true)))

			// buy amount under (exclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Buy && order.GetPrice().LT(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.BuyAmountUnder(price, false)))

			// sell amount under (inclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Sell && order.GetPrice().LTE(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.SellAmountUnder(price, true)))

			// sell amount under (exclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Sell && order.GetPrice().LT(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.SellAmountUnder(price, false)))

			// sell amount over (inclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Sell && order.GetPrice().GTE(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.SellAmountOver(price, true)))

			// sell amount over (exclusive)
			expected = sdk.ZeroInt()
			for _, order := range orders {
				if order.GetDirection() == amm.Sell && order.GetPrice().GT(price) {
					expected = expected.Add(order.GetAmount())
				}
			}
			require.True(sdk.IntEq(t, expected, ov.SellAmountOver(price, false)))
		}
	}
}

func TestOrderBookView_NoMatch(t *testing.T) {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
	)

	ov := ob.MakeView()
	ov.Match()

	require.True(sdk.IntEq(t, sdk.NewInt(10000), ov.BuyAmountOver(utils.ParseDec("1.0"), true)))
	require.True(sdk.IntEq(t, sdk.NewInt(10000), ov.SellAmountUnder(utils.ParseDec("1.1"), true)))
}
