package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestQueryAllOrders() {
	market1 := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("4.99"), sdk.NewInt(1000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr2, false, utils.ParseDec("5.01"), sdk.NewInt(2000000), time.Hour)
	market2 := s.CreateMarket(utils.TestAddress(0), "uatom", "ucre", true)
	s.PlaceLimitOrder(market2.Id, ordererAddr1, false, utils.ParseDec("3"), sdk.NewInt(1500000), time.Hour)
	s.PlaceLimitOrder(market2.Id, ordererAddr2, true, utils.ParseDec("2.9"), sdk.NewInt(2500000), time.Hour)

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllOrdersRequest
		expectedErr string
		postRun     func(resp *types.QueryAllOrdersResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"happy case",
			&types.QueryAllOrdersRequest{},
			"",
			func(resp *types.QueryAllOrdersResponse) {
				s.Require().Len(resp.Orders, 4)
			},
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

func (s *KeeperTestSuite) TestQueryBestSwapExactAmountInRoutes() {
	creatorAddr := utils.TestAddress(1)
	s.FundAccount(creatorAddr, utils.ParseCoins("100000_000000ucre,100000_000000uatom,100000_000000uusd"))

	market1 := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	market2 := s.CreateMarket(utils.TestAddress(0), "uatom", "ucre", true)
	market3 := s.CreateMarket(utils.TestAddress(0), "uatom", "uusd", true)

	pool1 := s.CreatePool(creatorAddr, market1.Id, utils.ParseDec("9.7"), true)
	s.AddLiquidity(creatorAddr, creatorAddr, pool1.Id, utils.ParseDec("9.5"), utils.ParseDec("10"),
		utils.ParseCoins("1000_000000ucre,10000_000000uusd"))
	pool2 := s.CreatePool(creatorAddr, market2.Id, utils.ParseDec("1.04"), true)
	s.AddLiquidity(creatorAddr, creatorAddr, pool2.Id, utils.ParseDec("1"), utils.ParseDec("1.2"),
		utils.ParseCoins("1000_000000uatom,1000_000000ucre"))
	pool3 := s.CreatePool(creatorAddr, market3.Id, utils.ParseDec("10.3"), true)
	s.AddLiquidity(creatorAddr, creatorAddr, pool3.Id, utils.ParseDec("9.7"), utils.ParseDec("11"),
		utils.ParseCoins("1000_000000uatom,10000_000000uusd"))

	querier := keeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.BestSwapExactAmountInRoutes(sdk.WrapSDKContext(s.Ctx), &types.QueryBestSwapExactAmountInRoutesRequest{
		Input:       "100000000ucre",
		OutputDenom: "uusd",
	})
	s.Require().NoError(err)

	s.Require().EqualValues([]uint64{2, 3}, resp.Routes)
	s.Require().Equal("972699534uusd", resp.Output.String())
	s.Require().Len(resp.Results, 2)
	s.Require().EqualValues(2, resp.Results[0].MarketId)
	s.Require().Equal("100000000ucre", resp.Results[0].Input.String())
	s.Require().Equal("95135825uatom", resp.Results[0].Output.String())
	s.Require().Equal("142919uatom", resp.Results[0].Fee.String())
	s.Require().EqualValues(3, resp.Results[1].MarketId)
	s.Require().Equal("95135825uatom", resp.Results[1].Input.String())
	s.Require().Equal("972699534uusd", resp.Results[1].Output.String())
	s.Require().Equal("1461242uusd", resp.Results[1].Fee.String())
}
