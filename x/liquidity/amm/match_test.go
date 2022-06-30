package amm_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

func newOrder(dir amm.OrderDirection, price sdk.Dec, amt sdk.Int) amm.Order {
	return amm.DefaultOrderer.Order(dir, price, amt)
}

func TestFindMatchPrice(t *testing.T) {
	for _, tc := range []struct {
		name       string
		ov         amm.OrderView
		found      bool
		matchPrice sdk.Dec
	}{
		{
			"happy case",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
			).MakeView(),
			true,
			utils.ParseDec("1.0"),
		},
		{
			"buy order only",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			).MakeView(),
			false,
			sdk.Dec{},
		},
		{
			"sell order only",
			amm.NewOrderBook(
				newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			).MakeView(),
			false,
			sdk.Dec{},
		},
		{
			"highest buy price is lower than lowest sell price",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("0.9"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
			).MakeView(),
			false,
			sdk.Dec{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			matchPrice, found := amm.FindMatchPrice(tc.ov, int(defTickPrec))
			require.Equal(t, tc.found, found)
			if found {
				require.Equal(t, tc.matchPrice, matchPrice)
			}
		})
	}
}

func TestFindMatchPrice_Rounding(t *testing.T) {
	basePrice := utils.ParseDec("0.9990")

	for i := 0; i < 50; i++ {
		ob := amm.NewOrderBook(
			newOrder(amm.Buy, defTickPrec.UpTick(defTickPrec.UpTick(basePrice)), sdk.NewInt(80)),
			newOrder(amm.Sell, defTickPrec.UpTick(basePrice), sdk.NewInt(20)),
			newOrder(amm.Buy, basePrice, sdk.NewInt(10)), newOrder(amm.Sell, basePrice, sdk.NewInt(10)),
			newOrder(amm.Sell, defTickPrec.DownTick(basePrice), sdk.NewInt(70)),
		)
		matchPrice, found := amm.FindMatchPrice(ob.MakeView(), int(defTickPrec))
		require.True(t, found)
		require.True(sdk.DecEq(t,
			defTickPrec.RoundPrice(basePrice.Add(defTickPrec.UpTick(basePrice)).QuoInt64(2)),
			matchPrice))

		basePrice = defTickPrec.UpTick(basePrice)
	}
}

func TestMatchOrders(t *testing.T) {
	_, _, matched := amm.NewOrderBook().Match(utils.ParseDec("1.0"))
	require.False(t, matched)

	for _, tc := range []struct {
		name          string
		ob            *amm.OrderBook
		lastPrice     sdk.Dec
		matched       bool
		matchPrice    sdk.Dec
		quoteCoinDust sdk.Int
	}{
		{
			"happy case",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			),
			utils.ParseDec("1.0"),
			true,
			utils.ParseDec("1.0"),
			sdk.ZeroInt(),
		},
		{
			"happy case #2",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
			),
			utils.ParseDec("1.0"),
			true,
			utils.ParseDec("1.0"),
			sdk.ZeroInt(),
		},
		{
			"positive quote coin dust",
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				newOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				newOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				newOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
			),
			utils.ParseDec("0.9999"),
			true,
			utils.ParseDec("0.9999"),
			sdk.NewInt(2),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			matchPrice, quoteCoinDust, matched := tc.ob.Match(tc.lastPrice)
			require.Equal(t, tc.matched, matched)
			require.True(sdk.DecEq(t, tc.matchPrice, matchPrice))
			if matched {
				require.True(sdk.IntEq(t, tc.quoteCoinDust, quoteCoinDust))
				for _, order := range tc.ob.Orders() {
					if order.IsMatched() {
						paid := order.GetPaidOfferCoinAmount()
						received := order.GetReceivedDemandCoinAmount()
						var effPrice sdk.Dec // Effective swap price
						switch order.GetDirection() {
						case amm.Buy:
							effPrice = paid.ToDec().QuoInt(received)
						case amm.Sell:
							effPrice = received.ToDec().QuoInt(paid)
						}
						require.True(t, utils.DecApproxEqual(tc.lastPrice, effPrice))
					}
				}
			}
		})
	}
}

func TestFindMatchableAmountAtSinglePrice(t *testing.T) {
	for _, tc := range []struct {
		orders       []amm.Order
		matchPrice   sdk.Dec
		found        bool
		matchableAmt sdk.Int
	}{
		{
			[]amm.Order{
				newOrder(amm.Sell, utils.ParseDec("0.100"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("0.099"), sdk.NewInt(9995)),
				newOrder(amm.Buy, utils.ParseDec("0.101"), sdk.NewInt(10000)),
			},
			utils.ParseDec("0.100"),
			true,
			sdk.NewInt(9995),
		},
		{
			[]amm.Order{
				newOrder(amm.Sell, utils.ParseDec("0.100"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("0.099"), sdk.NewInt(9995)),
				newOrder(amm.Buy, utils.ParseDec("0.101"), sdk.NewInt(10000)),
				newOrder(amm.Buy, utils.ParseDec("0.100"), sdk.NewInt(1000)),
			},
			utils.ParseDec("0.100"),
			true,
			sdk.NewInt(11000),
		},
	} {
		t.Run("", func(t *testing.T) {
			ob := amm.NewOrderBook(tc.orders...)
			matchableAmt, found := ob.FindMatchableAmountAtSinglePrice(tc.matchPrice)
			require.Equal(t, tc.found, found)
			if found {
				require.True(sdk.IntEq(t, tc.matchableAmt, matchableAmt))
			}
		})
	}
}

func TestMatch_edgecase1(t *testing.T) {
	orders := []amm.Order{
		newOrder(amm.Sell, utils.ParseDec("0.100"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("0.099"), sdk.NewInt(9995)),
		newOrder(amm.Buy, utils.ParseDec("0.101"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("0.100"), sdk.NewInt(5000)),
	}
	ob := amm.NewOrderBook(orders...)
	_, _, matched := ob.Match(utils.ParseDec("0.098"))
	require.True(t, matched)
	for _, order := range orders {
		fmt.Printf(
			"%s %s (%s/%s) paid=%s, received=%s\n",
			order.GetDirection(), order.GetPrice(), order.GetOpenAmount(), order.GetAmount(),
			order.GetPaidOfferCoinAmount(), order.GetReceivedDemandCoinAmount())
	}
}
