package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func ExampleMakeOrderBookPairResponse() {
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
	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, tickPrec, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    20,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)

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

func ExampleMakeOrderBookPairResponse_pool() {
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
	tickPrec := 3
	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, tickPrec, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)

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
	// |                              0.988400000000000000                      |
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

func ExampleMakeOrderBookPairResponse_userOrder() {
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
	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, tickPrec, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp.OrderBooks[0], basePrice)

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

func ExampleMakeOrderBookPairResponse_match() {
	ob := amm.NewOrderBook(
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.001"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(10000)),
	)

	ov := ob.MakeView()
	ov.Match()
	tickPrec := 3
	resp := types.MakeOrderBookPairResponse(1, ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), tickPrec, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	basePrice, found := types.OrderBookBasePrice(ov, tickPrec)
	if !found {
		panic("base price not found")
	}
	types.PrintOrderBookResponse(resp.OrderBooks[0], basePrice)

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

func ExampleMakeOrderBookPairResponse_zigzag() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.002"), sdk.NewInt(10000)),
		newOrder(amm.Buy, utils.ParseDec("1.001"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(50000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(100000)),
	)
	ov := ob.MakeView()
	ov.Match()

	basePrice, _ := types.OrderBookBasePrice(ov, 4)
	resp := types.MakeOrderBookPairResponse(1, ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    20,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], basePrice)

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

func ExampleMakeOrderBookPairResponse_edgecase1() {
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
	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, 3, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], basePrice)

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

