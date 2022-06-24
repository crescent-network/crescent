package types_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func newOrder(dir amm.OrderDirection, price sdk.Dec, amt sdk.Int) amm.Order {
	return amm.DefaultOrderer.Order(dir, price, amt)
}

func ExampleMakeOrderBookResponse() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("15.0"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("13.0"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("10.01"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("10.0"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("10.01"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("10.0"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("9.0"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("3.0"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("0.1"), sdk.NewInt(10000)),
	)
	ov := ob.MakeView()
	ov.Match()
	tickPrec := 1
	resp := types.MakeOrderBookResponse(ov, tickPrec, 20)
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
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

func ExampleMakeOrderBookResponse_pool() {
	pool1 := amm.NewBasicPool(sdk.NewInt(1050_000000), sdk.NewInt(1000_000000), sdk.Int{})
	pool2, err := amm.CreateRangedPool(
		sdk.NewInt(1000_000000), sdk.NewInt(1000_000000),
		utils.ParseDec("0.1"), utils.ParseDec("5.0"), utils.ParseDec("0.95"))
	if err != nil {
		panic(err)
	}

	lastPrice := utils.ParseDec("1.0")
	lowestPrice := lastPrice.Mul(utils.ParseDec("0.9"))
	highestPrice := lastPrice.Mul(utils.ParseDec("1.1"))
	ob := amm.NewOrderBook()
	ob.AddOrder(amm.PoolOrders(pool1, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)
	ob.AddOrder(amm.PoolOrders(pool2, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)

	ov := ob.MakeView()
	ov.Match()
	tickPrec := 2
	resp := types.MakeOrderBookResponse(ov, tickPrec, 20)
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |           12482141 |         1.100000000000000000 |                    |
	// |           11488898 |         1.090000000000000000 |                    |
	// |            8280038 |         1.080000000000000000 |                    |
	// |           12100493 |         1.070000000000000000 |                    |
	// |           12139873 |         1.060000000000000000 |                    |
	// |            7381973 |         1.050000000000000000 |                    |
	// |            6623054 |         1.040000000000000000 |                    |
	// |            8863846 |         1.030000000000000000 |                    |
	// |            7489932 |         1.020000000000000000 |                    |
	// |            6572362 |         1.010000000000000000 |                    |
	// |            8736758 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.990000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.980000000000000000 | 3406177            |
	// |                    |         0.970000000000000000 | 5349384            |
	// |                    |         0.960000000000000000 | 6256722            |
	// |                    |         0.950000000000000000 | 4655698            |
	// |                    |         0.940000000000000000 | 13393535           |
	// |                    |         0.930000000000000000 | 13541701           |
	// |                    |         0.920000000000000000 | 17097505           |
	// |                    |         0.910000000000000000 | 14039668           |
	// |                    |         0.900000000000000000 | 14822646           |
	// +------------------------------------------------------------------------+
}

func TestMakeOrderBookResponse(t *testing.T) {
	pool := amm.NewBasicPool(sdk.NewInt(862431695563), sdk.NewInt(37852851767), sdk.Int{})
	lowestPrice := pool.Price().Mul(sdk.NewDecWithPrec(9, 1))
	highestPrice := pool.Price().Mul(sdk.NewDecWithPrec(11, 1))

	ob := amm.NewOrderBook()
	ob.AddOrder(amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)

	ov := ob.MakeView()
	basePrice, found := types.OrderBookBasePrice(ov, 4)
	if !found {
		panic("base price not found")
	}

	resp := types.MakeOrderBookResponse(ov, 3, 20)
	types.PrintOrderBookResponse(resp, basePrice)
}

func BenchmarkMakeOrderBookResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		makeOrderBookPairResponse(100, 10, 20, 4)
	}
}

func makeOrderBookPairResponse(numOrders, numPools, numTicks, tickPrec int) *types.OrderBookPairResponse {
	r := rand.New(rand.NewSource(0))
	ob := amm.NewOrderBook()
	for i := 0; i < numOrders; i++ {
		var dir amm.OrderDirection
		if r.Intn(2) == 0 {
			dir = amm.Buy
		} else {
			dir = amm.Sell
		}
		price := amm.PriceToDownTick(
			utils.RandomDec(r, utils.ParseDec("0.5"), utils.ParseDec("1.5")), tickPrec)
		amt := utils.RandomInt(r, sdk.NewInt(1000), sdk.NewInt(100000))
		ob.AddOrder(newOrder(dir, price, amt))
	}

	for i := 0; i < numPools; i++ {
		rx := utils.RandomInt(r, sdk.NewInt(10000_000000), sdk.NewInt(11000_000000))
		ry := utils.RandomInt(r, sdk.NewInt(10000_000000), sdk.NewInt(11000_000000))
		pool := amm.NewBasicPool(rx, ry, sdk.Int{})
		ob.AddOrder(amm.PoolOrders(pool, amm.DefaultOrderer, utils.ParseDec("0.9"), utils.ParseDec("1.1"), tickPrec)...)
	}

	ov := ob.MakeView()
	ov.Match()

	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}

	resp := &types.OrderBookPairResponse{
		PairId:    1,
		BasePrice: basePrice,
	}
	for _, tickPrec := range []int{1, 2, 3, 4} {
		resp.OrderBooks = append(resp.OrderBooks, types.MakeOrderBookResponse(ov, tickPrec, numTicks))
	}
	return resp
}
