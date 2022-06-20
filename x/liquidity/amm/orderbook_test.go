package amm_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
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

func BenchmarkOrderBook_AddOrder(b *testing.B) {
	/*
		BenchmarkOrderBook_AddOrder/1000_orders
		BenchmarkOrderBook_AddOrder/1000_orders-8         	    2983	    419432 ns/op
		BenchmarkOrderBook_AddOrder/5000_orders
		BenchmarkOrderBook_AddOrder/5000_orders-8         	     463	   2851024 ns/op
		BenchmarkOrderBook_AddOrder/10000_orders
		BenchmarkOrderBook_AddOrder/10000_orders-8        	     204	   6380550 ns/op
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
