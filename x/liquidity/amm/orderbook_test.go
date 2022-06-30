package amm_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

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

	highest, found := ob.HighestPrice()
	require.True(t, found)
	require.True(sdk.DecEq(t, utils.ParseDec("10.01"), highest))
	lowest, found := ob.LowestPrice()
	require.True(t, found)
	require.True(sdk.DecEq(t, utils.ParseDec("9.996"), lowest))
}

func TestOrderBook_BuyOrdersAt(t *testing.T) {
	order1 := newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order2 := newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order3 := newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order4 := newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000))
	order5 := newOrder(amm.Buy, utils.ParseDec("1.2"), sdk.NewInt(10000))

	ob := amm.NewOrderBook(order1, order2, order3, order4, order5)
	buyOrders := ob.BuyOrdersAt(utils.ParseDec("1.1"))

	require.Len(t, buyOrders, 2)
	require.Contains(t, buyOrders, order2)
	require.Contains(t, buyOrders, order3)

	require.Empty(t, ob.BuyOrdersAt(utils.ParseDec("0.9")))
}

func TestOrderBook_SellOrdersAt(t *testing.T) {
	order1 := newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order2 := newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order3 := newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000))
	order4 := newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000))
	order5 := newOrder(amm.Sell, utils.ParseDec("1.2"), sdk.NewInt(10000))

	ob := amm.NewOrderBook(order1, order2, order3, order4, order5)
	sellOrders := ob.SellOrdersAt(utils.ParseDec("1.1"))

	require.Len(t, sellOrders, 2)
	require.Contains(t, sellOrders, order2)
	require.Contains(t, sellOrders, order3)

	require.Empty(t, ob.SellOrdersAt(utils.ParseDec("0.9")))
}

func TestOrderBook_HighestPriceLowestPrice(t *testing.T) {
	for _, tc := range []struct {
		ob           *amm.OrderBook
		found        bool
		highestPrice sdk.Dec
		lowestPrice  sdk.Dec
	}{
		{
			amm.NewOrderBook(),
			false, sdk.Dec{}, sdk.Dec{},
		},
		{
			amm.NewOrderBook(
				newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			),
			true, utils.ParseDec("1.1"), utils.ParseDec("1.0"),
		},
		{
			amm.NewOrderBook(
				newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
			),
			true, utils.ParseDec("1.1"), utils.ParseDec("1.0"),
		},
		{
			amm.NewOrderBook(
				newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
				newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
				newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
				newOrder(amm.Buy, utils.ParseDec("0.9"), sdk.NewInt(10000)),
			),
			true, utils.ParseDec("1.1"), utils.ParseDec("0.9"),
		},
	} {
		t.Run("", func(t *testing.T) {
			highestPrice, foundHighestPrice := tc.ob.HighestPrice()
			require.Equal(t, tc.found, foundHighestPrice)
			lowestPrice, foundLowestPrice := tc.ob.LowestPrice()
			require.Equal(t, tc.found, foundLowestPrice)
			if tc.found {
				require.True(sdk.DecEq(t, tc.highestPrice, highestPrice))
				require.True(sdk.DecEq(t, tc.lowestPrice, lowestPrice))
			}
		})
	}
}

func ExampleOrderBook_String() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.2"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("1.17"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.15"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(5000)),
		newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(1000)),
		newOrder(amm.Sell, utils.ParseDec("1.09"), sdk.NewInt(6000)),
		newOrder(amm.Sell, utils.ParseDec("1.09"), sdk.NewInt(4000)),
		newOrder(amm.Buy, utils.ParseDec("1.06"), sdk.NewInt(15000)),
	)
	fmt.Println(ob.String())

	// Output:
	// +--------sell--------+------------price-------------+--------buy---------+
	// |              10000 |         1.200000000000000000 | 0                  |
	// |                  0 |         1.170000000000000000 | 10000              |
	// |               5000 |         1.150000000000000000 | 0                  |
	// |               3000 |         1.100000000000000000 | 6000               |
	// |              10000 |         1.090000000000000000 | 0                  |
	// |                  0 |         1.060000000000000000 | 15000              |
	// +--------------------+------------------------------+--------------------+
}

func ExampleOrderBook_FullString() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.2"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("1.17"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.15"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(5000)),
		newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(1000)),
		newOrder(amm.Sell, utils.ParseDec("1.09"), sdk.NewInt(6000)),
		newOrder(amm.Sell, utils.ParseDec("1.09"), sdk.NewInt(4000)),
		newOrder(amm.Buy, utils.ParseDec("1.06"), sdk.NewInt(15000)),
	)
	fmt.Println(ob.FullString(2))

	// Output:
	// +--------sell--------+------------price-------------+--------buy---------+
	// |              10000 |         1.200000000000000000 | 0                  |
	// |                  0 |         1.190000000000000000 | 0                  |
	// |                  0 |         1.180000000000000000 | 0                  |
	// |                  0 |         1.170000000000000000 | 10000              |
	// |                  0 |         1.160000000000000000 | 0                  |
	// |               5000 |         1.150000000000000000 | 0                  |
	// |                  0 |         1.140000000000000000 | 0                  |
	// |                  0 |         1.130000000000000000 | 0                  |
	// |                  0 |         1.120000000000000000 | 0                  |
	// |                  0 |         1.110000000000000000 | 0                  |
	// |               3000 |         1.100000000000000000 | 6000               |
	// |              10000 |         1.090000000000000000 | 0                  |
	// |                  0 |         1.080000000000000000 | 0                  |
	// |                  0 |         1.070000000000000000 | 0                  |
	// |                  0 |         1.060000000000000000 | 15000              |
	// +--------------------+------------------------------+--------------------+
}

func BenchmarkOrderBook_AddOrder(b *testing.B) {
	/*
		Before optimization:
		BenchmarkOrderBook_AddOrder/1000_orders
		BenchmarkOrderBook_AddOrder/1000_orders-8         	    2983	    419432 ns/op
		BenchmarkOrderBook_AddOrder/5000_orders
		BenchmarkOrderBook_AddOrder/5000_orders-8         	     463	   2851024 ns/op
		BenchmarkOrderBook_AddOrder/10000_orders
		BenchmarkOrderBook_AddOrder/10000_orders-8        	     204	   6380550 ns/op

		After optimization:
		BenchmarkOrderBook_AddOrder/1000_orders
		BenchmarkOrderBook_AddOrder/1000_orders-8         	    5949	    203521 ns/op
		BenchmarkOrderBook_AddOrder/5000_orders
		BenchmarkOrderBook_AddOrder/5000_orders-8         	     745	   1714147 ns/op
		BenchmarkOrderBook_AddOrder/10000_orders
		BenchmarkOrderBook_AddOrder/10000_orders-8        	     252	   4121103 ns/op
	*/
	r := rand.New(rand.NewSource(0))
	orders := make([]amm.Order, 10000)
	for i := range orders {
		var dir amm.OrderDirection
		if r.Intn(2) == 0 {
			dir = amm.Buy
		} else {
			dir = amm.Sell
		}
		price := defTickPrec.PriceToDownTick(
			utils.RandomDec(r, utils.ParseDec("0.01"), utils.ParseDec("100.0")))
		orders[i] = newOrder(dir, price, sdk.NewInt(10000))
	}
	for _, numOrders := range []int{1000, 5000, 10000} {
		b.Run(fmt.Sprintf("%d orders", numOrders), func(b *testing.B) {
			ob := amm.NewOrderBook()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ob.AddOrder(orders[:numOrders]...)
			}
		})
	}
}
