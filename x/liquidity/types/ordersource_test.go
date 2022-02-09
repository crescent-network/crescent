package types_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func testOrderBookTicks() *types.OrderBookTicks {
	ticks := types.NewOrderBookTicks(tickPrec)
	ticks.AddOrders(
		newBuyOrder(parseDec("20.0"), newInt(1000)),
		newBuyOrder(parseDec("19.0"), newInt(1000)),
		newBuyOrder(parseDec("18.0"), newInt(1000)),
		newBuyOrder(parseDec("17.0"), newInt(1000)),
		newBuyOrder(parseDec("16.0"), newInt(1000)),
		newBuyOrder(parseDec("15.0"), newInt(1000)).SetOpenAmount(sdk.ZeroInt()),
		newBuyOrder(parseDec("14.0"), newInt(1000)),
		newBuyOrder(parseDec("13.0"), newInt(1000)).SetOpenAmount(sdk.ZeroInt()),
		newBuyOrder(parseDec("12.0"), newInt(1000)),
		newBuyOrder(parseDec("11.0"), newInt(1000)),
		newBuyOrder(parseDec("10.0"), newInt(1000)),
	)
	return ticks
}

func TestOrderBookTicks_FindPrice(t *testing.T) {
	// An empty order book ticks must return (0, false).
	i, exact := types.NewOrderBookTicks(tickPrec).FindPrice(parseDec("20.0"))
	require.False(t, exact)
	require.Equal(t, 0, i)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		i     int
		exact bool
	}{
		{parseDec("20.0"), 0, true},
		{parseDec("19.99999999999999999"), 1, false},
		{parseDec("19.00000000000000001"), 1, false},
		{parseDec("19.0"), 1, true},
		{parseDec("18.99999999999999999"), 2, false},
		{parseDec("18.00000000000000001"), 2, false},
		{parseDec("18.0"), 2, true},
		{parseDec("9.999999999999999999"), 11, false},
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
		require.True(t, sort.SliceIsSorted(ticks.Ticks, func(i, j int) bool {
			return ticks.Ticks[i].Price.GTE(ticks.Ticks[j].Price)
		}), "ticks must be sorted")
	}

	ticks := testOrderBookTicks()
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 11)

	// Same price already exists
	ticks.AddOrder(newBuyOrder(parseDec("18.0"), newInt(1000)))
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 11)

	// New price. We don't care about the tick precision here
	ticks.AddOrder(newBuyOrder(parseDec("18.000000000000000001"), newInt(1000)))
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 12)

	// Add an order with same price as above again
	ticks.AddOrder(newBuyOrder(parseDec("18.000000000000000001"), newInt(1000)))
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 12)

	// Add an order with higher price than the highest price in ticks.
	ticks.AddOrder(newBuyOrder(parseDec("21.0"), newInt(1000)))
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 13)

	// Add an order with lower price than the lowest price in ticks.
	ticks.AddOrder(newBuyOrder(parseDec("9.0"), newInt(1000)))
	checkSorted(ticks)
	require.Len(t, ticks.Ticks, 14)
}

func TestOrderBookTicks_AmountGTE(t *testing.T) {
	// An empty order book ticks
	require.True(sdk.IntEq(t, sdk.ZeroInt(), types.NewOrderBookTicks(tickPrec).AmountGTE(parseDec("20.0"))))

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Int
	}{
		{parseDec("20.000000000000000001"), sdk.ZeroInt()},
		{parseDec("20.0"), sdk.NewInt(1000)},
		{parseDec("19.999999999999999999"), sdk.NewInt(1000)},
		{parseDec("19.000000000000000001"), sdk.NewInt(1000)},
		{parseDec("19.0"), sdk.NewInt(2000)},
		{parseDec("9.999999999999999999"), sdk.NewInt(9000)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.expected, ticks.AmountGTE(tc.price)))
		})
	}
}

func TestOrderBookTicks_AmountLTE(t *testing.T) {
	// An empty order book ticks
	require.True(sdk.IntEq(t, sdk.ZeroInt(), types.OrderBookTicks{}.AmountLTE(parseDec("20.0"))))

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Int
	}{
		{parseDec("20.000000000000000001"), sdk.NewInt(9000)},
		{parseDec("20.0"), sdk.NewInt(9000)},
		{parseDec("19.999999999999999999"), sdk.NewInt(8000)},
		{parseDec("19.000000000000000001"), sdk.NewInt(8000)},
		{parseDec("19.0"), sdk.NewInt(8000)},
		{parseDec("9.999999999999999999"), sdk.ZeroInt()},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.expected, ticks.AmountLTE(tc.price)))
		})
	}
}

