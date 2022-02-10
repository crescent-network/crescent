package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func TestMatchEngine_Matchable(t *testing.T) {
	for _, tc := range []struct {
		name      string
		orders    []types.Order
		matchable bool
	}{
		{
			"no orders",
			[]types.Order{},
			false,
		},
		{
			"only one order",
			[]types.Order{
				newBuyOrder(parseDec("1.0"), newInt(100)),
			},
			false,
		},
		{
			"only one order",
			[]types.Order{
				newSellOrder(parseDec("1.0"), newInt(100)),
			},
			false,
		},
		{
			"two orders with same price",
			[]types.Order{
				newBuyOrder(parseDec("1.0"), newInt(100)),
				newSellOrder(parseDec("1.0"), newInt(100)),
			},
			true,
		},
		{
			"two orders with different prices",
			[]types.Order{
				newBuyOrder(parseDec("1.5"), newInt(100)),
				newSellOrder(parseDec("0.5"), newInt(100)),
			},
			true,
		},
		{
			"two orders with not matchable prices",
			[]types.Order{
				newBuyOrder(parseDec("0.5"), newInt(100)),
				newSellOrder(parseDec("1.5"), newInt(100)),
			},
			false,
		},
		{
			"orders with matchable prices",
			[]types.Order{
				newBuyOrder(parseDec("1.5"), newInt(100)),
				newBuyOrder(parseDec("1.3"), newInt(100)),
				newSellOrder(parseDec("1.4"), newInt(100)),
				newSellOrder(parseDec("1.6"), newInt(100)),
			},
			true,
		},
		{
			"orders with not matchable prices",
			[]types.Order{
				newBuyOrder(parseDec("1.4"), newInt(100)),
				newBuyOrder(parseDec("1.3"), newInt(100)),
				newSellOrder(parseDec("1.5"), newInt(100)),
				newSellOrder(parseDec("1.6"), newInt(100)),
			},
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ob := types.NewOrderBook(tickPrec)
			ob.AddOrders(tc.orders...)
			engine := types.NewMatchEngineFromOrderBook(ob)
			require.Equal(t, tc.matchable, engine.Matchable())
		})
	}
}

func TestMatchEngine_EstimatedPriceDirection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		orders   []types.Order
		midPrice sdk.Dec
		dir      types.PriceDirection
	}{
		{
			"increasing",
			[]types.Order{
				newBuyOrder(parseDec("1.5"), newInt(100)),
				newSellOrder(parseDec("0.5"), newInt(99)),
			},
			parseDec("1.0"),
			types.PriceIncreasing,
		},
		{
			"decreasing",
			[]types.Order{
				newBuyOrder(parseDec("1.5"), newInt(99)),
				newSellOrder(parseDec("0.5"), newInt(100)),
			},
			parseDec("1.0"),
			types.PriceDecreasing,
		},
		{
			"staying",
			[]types.Order{
				newBuyOrder(parseDec("1.5"), newInt(100)),
				newSellOrder(parseDec("0.5"), newInt(100)),
			},
			parseDec("1.0"),
			types.PriceStaying,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ob := types.NewOrderBook(tickPrec)
			ob.AddOrders(tc.orders...)
			engine := types.NewMatchEngineFromOrderBook(ob)
			require.Equal(t, tc.dir, engine.EstimatedPriceDirection(tc.midPrice))
		})
	}
}
