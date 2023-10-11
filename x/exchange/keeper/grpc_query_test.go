package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// SetupSampleScenario creates markets and orders for query tests.
func (s *KeeperTestSuite) SetupSampleScenario() {
	s.T().Helper()

	mmAddr := s.FundedAccount(100, enoughCoins)
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "uusd")
	s.MakeLastPrice(market1.Id, mmAddr, utils.ParseDec("5"))
	s.MakeLastPrice(market2.Id, mmAddr, utils.ParseDec("10"))

	aliceAddr := s.FundedAccount(1, enoughCoins)
	bobAddr := s.FundedAccount(2, enoughCoins)

	s.PlaceLimitOrder(market1.Id, aliceAddr, true, utils.ParseDec("4.9999"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, aliceAddr, true, utils.ParseDec("4.9998"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, aliceAddr, false, utils.ParseDec("5.0001"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, aliceAddr, false, utils.ParseDec("5.0002"), sdk.NewInt(10_000000), time.Hour)

	s.PlaceLimitOrder(market1.Id, bobAddr, true, utils.ParseDec("4.99"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, bobAddr, true, utils.ParseDec("4.98"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, bobAddr, false, utils.ParseDec("5.01"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, bobAddr, false, utils.ParseDec("5.02"), sdk.NewInt(100_000000), time.Hour)

	s.PlaceLimitOrder(market2.Id, aliceAddr, true, utils.ParseDec("9.9999"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, aliceAddr, true, utils.ParseDec("9.9998"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, aliceAddr, false, utils.ParseDec("10.001"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, aliceAddr, false, utils.ParseDec("10.002"), sdk.NewInt(10_000000), time.Hour)

	s.PlaceLimitOrder(market2.Id, bobAddr, true, utils.ParseDec("9.99"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, bobAddr, true, utils.ParseDec("9.98"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, bobAddr, false, utils.ParseDec("10.01"), sdk.NewInt(100_000000), time.Hour)
	s.PlaceLimitOrder(market2.Id, bobAddr, false, utils.ParseDec("10.02"), sdk.NewInt(100_000000), time.Hour)
}

func (s *KeeperTestSuite) TestQueryParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.Ctx), resp.Params)
}

func (s *KeeperTestSuite) TestQueryAllMarkets() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllMarketsRequest
		expectedErr string
		postRun     func(resp *types.QueryAllMarketsResponse)
	}{
		{
			"happy case",
			&types.QueryAllMarketsRequest{},
			"",
			func(resp *types.QueryAllMarketsResponse) {
				s.Require().Len(resp.Markets, 2)
				s.Require().EqualValues(1, resp.Markets[0].Id)
				s.Require().EqualValues(2, resp.Markets[1].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllMarkets(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryMarket() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryMarketRequest
		expectedErr string
		postRun     func(resp *types.QueryMarketResponse)
	}{
		{
			"happy case",
			&types.QueryMarketRequest{
				MarketId: 2,
			},
			"",
			func(resp *types.QueryMarketResponse) {
				s.Require().EqualValues(2, resp.Market.Id)
				s.AssertEqual(utils.ParseDec("0.0015"), resp.Market.Fees.MakerFeeRate)
				s.AssertEqual(utils.ParseDec("0.003"), resp.Market.Fees.TakerFeeRate)
				s.AssertEqual(utils.ParseDec("0.5"), resp.Market.Fees.OrderSourceFeeRatio)
			},
		},
		{
			"market not found",
			&types.QueryMarketRequest{
				MarketId: 3,
			},
			"rpc error: code = NotFound desc = market not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Market(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryAllOrders() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllOrdersRequest
		expectedErr string
		postRun     func(resp *types.QueryAllOrdersResponse)
	}{
		{
			"happy case",
			&types.QueryAllOrdersRequest{},
			"",
			func(resp *types.QueryAllOrdersResponse) {
				s.Require().Len(resp.Orders, 16)
				for i, order := range resp.Orders {
					// First 4 orders were for making last price
					s.Require().EqualValues(i+5, order.Id)
				}
			},
		},
		{
			"with orderer",
			&types.QueryAllOrdersRequest{
				Orderer: utils.TestAddress(1).String(), // Alice
			},
			"",
			func(resp *types.QueryAllOrdersResponse) {
				s.Require().Len(resp.Orders, 8)
				for _, order := range resp.Orders {
					s.Require().Equal(utils.TestAddress(1).String(), order.Orderer)
				}
			},
		},
		{
			"with market",
			&types.QueryAllOrdersRequest{
				MarketId: 2,
			},
			"",
			func(resp *types.QueryAllOrdersResponse) {
				s.Require().Len(resp.Orders, 8)
				for _, order := range resp.Orders {
					s.Require().EqualValues(2, order.MarketId)
				}
			},
		},
		{
			"with orderer and market",
			&types.QueryAllOrdersRequest{
				Orderer:  utils.TestAddress(2).String(), // Bob
				MarketId: 2,
			},
			"",
			func(resp *types.QueryAllOrdersResponse) {
				s.Require().Len(resp.Orders, 4)
				for _, order := range resp.Orders {
					s.Require().Equal(utils.TestAddress(2).String(), order.Orderer)
					s.Require().EqualValues(2, order.MarketId)
				}
			},
		},
		{
			"invalid orderer",
			&types.QueryAllOrdersRequest{
				Orderer: "invalid",
			},
			"rpc error: code = InvalidArgument desc = invalid orderer: decoding bech32 failed: invalid bech32 string length 7",
			nil,
		},
		{
			"market not found",
			&types.QueryAllOrdersRequest{
				MarketId: 3,
			},
			"rpc error: code = NotFound desc = market not found",
			nil,
		},
		{
			"market not found 2",
			&types.QueryAllOrdersRequest{
				Orderer:  utils.TestAddress(2).String(), // Bob
				MarketId: 3,
			},
			"rpc error: code = NotFound desc = market not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllOrders(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryOrder() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryOrderRequest
		expectedErr string
		postRun     func(resp *types.QueryOrderResponse)
	}{
		{
			"happy case",
			&types.QueryOrderRequest{
				OrderId: 6,
			},
			"",
			func(resp *types.QueryOrderResponse) {
				s.Require().EqualValues(6, resp.Order.Id)
			},
		},
		{
			"order not found",
			&types.QueryOrderRequest{
				OrderId: 100,
			},
			"rpc error: code = NotFound desc = order not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Order(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryBestSwapExactAmountInRoutes() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryBestSwapExactAmountInRoutesRequest
		expectedErr string
		postRun     func(resp *types.QueryBestSwapExactAmountInRoutesResponse)
	}{
		{
			"happy case",
			&types.QueryBestSwapExactAmountInRoutesRequest{
				Input:       "5000000uatom",
				OutputDenom: "ucre",
			},
			"",
			func(resp *types.QueryBestSwapExactAmountInRoutesResponse) {
				s.Require().Equal([]uint64{2, 1}, resp.Routes)
				s.AssertEqual(utils.ParseCoin("9969722ucre"), resp.Output)
				s.Assert().EqualValues(2, resp.Results[0].MarketId)
				s.AssertEqual(utils.ParseCoin("5000000uatom"), resp.Results[0].Input)
				s.AssertEqual(utils.ParseCoin("49924500uusd"), resp.Results[0].Output)
				s.AssertEqual(utils.ParseCoin("75000uusd"), resp.Results[0].Fee)
				s.Assert().EqualValues(1, resp.Results[1].MarketId)
				s.AssertEqual(utils.ParseCoin("49924500uusd"), resp.Results[1].Input)
				s.AssertEqual(utils.ParseCoin("9969722ucre"), resp.Results[1].Output)
				s.AssertEqual(utils.ParseCoin("14978ucre"), resp.Results[1].Fee)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryOrderBook() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryOrderBookRequest
		expectedErr string
		postRun     func(resp *types.QueryOrderBookResponse)
	}{
		{
			"happy case",
			&types.QueryOrderBookRequest{
				MarketId: 1,
			},
			"",
			func(resp *types.QueryOrderBookResponse) {
				s.Require().Len(resp.OrderBooks, 3)
				s.Require().Equal(utils.ParseDec("0.0001"), resp.OrderBooks[0].PriceInterval)
				s.Require().Len(resp.OrderBooks[0].Sells, 4)
				s.Require().Equal(utils.ParseDec("5.0001"), resp.OrderBooks[0].Sells[0].P)
				s.Require().Equal(sdk.NewInt(10000000), resp.OrderBooks[0].Sells[0].Q)
				s.Require().Equal(utils.ParseDec("5.0002"), resp.OrderBooks[0].Sells[1].P)
				s.Require().Equal(sdk.NewInt(10000000), resp.OrderBooks[0].Sells[1].Q)
				s.Require().Equal(utils.ParseDec("5.01"), resp.OrderBooks[0].Sells[2].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[0].Sells[2].Q)
				s.Require().Equal(utils.ParseDec("5.02"), resp.OrderBooks[0].Sells[3].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[0].Sells[3].Q)
				s.Require().Len(resp.OrderBooks[0].Buys, 4)
				s.Require().Equal(utils.ParseDec("4.9999"), resp.OrderBooks[0].Buys[0].P)
				s.Require().Equal(sdk.NewInt(10000000), resp.OrderBooks[0].Buys[0].Q)
				s.Require().Equal(utils.ParseDec("4.9998"), resp.OrderBooks[0].Buys[1].P)
				s.Require().Equal(sdk.NewInt(10000000), resp.OrderBooks[0].Buys[1].Q)
				s.Require().Equal(utils.ParseDec("4.99"), resp.OrderBooks[0].Buys[2].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[0].Buys[2].Q)
				s.Require().Equal(utils.ParseDec("4.98"), resp.OrderBooks[0].Buys[3].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[0].Buys[3].Q)

				s.Require().Equal(utils.ParseDec("0.001"), resp.OrderBooks[1].PriceInterval)
				s.Require().Len(resp.OrderBooks[1].Sells, 3)
				s.Require().Equal(utils.ParseDec("5.001"), resp.OrderBooks[1].Sells[0].P)
				s.Require().Equal(sdk.NewInt(20000000), resp.OrderBooks[1].Sells[0].Q)
				s.Require().Equal(utils.ParseDec("5.01"), resp.OrderBooks[1].Sells[1].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[1].Sells[1].Q)
				s.Require().Equal(utils.ParseDec("5.02"), resp.OrderBooks[1].Sells[2].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[1].Sells[2].Q)
				s.Require().Len(resp.OrderBooks[1].Buys, 3)
				s.Require().Equal(utils.ParseDec("4.999"), resp.OrderBooks[1].Buys[0].P)
				s.Require().Equal(sdk.NewInt(20000000), resp.OrderBooks[1].Buys[0].Q)
				s.Require().Equal(utils.ParseDec("4.99"), resp.OrderBooks[1].Buys[1].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[1].Buys[1].Q)
				s.Require().Equal(utils.ParseDec("4.98"), resp.OrderBooks[1].Buys[2].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[1].Buys[2].Q)

				s.Require().Equal(utils.ParseDec("0.01"), resp.OrderBooks[2].PriceInterval)
				s.Require().Len(resp.OrderBooks[2].Sells, 2)
				s.Require().Equal(utils.ParseDec("5.01"), resp.OrderBooks[2].Sells[0].P)
				s.Require().Equal(sdk.NewInt(120000000), resp.OrderBooks[2].Sells[0].Q)
				s.Require().Equal(utils.ParseDec("5.02"), resp.OrderBooks[2].Sells[1].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[2].Sells[1].Q)
				s.Require().Len(resp.OrderBooks[2].Buys, 2)
				s.Require().Equal(utils.ParseDec("4.99"), resp.OrderBooks[2].Buys[0].P)
				s.Require().Equal(sdk.NewInt(120000000), resp.OrderBooks[2].Buys[0].Q)
				s.Require().Equal(utils.ParseDec("4.98"), resp.OrderBooks[2].Buys[1].P)
				s.Require().Equal(sdk.NewInt(100000000), resp.OrderBooks[2].Buys[1].Q)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.OrderBook(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFindBestSwapExactAmountInRoutes() {
	s.CreateMarket("ucre", "uusd")
	s.CreateMarket("uatom", "ucre")
	s.CreateMarket("stake", "uatom")
	s.CreateMarket("uatom", "stake")

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(1, lpAddr, utils.ParseDec("5"))
	s.MakeLastPrice(2, lpAddr, utils.ParseDec("2"))
	s.MakeLastPrice(3, lpAddr, utils.ParseDec("0.33333"))
	s.MakeLastPrice(4, lpAddr, utils.ParseDec("3"))

	s.createLiquidity(1, lpAddr, utils.ParseDec("5"), sdk.NewInt(1000_000000))
	s.createLiquidity(2, lpAddr, utils.ParseDec("2"), sdk.NewInt(1000_000000))
	s.createLiquidity(3, lpAddr, utils.ParseDec("0.33333"), sdk.NewInt(3000_000000))
	s.createLiquidity(4, lpAddr, utils.ParseDec("3"), sdk.NewInt(100_000000))

	routes := s.keeper.FindAllRoutes(s.Ctx, "uusd", "stake", 3)
	s.Require().Equal([][]uint64{{1, 2, 3}, {1, 2, 4}}, routes)

	resp, err := s.querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
		Input:       "2000000000uusd", // 2000_000000uusd
		OutputDenom: "stake",
	})
	s.Require().NoError(err)
	s.AssertEqual(utils.ParseCoin("594299197stake"), resp.Output)

	ordererAddr := s.FundedAccount(2, utils.ParseCoins("2000_000000uusd"))
	output, _ := s.SwapExactAmountIn(ordererAddr, resp.Routes, utils.ParseCoin("2000_000000uusd"), resp.Output, false)
	s.AssertEqual(resp.Output, output)
}