func TestOrderBookTicks_Orders(t *testing.T) {
	ticks := types.OrderBookTicks{}

	orderMap := map[string]types.Orders{
		"20.0": {newBuyOrder(parseDec("20.0"), newInt(1000)), newBuyOrder(parseDec("20.0"), newInt(1000))},
		"19.0": {newBuyOrder(parseDec("19.0"), newInt(500)), newBuyOrder(parseDec("19.0"), newInt(1000))},
		"18.0": {newBuyOrder(parseDec("18.0"), newInt(1000))},
		"17.0": {newBuyOrder(parseDec("17.0"), newInt(1000)), newBuyOrder(parseDec("17.0"), newInt(2000))},
	}

	for _, orders := range orderMap {
		ticks.AddOrders(orders...)
	}

	// Price not found
	require.Len(t, ticks.Orders(parseDec("100.0")), 0)

	for price, orders := range orderMap {
		orders2 := ticks.Orders(parseDec(price))
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
	_, found := types.NewOrderBookTicks(tickPrec).UpTick(parseDec("0.1"))
	require.False(t, found)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		tick  sdk.Dec
		found bool
	}{
		{parseDec("20.000000000000000001"), sdk.Dec{}, false},
		{parseDec("20.0"), sdk.Dec{}, false},
		{parseDec("19.999999999999999999"), parseDec("20.0"), true},
		{parseDec("19.000000000000000001"), parseDec("20.0"), true},
		{parseDec("19.0"), parseDec("20.0"), true},
		{parseDec("18.999999999999999999"), parseDec("19.0"), true},
		{parseDec("18.000000000000000001"), parseDec("19.0"), true},
		{parseDec("18.0"), parseDec("19.0"), true},
		{parseDec("14.999999999999999999"), parseDec("16.0"), true},
		{parseDec("10.0"), parseDec("11.0"), true},
		{parseDec("9.999999999999999999"), parseDec("10.0"), true},
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
	_, found := types.NewOrderBookTicks(tickPrec).UpTick(parseDec("0.1"))
	require.False(t, found)

	ticks := testOrderBookTicks()

	for _, tc := range []struct {
		price sdk.Dec
		tick  sdk.Dec
		found bool
	}{
		{parseDec("20.000000000000000001"), parseDec("20.0"), true},
		{parseDec("20.0"), parseDec("19.0"), true},
		{parseDec("19.999999999999999999"), parseDec("19.0"), true},
		{parseDec("19.000000000000000001"), parseDec("19.0"), true},
		{parseDec("19.0"), parseDec("18.0"), true},
		{parseDec("18.999999999999999999"), parseDec("18.0"), true},
		{parseDec("18.000000000000000001"), parseDec("18.0"), true},
		{parseDec("18.0"), parseDec("17.0"), true},
		{parseDec("15.000000000000000001"), parseDec("14.0"), true},
		{parseDec("10.000000000000000001"), parseDec("10.0"), true},
		{parseDec("10.0"), sdk.Dec{}, false},
		{parseDec("9.999999999999999999"), sdk.Dec{}, false},
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
	require.True(sdk.DecEq(t, parseDec("20.0"), tick))

	// Test with orders with zero remaining amount
	ticks = types.NewOrderBookTicks(tickPrec)
	ticks.AddOrders(
		newBuyOrder(parseDec("10.0"), newInt(1000)).SetOpenAmount(sdk.ZeroInt()),
		newBuyOrder(parseDec("9.0"), newInt(1000)),
		newBuyOrder(parseDec("8.0"), newInt(1000)),
	)

	tick, found = ticks.HighestTick()
	require.True(t, found)
	require.True(sdk.DecEq(t, parseDec("9.0"), tick))
}

func TestOrderBookTicks_LowestTick(t *testing.T) {
	// An empty order book ticks
	_, found := types.OrderBookTicks{}.LowestTick()
	require.False(t, found)

	ticks := testOrderBookTicks()
	tick, found := ticks.LowestTick()
	require.True(t, found)
	require.True(sdk.DecEq(t, parseDec("10.0"), tick))

	// Test with orders with zero remaining amount
	ticks = types.NewOrderBookTicks(tickPrec)
	ticks.AddOrders(
		newBuyOrder(parseDec("10.0"), newInt(1000)),
		newBuyOrder(parseDec("9.0"), newInt(1000)),
		newBuyOrder(parseDec("8.0"), newInt(1000)).SetOpenAmount(sdk.ZeroInt()),
	)

	tick, found = ticks.LowestTick()
	require.True(t, found)
	require.True(sdk.DecEq(t, parseDec("9.0"), tick))
}
