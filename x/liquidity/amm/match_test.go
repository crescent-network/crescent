package amm_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

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
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
			).MakeView(),
			true,
			utils.ParseDec("1.0"),
		},
		{
			"buy order only",
			amm.NewOrderBook(
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			).MakeView(),
			false,
			sdk.Dec{},
		},
		{
			"sell order only",
			amm.NewOrderBook(
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			).MakeView(),
			false,
			sdk.Dec{},
		},
		{
			"highest buy price is lower than lowest sell price",
			amm.NewOrderBook(
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("0.9"), sdk.NewInt(10000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
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
			amm.NewBaseOrder(amm.Buy, defTickPrec.UpTick(defTickPrec.UpTick(basePrice)), sdk.NewInt(80)),
			amm.NewBaseOrder(amm.Sell, defTickPrec.UpTick(basePrice), sdk.NewInt(20)),
			amm.NewBaseOrder(amm.Buy, basePrice, sdk.NewInt(10)), amm.NewBaseOrder(amm.Sell, basePrice, sdk.NewInt(10)),
			amm.NewBaseOrder(amm.Sell, defTickPrec.DownTick(basePrice), sdk.NewInt(70)),
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
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			),
			utils.ParseDec("1.0"),
			true,
			utils.ParseDec("1.0"),
			sdk.ZeroInt(),
		},
		{
			"happy case #2",
			amm.NewOrderBook(
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
			),
			utils.ParseDec("1.0"),
			true,
			utils.ParseDec("1.0"),
			sdk.ZeroInt(),
		},
		{
			"positive quote coin dust",
			amm.NewOrderBook(
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				amm.NewBaseOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
				amm.NewBaseOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
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
