package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.Ctx), resp.Params)
}

func (s *KeeperTestSuite) TestQueryLiquidFarms() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryLiquidFarmsRequest
		expectedErr string
		postRun     func(resp *types.QueryLiquidFarmsResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmsRequest{},
			"",
			func(resp *types.QueryLiquidFarmsResponse) {
				s.Require().Len(resp.LiquidFarms, 1)
				liquidFarm := resp.LiquidFarms[0]
				s.Require().EqualValues(1, liquidFarm.Id)
				s.Require().EqualValues(2, liquidFarm.LastRewardsAuctionId)
				s.Require().Equal(
					s.App.BankKeeper.GetSupply(s.Ctx, types.ShareDenom(liquidFarm.Id)),
					liquidFarm.TotalShare)
				s.Require().Equal(sdk.NewInt(43138144377), liquidFarm.Liquidity)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarms(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryLiquidFarm() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryLiquidFarmRequest
		expectedErr string
		postRun     func(resp *types.QueryLiquidFarmResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryLiquidFarmRequest{
				LiquidFarmId: 0,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"liquid farm not found",
			&types.QueryLiquidFarmRequest{
				LiquidFarmId: 2,
			},
			"rpc error: code = NotFound desc = liquid farm not found",
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmRequest{
				LiquidFarmId: 1,
			},
			"",
			func(resp *types.QueryLiquidFarmResponse) {
				liquidFarm := resp.LiquidFarm
				s.Require().EqualValues(1, liquidFarm.Id)
				s.Require().EqualValues(2, liquidFarm.LastRewardsAuctionId)
				s.Require().Equal(
					s.App.BankKeeper.GetSupply(s.Ctx, types.ShareDenom(liquidFarm.Id)),
					liquidFarm.TotalShare)
				s.Require().Equal(sdk.NewInt(43138144377), liquidFarm.Liquidity)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarm(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryRewardsAuctions() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryRewardsAuctionsRequest
		expectedErr string
		postRun     func(*types.QueryRewardsAuctionsResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 0,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"liquid farm not found",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 2,
			},
			"rpc error: code = NotFound desc = liquid farm not found",
			nil,
		},
		{
			"invalid auction status",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 1,
				Status:       "blah",
			},
			"rpc error: code = InvalidArgument desc = invalid auction status blah",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 1,
			},
			"",
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 2)
				auction := resp.RewardsAuctions[0]
				s.Require().EqualValues(1, auction.Id)
				s.Require().Equal(types.AuctionStatusFinished, auction.Status)
				s.Require().Equal(utils.ParseCoin("1307216496lfshare1"), auction.WinningBid.Share)
				auction = resp.RewardsAuctions[1]
				s.Require().EqualValues(2, auction.Id)
				s.Require().Equal(types.AuctionStatusStarted, auction.Status)
				s.Require().Equal(utils.ParseCoin("814642164lfshare1"), auction.WinningBid.Share)
			},
		},
		{
			"query by status",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 1,
				Status:       types.AuctionStatusStarted.String(),
			},
			"",
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 1)
				s.Require().EqualValues(2, resp.RewardsAuctions[0].Id)
			},
		},
		{
			"query by status 2",
			&types.QueryRewardsAuctionsRequest{
				LiquidFarmId: 1,
				Status:       types.AuctionStatusFinished.String(),
			},
			"",
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 1)
				s.Require().EqualValues(1, resp.RewardsAuctions[0].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuctions(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryRewardsAuction() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryRewardsAuctionRequest
		expectedErr string
		postRun     func(*types.QueryRewardsAuctionResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 0,
				AuctionId:    1,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"liquid farm not found",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 2,
				AuctionId:    1,
			},
			"rpc error: code = NotFound desc = liquid farm not found",
			nil,
		},
		{
			"invalid auction id",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 1,
				AuctionId:    0,
			},
			"rpc error: code = InvalidArgument desc = auction id must not be 0",
			nil,
		},
		{
			"auction not found",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 1,
				AuctionId:    3,
			},
			"rpc error: code = NotFound desc = auction not found",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 1,
				AuctionId:    1,
			},
			"",
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().EqualValues(1, resp.RewardsAuction.Id)
				s.Require().Equal(types.AuctionStatusFinished, resp.RewardsAuction.Status)
			},
		},
		{
			"happy case 2",
			&types.QueryRewardsAuctionRequest{
				LiquidFarmId: 1,
				AuctionId:    2,
			},
			"",
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().EqualValues(2, resp.RewardsAuction.Id)
				s.Require().Equal(types.AuctionStatusStarted, resp.RewardsAuction.Status)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuction(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryBids() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryBidsRequest
		expectedErr string
		postRun     func(*types.QueryBidsResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryBidsRequest{
				LiquidFarmId: 0,
				AuctionId:    1,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"liquid farm not found",
			&types.QueryBidsRequest{
				LiquidFarmId: 2,
				AuctionId:    1,
			},
			"rpc error: code = NotFound desc = liquid farm not found",
			nil,
		},
		{
			"invalid auction id",
			&types.QueryBidsRequest{
				LiquidFarmId: 1,
				AuctionId:    0,
			},
			"rpc error: code = InvalidArgument desc = auction id must not be 0",
			nil,
		},
		{
			"auction not found",
			&types.QueryBidsRequest{
				LiquidFarmId: 1,
				AuctionId:    3,
			},
			"rpc error: code = NotFound desc = auction not found",
			nil,
		},
		{
			"happy case",
			&types.QueryBidsRequest{
				LiquidFarmId: 1,
				AuctionId:    1,
			},
			"",
			func(resp *types.QueryBidsResponse) {
				// All bids have been deleted since the auction is finished.
				s.Require().Empty(resp.Bids)
			},
		},
		{
			"happy case 2",
			&types.QueryBidsRequest{
				LiquidFarmId: 1,
				AuctionId:    2,
			},
			"",
			func(resp *types.QueryBidsResponse) {
				s.Require().Len(resp.Bids, 2)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Bids(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryRewards() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryRewardsRequest
		expectedErr string
		postRun     func(*types.QueryRewardsResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryRewardsRequest{
				LiquidFarmId: 0,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsRequest{
				LiquidFarmId: 1,
			},
			"",
			func(resp *types.QueryRewardsResponse) {
				s.Require().True(resp.Rewards.IsAllPositive())
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Rewards(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCExchangeRate() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryExchangeRateRequest
		expectedErr string
		postRun     func(*types.QueryExchangeRateResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid liquid farm id",
			&types.QueryExchangeRateRequest{
				LiquidFarmId: 0,
			},
			"rpc error: code = InvalidArgument desc = liquid farm id must not be 0",
			nil,
		},
		{
			"liquid farm not found",
			&types.QueryExchangeRateRequest{
				LiquidFarmId: 2,
			},
			"rpc error: code = InvalidArgument desc = liquid farm not found",
			nil,
		},
		{
			"happy case",
			&types.QueryExchangeRateRequest{
				LiquidFarmId: 1,
			},
			"",
			func(resp *types.QueryExchangeRateResponse) {
				s.Require().Equal(utils.ParseDec("0.934782608695148232"), resp.ExchangeRate.MintRate)
				s.Require().Equal(utils.ParseDec("1.037350246659894737"), resp.ExchangeRate.BurnRate)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.ExchangeRate(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
