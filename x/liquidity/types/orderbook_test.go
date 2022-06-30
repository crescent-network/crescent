package types_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func newOrder(dir amm.OrderDirection, price sdk.Dec, amt sdk.Int) amm.Order {
	return amm.DefaultOrderer.Order(dir, price, amt)
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
