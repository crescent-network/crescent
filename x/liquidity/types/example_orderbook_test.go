package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

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

func ExampleMakeOrderBookResponse_userOrder() {
	pool := amm.NewBasicPool(sdk.NewInt(895590740832), sdk.NewInt(675897553075), sdk.Int{})

	lastPrice := utils.ParseDec("1.325")
	lowestPrice := lastPrice.Mul(utils.ParseDec("0.9"))
	highestPrice := lastPrice.Mul(utils.ParseDec("1.1"))
	ob := amm.NewOrderBook()
	ob.AddOrder(amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)
	ob.AddOrder(
		newOrder(amm.Buy, utils.ParseDec("1.316"), sdk.NewInt(111000000)),
		newOrder(amm.Buy, utils.ParseDec("1.32"), sdk.NewInt(111000000)),
		newOrder(amm.Buy, utils.ParseDec("1.325"), sdk.NewInt(11000000)),
		newOrder(amm.Buy, utils.ParseDec("1.325"), sdk.NewInt(111000000)),
		newOrder(amm.Buy, utils.ParseDec("1.325"), sdk.NewInt(20000000)),
		newOrder(amm.Buy, utils.ParseDec("1.325"), sdk.NewInt(111000000)),
	)

	ov := ob.MakeView()
	ov.Match()
	tickPrec := 3
	resp := types.MakeOrderBookResponse(ov, tickPrec, 10)
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |          283061445 |         1.335000000000000000 |                    |
	// |          252607703 |         1.334000000000000000 |                    |
	// |          252895894 |         1.333000000000000000 |                    |
	// |          253180839 |         1.332000000000000000 |                    |
	// |          253466318 |         1.331000000000000000 |                    |
	// |          253752331 |         1.330000000000000000 |                    |
	// |          254038885 |         1.329000000000000000 |                    |
	// |          254325976 |         1.328000000000000000 |                    |
	// |          254613613 |         1.327000000000000000 |                    |
	// |          254901787 |         1.326000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              1.326000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         1.325000000000000000 | 272987175          |
	// |                    |         1.324000000000000000 | 255217324          |
	// |                    |         1.323000000000000000 | 255506641          |
	// |                    |         1.322000000000000000 | 255796508          |
	// |                    |         1.321000000000000000 | 256086923          |
	// |                    |         1.320000000000000000 | 367377883          |
	// |                    |         1.319000000000000000 | 256669399          |
	// |                    |         1.318000000000000000 | 256961468          |
	// |                    |         1.317000000000000000 | 257254090          |
	// |                    |         1.316000000000000000 | 368543354          |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookResponse_match() {
	ob := amm.NewOrderBook(
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.001"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(10000)),
	)

	ov := ob.MakeView()
	ov.Match()
	tickPrec := 3
	resp := types.MakeOrderBookResponse(ov, tickPrec, 10)
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |              10000 |         1.001000000000000000 |                    |
	// |               7000 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999500000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 10000              |
	// +------------------------------------------------------------------------+
}