func ExampleMakeOrderBookPairResponse_edgecase2() {
	ob := amm.NewOrderBook(
		newOrder(amm.Buy, utils.ParseDec("1.001"), sdk.NewInt(3000)),
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(10000)),
		newOrder(amm.Sell, utils.ParseDec("0.999"), sdk.NewInt(5000)),
		newOrder(amm.Sell, utils.ParseDec("0.998"), sdk.NewInt(3000)),
		newOrder(amm.Sell, utils.ParseDec("0.997"), sdk.NewInt(2000)),
	)

	ov := ob.MakeView()
	ov.Match()

	resp := types.MakeOrderBookPairResponse(1, ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |------------------------------------------------------------------------|
	// |                              1.000000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         1.000000000000000000 | 3000               |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookPairResponse_edgecase3() {
	ob := amm.NewOrderBook(
		newOrder(amm.Sell, utils.ParseDec("1.001"), sdk.NewInt(2000)),
		newOrder(amm.Sell, utils.ParseDec("1.000"), sdk.NewInt(2000)),
		newOrder(amm.Buy, utils.ParseDec("1.000"), sdk.NewInt(2000)),
		newOrder(amm.Buy, utils.ParseDec("0.999"), sdk.NewInt(2000)),
	)

	ov := ob.MakeView()
	ov.Match()

	resp := types.MakeOrderBookPairResponse(1, ov, utils.ParseDec("0.9"), utils.ParseDec("1.1"), 3, types.OrderBookConfig{
		PriceUnitPower: 0,
		MaxNumTicks:    10,
	})
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |               2000 |         1.001000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              1.000000000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 2000               |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookPairResponse_priceUnits1() {
	ob := amm.NewOrderBook()

	lastPrice := utils.ParseDec("0.9995")
	lowestPrice, highestPrice := types.PriceLimits(lastPrice, utils.ParseDec("0.1"), 4)
	pool := amm.NewBasicPool(sdk.NewInt(9995_000000), sdk.NewInt(10000_000000), sdk.Int{})
	ob.AddOrder(amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)

	ov := ob.MakeView()
	ov.Match()

	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, 4,
		types.OrderBookConfig{
			PriceUnitPower: 0,
			MaxNumTicks:    10,
		},
		types.OrderBookConfig{
			PriceUnitPower: 1,
			MaxNumTicks:    10,
		},
		types.OrderBookConfig{
			PriceUnitPower: 2,
			MaxNumTicks:    10,
		},
	)
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)
	types.PrintOrderBookResponse(resp.OrderBooks[1], resp.BasePrice)
	types.PrintOrderBookResponse(resp.OrderBooks[2], resp.BasePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |             300018 |         0.999870000000000000 |                    |
	// |              99970 |         0.999830000000000000 |                    |
	// |             300047 |         0.999790000000000000 |                    |
	// |              99989 |         0.999750000000000000 |                    |
	// |             300076 |         0.999710000000000000 |                    |
	// |             100009 |         0.999670000000000000 |                    |
	// |             300104 |         0.999630000000000000 |                    |
	// |             100029 |         0.999590000000000000 |                    |
	// |             300132 |         0.999550000000000000 |                    |
	// |             100049 |         0.999510000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999500000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999490000000000000 | 100051             |
	// |                    |         0.999450000000000000 | 300169             |
	// |                    |         0.999410000000000000 | 100071             |
	// |                    |         0.999370000000000000 | 300197             |
	// |                    |         0.999330000000000000 | 100091             |
	// |                    |         0.999290000000000000 | 300225             |
	// |                    |         0.999250000000000000 | 100111             |
	// |                    |         0.999210000000000000 | 300253             |
	// |                    |         0.999170000000000000 | 100132             |
	// |                    |         0.999130000000000000 | 300280             |
	// +------------------------------------------------------------------------+
	// +------------------------------------------------------------------------+
	// |             998495 |         1.000900000000000000 |                    |
	// |             998608 |         1.000700000000000000 |                    |
	// |             999095 |         1.000500000000000000 |                    |
	// |             999206 |         1.000300000000000000 |                    |
	// |             999695 |         1.000100000000000000 |                    |
	// |             499871 |         1.000000000000000000 |                    |
	// |             399988 |         0.999900000000000000 |                    |
	// |             700112 |         0.999800000000000000 |                    |
	// |             400113 |         0.999700000000000000 |                    |
	// |             500210 |         0.999600000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999500000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999400000000000000 | 500291             |
	// |                    |         0.999300000000000000 | 400288             |
	// |                    |         0.999200000000000000 | 700589             |
	// |                    |         0.999100000000000000 | 400412             |
	// |                    |         0.999000000000000000 | 500633             |
	// |                    |         0.998900000000000000 | 400528             |
	// |                    |         0.998800000000000000 | 700969             |
	// |                    |         0.998700000000000000 | 400653             |
	// |                    |         0.998600000000000000 | 500972             |
	// |                    |         0.998500000000000000 | 400768             |
	// +------------------------------------------------------------------------+
	// +------------------------------------------------------------------------+
	// |            4935785 |         1.009000000000000000 |                    |
	// |            4942943 |         1.008000000000000000 |                    |
	// |            4950502 |         1.007000000000000000 |                    |
	// |            4957699 |         1.006000000000000000 |                    |
	// |            4965293 |         1.005000000000000000 |                    |
	// |            4972529 |         1.004000000000000000 |                    |
	// |            4980159 |         1.003000000000000000 |                    |
	// |            4987432 |         1.002000000000000000 |                    |
	// |            4995099 |         1.001000000000000000 |                    |
	// |            2500294 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999500000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 2502213            |
	// |                    |         0.998000000000000000 | 5210466            |
	// |                    |         0.997000000000000000 | 5017777            |
	// |                    |         0.996000000000000000 | 5025346            |
	// |                    |         0.995000000000000000 | 4932203            |
	// |                    |         0.994000000000000000 | 4939515            |
	// |                    |         0.993000000000000000 | 4947464            |
	// |                    |         0.992000000000000000 | 4954234            |
	// |                    |         0.991000000000000000 | 5163351            |
	// |                    |         0.990000000000000000 | 5274693            |
	// +------------------------------------------------------------------------+
}

