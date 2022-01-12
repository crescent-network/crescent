package types_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func testOrderBookTicks() *types.OrderBookTicks {
	ticks := types.NewOrderBookTicks(tickPrec)
	ticks.AddOrders(
		newBuyOrder("20.0", 1000),
		newBuyOrder("19.0", 1000),
		newBuyOrder("18.0", 1000),
		newBuyOrder("17.0", 1000),
		newBuyOrder("16.0", 1000),
		newBuyOrder("15.0", 1000),
		newBuyOrder("14.0", 1000),
		newBuyOrder("13.0", 1000),
		newBuyOrder("12.0", 1000),
		newBuyOrder("11.0", 1000),
		newBuyOrder("10.0", 1000),
	)
	return ticks
}

func TestOrderBookTicks_FindPrice(t *testing.T) {
	// An empty order book ticks must return (0, false).
	i, exact := types.NewOrderBookTicks(tickPrec).FindPrice(newDec("20.0"))
	require.False(t, exact)
	require.Equal(t, 0, i)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		i     int
		exact bool
	}{
		{newDec("20.0"), 0, true},
		{newDec("19.99999999999999999"), 1, false},
		{newDec("19.00000000000000001"), 1, false},
		{newDec("19.0"), 1, true},
		{newDec("18.99999999999999999"), 2, false},
		{newDec("18.00000000000000001"), 2, false},
		{newDec("18.0"), 2, true},
		{newDec("9.999999999999999999"), 11, false},
	} {
		t.Run("", func(t *testing.T) {
			i, exact := ticks.FindPrice(tc.price)
			require.Equal(t, tc.i, i)
			require.Equal(t, tc.exact, exact)
		})
	}
}

func TestOrderBookTicks_AddOrder(t *testing.T) {
	checkSorted := func(ticks *types.OrderBookTicks) {
		require.True(t, sort.SliceIsSorted(ticks, func(i, j int) bool {
			return ticks.Ticks[i].Price.GTE(ticks.Ticks[j].Price)
		}), "ticks must be sorted")
	}

	ticks := testOrderBookTicks()
	checkSorted(ticks)
	require.Len(t, ticks, 11)

	// Same price already exists
	ticks.AddOrder(newBuyOrder("18.0", 1000))
	checkSorted(ticks)
	require.Len(t, ticks, 11)

	// New price. We don't care about the tick precision here
	ticks.AddOrder(newBuyOrder("18.000000000000000001", 1000))
	checkSorted(ticks)
	require.Len(t, ticks, 12)

	// Add an order with same price as above again
	ticks.AddOrder(newBuyOrder("18.000000000000000001", 1000))
	checkSorted(ticks)
	require.Len(t, ticks, 12)

	// Add an order with higher price than the highest price in ticks.
	ticks.AddOrder(newBuyOrder("21.0", 1000))
	checkSorted(ticks)
	require.Len(t, ticks, 13)

	// Add an order with lower price than the lowest price in ticks.
	ticks.AddOrder(newBuyOrder("9.0", 1000))
	checkSorted(ticks)
	require.Len(t, ticks, 14)
}

func TestOrderBookTicks_AmountGTE(t *testing.T) {
	// An empty order book ticks
	require.True(sdk.IntEq(t, sdk.ZeroInt(), types.NewOrderBookTicks(tickPrec).AmountGTE(newDec("20.0"))))

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Int
	}{
		{newDec("20.000000000000000001"), sdk.ZeroInt()},
		{newDec("20.0"), sdk.NewInt(1000)},
		{newDec("19.999999999999999999"), sdk.NewInt(1000)},
		{newDec("19.000000000000000001"), sdk.NewInt(1000)},
		{newDec("19.0"), sdk.NewInt(2000)},
		{newDec("9.999999999999999999"), sdk.NewInt(11000)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.expected, ticks.AmountGTE(tc.price)))
		})
	}
}

func TestOrderBookTicks_AmountLTE(t *testing.T) {
	// An empty order book ticks
	require.True(sdk.IntEq(t, sdk.ZeroInt(), types.OrderBookTicks{}.AmountLTE(newDec("20.0"))))

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Int
	}{
		{newDec("20.000000000000000001"), sdk.NewInt(11000)},
		{newDec("20.0"), sdk.NewInt(11000)},
		{newDec("19.999999999999999999"), sdk.NewInt(10000)},
		{newDec("19.000000000000000001"), sdk.NewInt(10000)},
		{newDec("19.0"), sdk.NewInt(10000)},
		{newDec("9.999999999999999999"), sdk.ZeroInt()},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.expected, ticks.AmountLTE(tc.price)))
		})
	}
}

