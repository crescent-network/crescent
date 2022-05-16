package types_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func ExampleMakeOrderBookResponse() {
	ob := amm.NewOrderBook(
		newUserOrder(amm.Sell, 1, utils.ParseDec("15.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Sell, 2, utils.ParseDec("13.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Sell, 3, utils.ParseDec("10.01"), sdk.NewInt(10000)),
		newUserOrder(amm.Sell, 4, utils.ParseDec("10.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Buy, 5, utils.ParseDec("10.01"), sdk.NewInt(10000)),
		newUserOrder(amm.Buy, 6, utils.ParseDec("10.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Buy, 7, utils.ParseDec("9.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Buy, 8, utils.ParseDec("3.0"), sdk.NewInt(10000)),
		newUserOrder(amm.Buy, 9, utils.ParseDec("0.1"), sdk.NewInt(10000)),
	)
	tickPrec := 1
	basePrice, found := types.OrderBookBasePrice(ob, tickPrec)
	if !found {
		panic("base price not found")
	}
	resp := types.MakeOrderBookResponse(ob, basePrice, tickPrec, 20)
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |              10000 |        15.000000000000000000 |                    |
	// |              10000 |        13.000000000000000000 |                    |
	// |              10000 |        11.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                             10.000000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |        10.000000000000000000 | 10000              |
	// |                    |         9.000000000000000000 | 10000              |
	// |                    |         3.000000000000000000 | 10000              |
	// |                    |         0.000000000000000000 | 10000              |
	// +------------------------------------------------------------------------+
}

func BenchmarkMakeOrderBookResponse(b *testing.B) {
	const numOrders = 5000

	r := rand.New(rand.NewSource(0))

	ob := amm.NewOrderBook()
	for i := 0; i < numOrders; i++ {
		var dir amm.OrderDirection
		if r.Intn(2) == 0 {
			dir = amm.Buy
		} else {
			dir = amm.Sell
		}
		price := utils.RandomDec(r, utils.ParseDec("0.5"), utils.ParseDec("1.5"))
		amt := utils.RandomInt(r, sdk.NewInt(1000), sdk.NewInt(100000))
		ob.Add(newUserOrder(dir, uint64(i+1), price, amt))
	}
	pool := amm.NewBasicPool(sdk.NewInt(10000_000000), sdk.NewInt(10000_000000), sdk.Int{})
	ov := amm.MergeOrderViews(ob, pool)

	tickPrec := 1
	numTicks := 20
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	require.True(b, found)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		types.MakeOrderBookResponse(ov, basePrice, tickPrec, numTicks)
	}
}