func ExampleMakeOrderBookPairResponse_priceUnits2() {
	ob := amm.NewOrderBook()

	lastPrice := utils.ParseDec("0.9999")
	lowestPrice, highestPrice := types.PriceLimits(lastPrice, utils.ParseDec("0.1"), 4)
	pool := amm.NewBasicPool(sdk.NewInt(9999_000000), sdk.NewInt(10000_000000), sdk.Int{})
	ob.AddOrder(amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)...)

	ov := ob.MakeView()
	ov.Match()

	resp := types.MakeOrderBookPairResponse(1, ov, lowestPrice, highestPrice, 4,
		types.OrderBookConfig{
			PriceUnitPower: 0,
			MaxNumTicks:    10,
		},
		types.OrderBookConfig{
			PriceUnitPower: 1,
			MaxNumTicks:    10,
		},
		types.OrderBookConfig{
			PriceUnitPower: 2,
			MaxNumTicks:    10,
		},
	)
	types.PrintOrderBookResponse(resp.OrderBooks[0], resp.BasePrice)
	types.PrintOrderBookResponse(resp.OrderBooks[1], resp.BasePrice)
	types.PrintOrderBookResponse(resp.OrderBooks[2], resp.BasePrice)

	// Output:
	// +------------------------------------------------------------------------+
	// |             997461 |         1.001700000000000000 |                    |
	// |             997650 |         1.001500000000000000 |                    |
	// |             998058 |         1.001300000000000000 |                    |
	// |             998247 |         1.001100000000000000 |                    |
	// |             998658 |         1.000900000000000000 |                    |
	// |             998845 |         1.000700000000000000 |                    |
	// |             999257 |         1.000500000000000000 |                    |
	// |             999444 |         1.000300000000000000 |                    |
	// |             999857 |         1.000100000000000000 |                    |
	// |             500010 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999900000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999800000000000000 | 500091             |
	// |                    |         0.999700000000000000 | 400128             |
	// |                    |         0.999600000000000000 | 700309             |
	// |                    |         0.999500000000000000 | 400252             |
	// |                    |         0.999400000000000000 | 500431             |
	// |                    |         0.999300000000000000 | 400369             |
	// |                    |         0.999200000000000000 | 700688             |
	// |                    |         0.999100000000000000 | 400492             |
	// |                    |         0.999000000000000000 | 500773             |
	// |                    |         0.998900000000000000 | 400608             |
	// +------------------------------------------------------------------------+
	// +------------------------------------------------------------------------+
	// |            4936732 |         1.009000000000000000 |                    |
	// |            4943972 |         1.008000000000000000 |                    |
	// |            4951453 |         1.007000000000000000 |                    |
	// |            4958730 |         1.006000000000000000 |                    |
	// |            4966248 |         1.005000000000000000 |                    |
	// |            4973563 |         1.004000000000000000 |                    |
	// |            4981117 |         1.003000000000000000 |                    |
	// |            4988468 |         1.002000000000000000 |                    |
	// |            4996061 |         1.001000000000000000 |                    |
	// |             500010 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999900000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.999000000000000000 | 4503533            |
	// |                    |         0.998000000000000000 | 5211459            |
	// |                    |         0.997000000000000000 | 5018778            |
	// |                    |         0.996000000000000000 | 4825162            |
	// |                    |         0.995000000000000000 | 5235152            |
	// |                    |         0.994000000000000000 | 5041557            |
	// |                    |         0.993000000000000000 | 4847036            |
	// |                    |         0.992000000000000000 | 5258875            |
	// |                    |         0.991000000000000000 | 4964041            |
	// |                    |         0.990000000000000000 | 5072107            |
	// +------------------------------------------------------------------------+
	// +------------------------------------------------------------------------+
	// |           39884892 |         1.090000000000000000 |                    |
	// |           35872783 |         1.080000000000000000 |                    |
	// |           49088294 |         1.070000000000000000 |                    |
	// |           52603184 |         1.060000000000000000 |                    |
	// |           43083979 |         1.050000000000000000 |                    |
	// |           43657842 |         1.040000000000000000 |                    |
	// |           50064733 |         1.030000000000000000 |                    |
	// |           50835962 |         1.020000000000000000 |                    |
	// |           49625631 |         1.010000000000000000 |                    |
	// |             500010 |         1.000000000000000000 |                    |
	// |------------------------------------------------------------------------|
	// |                              0.999900000000000000                      |
	// |------------------------------------------------------------------------|
	// |                    |         0.990000000000000000 | 49977700           |
	// |                    |         0.980000000000000000 | 50524375           |
	// |                    |         0.970000000000000000 | 54863157           |
	// |                    |         0.960000000000000000 | 52675477           |
	// |                    |         0.950000000000000000 | 57485995           |
	// |                    |         0.940000000000000000 | 40941734           |
	// |                    |         0.930000000000000000 | 65324860           |
	// |                    |         0.920000000000000000 | 58141770           |
	// |                    |         0.910000000000000000 | 43270158           |
	// |                    |         0.900000000000000000 | 72707068           |
	// +------------------------------------------------------------------------+
}