func TestOrderBookTicks_Orders(t *testing.T) {
	ticks := types.OrderBookTicks{}

	orderMap := map[string]types.Orders{
		"20.0": {newBuyOrder("20.0", 1000), newBuyOrder("20.0", 1000)},
		"19.0": {newBuyOrder("19.0", 500), newBuyOrder("19.0", 1000)},
		"18.0": {newBuyOrder("18.0", 1000)},
		"17.0": {newBuyOrder("17.0", 1000), newBuyOrder("17.0", 2000)},
	}

	for _, orders := range orderMap {
		ticks.AddOrders(orders...)
	}

	// Price not found
	require.Len(t, ticks.Orders(newDec("100.0")), 0)

	for price, orders := range orderMap {
		orders2 := ticks.Orders(newDec(price))
		require.Len(t, orders2, len(orders))
		for i := range orders {
			ok := false
			for j := range orders2 {
				if orders[i] == orders2[j] {
					ok = true
					break
				}
			}
			require.True(t, ok)
		}
	}
}

func TestOrderBookTicks_UpTickWithOrders(t *testing.T) {
	// An empty order book ticks
	_, found := types.NewOrderBookTicks(tickPrec).UpTick(newDec("0.1"))
	require.False(t, found)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		tick  sdk.Dec
		found bool
	}{
		{newDec("20.000000000000000001"), sdk.Dec{}, false},
		{newDec("20.0"), sdk.Dec{}, false},
		{newDec("19.999999999999999999"), newDec("20.0"), true},
		{newDec("19.000000000000000001"), newDec("20.0"), true},
		{newDec("19.0"), newDec("20.0"), true},
		{newDec("18.999999999999999999"), newDec("19.0"), true},
		{newDec("18.000000000000000001"), newDec("19.0"), true},
		{newDec("18.0"), newDec("19.0"), true},
		{newDec("10.0"), newDec("11.0"), true},
		{newDec("9.999999999999999999"), newDec("10.0"), true},
	} {
		t.Run("", func(t *testing.T) {
			tick, found := ticks.UpTickWithOrders(tc.price)
			require.Equal(t, tc.found, found)
			if found {
				require.True(sdk.DecEq(t, tc.tick, tick))
			}
		})
	}
}

func TestOrderBookTicks_DownTickWithOrders(t *testing.T) {
	// An empty order book ticks
	_, found := types.NewOrderBookTicks(tickPrec).UpTick(newDec("0.1"))
	require.False(t, found)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		tick  sdk.Dec
		found bool
	}{
		{newDec("20.000000000000000001"), newDec("20.0"), true},
		{newDec("20.0"), newDec("19.0"), true},
		{newDec("19.999999999999999999"), newDec("19.0"), true},
		{newDec("19.000000000000000001"), newDec("19.0"), true},
		{newDec("19.0"), newDec("18.0"), true},
		{newDec("18.999999999999999999"), newDec("18.0"), true},
		{newDec("18.000000000000000001"), newDec("18.0"), true},
		{newDec("18.0"), newDec("17.0"), true},
		{newDec("10.000000000000000001"), newDec("10.0"), true},
		{newDec("10.0"), sdk.Dec{}, false},
		{newDec("9.999999999999999999"), sdk.Dec{}, false},
	} {
		t.Run("", func(t *testing.T) {
			tick, found := ticks.DownTickWithOrders(tc.price)
			require.Equal(t, tc.found, found)
			if found {
				require.True(sdk.DecEq(t, tc.tick, tick))
			}
		})
	}
}

func TestOrderBookTicks_HighestTick(t *testing.T) {
	// An empty order book ticks
	_, found := types.OrderBookTicks{}.HighestTick()
	require.False(t, found)

	ticks := testOrderBookTicks()
	tick, found := ticks.HighestTick()
	require.True(t, found)
	require.True(sdk.DecEq(t, newDec("20.0"), tick))
}

func TestOrderBookTicks_LowestTick(t *testing.T) {
	// An empty order book ticks
	_, found := types.OrderBookTicks{}.LowestTick()
	require.False(t, found)

	ticks := testOrderBookTicks()
	tick, found := ticks.LowestTick()
	require.True(t, found)
	require.True(sdk.DecEq(t, newDec("10.0"), tick))
}
