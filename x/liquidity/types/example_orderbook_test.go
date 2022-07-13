package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
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
	lowestPrice := utils.ParseDec("0")
	highestPrice := utils.ParseDec("20")
	resp := types.MakeOrderBookResponse(ov, lowestPrice, highestPrice, tickPrec, 20)
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
	resp := types.MakeOrderBookResponse(ov, lowestPrice, highestPrice, tickPrec, 10)
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |            2299835 |         1.011000000000000000 |                    |
	// |            2342262 |         1.008000000000000000 |                    |
	// |            2170367 |         1.005000000000000000 |                    |
	// |            2059733 |         1.003000000000000000 |                    |
	// |            1914496 |         1.000000000000000000 |                    |
	// |            1846587 |         0.998000000000000000 |                    |
	// |            1729430 |         0.995000000000000000 |                    |
	// |            1674921 |         0.993000000000000000 |                    |
	// |            1571324 |         0.991000000000000000 |                    |
	// |            1307011 |         0.989000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.988000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.988000000000000000 | 1521174            |
	// |                    |         0.985000000000000000 | 1520420            |
	// |                    |         0.982000000000000000 | 1671594            |
	// |                    |         0.979000000000000000 | 1672618            |
	// |                    |         0.975000000000000000 | 1836801            |
	// |                    |         0.972000000000000000 | 1839965            |
	// |                    |         0.968000000000000000 | 2017741            |
	// |                    |         0.965000000000000000 | 2023466            |
	// |                    |         0.961000000000000000 | 2215515            |
	// |                    |         0.957000000000000000 | 2224296            |
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
	resp := types.MakeOrderBookResponse(ov, lowestPrice, highestPrice, tickPrec, 10)
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
	resp := types.MakeOrderBookResponse(ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), tickPrec, 10)
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

func ExampleMakeOrderBookResponse_zigzag() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.002"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("1.001"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(50000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(100000)),
	)
	ov := ob.MakeView()
	ov.Match()

	basePrice, _ := types.OrderBookBasePrice(ov, 4)
	resp := types.MakeOrderBookResponse(ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, 20)
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |              10000 |         1.002000000000000000 |                    |
	// |              45000 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999500000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 100000             |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookResponse_edgecase1() {
	basicPool := amm.NewBasicPool(sdk.NewInt(2603170018), sdk.NewInt(2731547352), sdk.Int{})
	rangedPool := amm.NewRangedPool(sdk.NewInt(9204969), sdk.NewInt(292104465), sdk.Int{}, utils.ParseDec("0.95"), utils.ParseDec("1.05"))
	lastPrice := utils.ParseDec("0.95299")

	ob := amm.NewOrderBook()
	lowestPrice := lastPrice.Mul(utils.ParseDec("0.9"))
	highestPrice := lastPrice.Mul(utils.ParseDec("1.1"))
	ob.AddOrder(amm.PoolOrders(basicPool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)
	ob.AddOrder(amm.PoolOrders(rangedPool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)

	ov := ob.MakeView()
	ov.Match()

	basePrice, _ := types.OrderBookBasePrice(ov, 4)
	resp := types.MakeOrderBookResponse(ov, lowestPrice, highestPrice, 3, 10)
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |             424830 |         0.953900000000000000 |                    |
	// |             502738 |         0.953800000000000000 |                    |
	// |             436285 |         0.953700000000000000 |                    |
	// |             502809 |         0.953600000000000000 |                    |
	// |             425098 |         0.953500000000000000 |                    |
	// |             503055 |         0.953400000000000000 |                    |
	// |             436559 |         0.953300000000000000 |                    |
	// |             497509 |         0.953200000000000000 |                    |
	// |             495757 |         0.953100000000000000 |                    |
	// |              59171 |         0.953000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.952980000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.952900000000000000 | 425397             |
	// |                    |         0.952800000000000000 | 415669             |
	// |                    |         0.952700000000000000 | 518993             |
	// |                    |         0.952600000000000000 | 503626             |
	// |                    |         0.952500000000000000 | 408371             |
	// |                    |         0.952400000000000000 | 503685             |
	// |                    |         0.952300000000000000 | 454527             |
	// |                    |         0.952200000000000000 | 503942             |
	// |                    |         0.952100000000000000 | 408638             |
	// |                    |         0.952000000000000000 | 504004             |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookResponse_edgecase2() {
	ob := amm.NewOrderBook(
		newOrder(amm.Buy, utils.ParseDec("1.001"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("0.999"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("0.998"), sdk.NewInt(3000)),
		newOrder(amm.Sell, utils.ParseDec("0.997"), sdk.NewInt(2000)),
	)

	ov := ob.MakeView()
	ov.Match()

	basePrice, _ := types.OrderBookBasePrice(ov, 4)
	resp := types.MakeOrderBookResponse(ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, 10)
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |------------------------------------------------------------------------|
	// |                              1.000000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         1.000000000000000000 | 3000               |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookResponse_edgecase3() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.001"), sdk.NewInt(2000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(2000)),
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(2000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(2000)),
	)

	ov := ob.MakeView()
	ov.Match()

	basePrice, _ := types.OrderBookBasePrice(ov, 4)
	resp := types.MakeOrderBookResponse(ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, 10)
	types.PrintOrderBookResponse(resp, basePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |               2000 |         1.001000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              1.000000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 2000               |
	// +------------------------------------------------------------------------+
}
