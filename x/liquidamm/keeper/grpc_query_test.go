package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.Ctx), resp.Params)
}

func (s *KeeperTestSuite) TestQueryPublicPositions() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryPublicPositionsRequest
		expectedErr string
		postRun     func(resp *types.QueryPublicPositionsResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"happy case",
			&types.QueryPublicPositionsRequest{},
			"",
			func(resp *types.QueryPublicPositionsResponse) {
				s.Require().Len(resp.PublicPositions, 3)
				publicPosition := resp.PublicPositions[0]
				s.Require().EqualValues(1, publicPosition.Id)
				s.Require().EqualValues(2, publicPosition.LastRewardsAuctionId)
				s.Require().Equal(
					s.App.BankKeeper.GetSupply(s.Ctx, types.ShareDenom(publicPosition.Id)),
					publicPosition.TotalShare)
				s.Require().Equal(sdk.NewInt(43138144377), publicPosition.Liquidity)
			},
		},
		{
			"pool not found",
			&types.QueryPublicPositionsRequest{PoolId: 3},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"by pool id",
			&types.QueryPublicPositionsRequest{PoolId: 2},
			"",
			func(resp *types.QueryPublicPositionsResponse) {
				s.Require().Len(resp.PublicPositions, 2)
				s.Require().Equal(uint64(2), resp.PublicPositions[0].PoolId)
				s.Require().Equal(uint64(2), resp.PublicPositions[1].PoolId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PublicPositions(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryPublicPosition() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryPublicPositionRequest
		expectedErr string
		postRun     func(resp *types.QueryPublicPositionResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid public position id",
			&types.QueryPublicPositionRequest{
				PublicPositionId: 0,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"public position not found",
			&types.QueryPublicPositionRequest{
				PublicPositionId: 4,
			},
			"rpc error: code = NotFound desc = public position not found",
			nil,
		},
		{
			"happy case",
			&types.QueryPublicPositionRequest{
				PublicPositionId: 1,
			},
			"",
			func(resp *types.QueryPublicPositionResponse) {
				publicPosition := resp.PublicPosition
				s.Require().EqualValues(1, publicPosition.Id)
				s.Require().EqualValues(2, publicPosition.LastRewardsAuctionId)
				s.Require().Equal(
					s.App.BankKeeper.GetSupply(s.Ctx, types.ShareDenom(publicPosition.Id)),
					publicPosition.TotalShare)
				s.Require().Equal(sdk.NewInt(43138144377), publicPosition.Liquidity)
				s.Require().Equal(uint64(2), publicPosition.PositionId)
			},
		},
		{
			"no amm position",
			&types.QueryPublicPositionRequest{
				PublicPositionId: 2,
			},
			"",
			func(resp *types.QueryPublicPositionResponse) {
				publicPosition := resp.PublicPosition
				s.Require().Equal(uint64(2), publicPosition.Id)
				s.Require().Equal(uint64(0), publicPosition.PositionId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PublicPosition(sdk.WrapSDKContext(s.Ctx), tc.req)
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
			"invalid public position id",
			&types.QueryRewardsAuctionsRequest{
				PublicPositionId: 0,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"public position not found",
			&types.QueryRewardsAuctionsRequest{
				PublicPositionId: 4,
			},
			"rpc error: code = NotFound desc = public position not found",
			nil,
		},
		{
			"invalid auction status",
			&types.QueryRewardsAuctionsRequest{
				PublicPositionId: 1,
				Status:           "blah",
			},
			"rpc error: code = InvalidArgument desc = invalid auction status blah",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsAuctionsRequest{
				PublicPositionId: 1,
			},
			"",
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 2)
				auction := resp.RewardsAuctions[0]
				s.Require().EqualValues(1, auction.Id)
				s.Require().Equal(types.AuctionStatusFinished, auction.Status)
				s.Require().Equal(utils.ParseCoin("1307216496sb1"), auction.WinningBid.Share)
				auction = resp.RewardsAuctions[1]
				s.Require().EqualValues(2, auction.Id)
				s.Require().Equal(types.AuctionStatusStarted, auction.Status)
				s.Require().Equal(utils.ParseCoin("814642164sb1"), auction.WinningBid.Share)
			},
		},
		{
			"query by status",
			&types.QueryRewardsAuctionsRequest{
				PublicPositionId: 1,
				Status:           types.AuctionStatusStarted.String(),
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
				PublicPositionId: 1,
				Status:           types.AuctionStatusFinished.String(),
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
			"invalid public position id",
			&types.QueryRewardsAuctionRequest{
				PublicPositionId: 0,
				AuctionId:        1,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"public position not found",
			&types.QueryRewardsAuctionRequest{
				PublicPositionId: 4,
				AuctionId:        1,
			},
			"rpc error: code = NotFound desc = public position not found",
			nil,
		},
		{
			"invalid auction id",
			&types.QueryRewardsAuctionRequest{
				PublicPositionId: 1,
				AuctionId:        0,
			},
			"rpc error: code = InvalidArgument desc = auction id must not be 0",
			nil,
		},
		{
			"auction not found",
			&types.QueryRewardsAuctionRequest{
				PublicPositionId: 1,
				AuctionId:        3,
			},
			"rpc error: code = NotFound desc = auction not found",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsAuctionRequest{
				PublicPositionId: 1,
				AuctionId:        1,
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
				PublicPositionId: 1,
				AuctionId:        2,
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
			"happy case",
			&types.QueryBidsRequest{},
			"",
			func(resp *types.QueryBidsResponse) {
				s.Require().Len(resp.Bids, 1)
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

func (s *KeeperTestSuite) TestQueryWinningBid() {
	s.SetupSampleScenario()
	for _, tc := range []struct {
		name        string
		req         *types.QueryWinningBidRequest
		expectedErr string
		postRun     func(*types.QueryWinningBidResponse)
	}{
		{
			"empty request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"invalid public position id",
			&types.QueryWinningBidRequest{
				PublicPositionId: 0,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"public position not found",
			&types.QueryWinningBidRequest{
				PublicPositionId: 4,
			},
			"rpc error: code = NotFound desc = public position not found",
			nil,
		},
		{
			"rewards auction not started yet",
			&types.QueryWinningBidRequest{
				PublicPositionId: 3,
			},
			"rpc error: code = NotFound desc = rewards auction not started yet",
			nil,
		},
		{
			"happy case",
			&types.QueryWinningBidRequest{
				PublicPositionId: 1,
			},
			"",
			func(resp *types.QueryWinningBidResponse) {
				// All bids have been deleted since the auction is finished.
				s.Require().NotNil(resp.WinningBid)
				s.AssertEqual(utils.ParseCoin("814642164sb1"), resp.WinningBid.Share)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.WinningBid(sdk.WrapSDKContext(s.Ctx), tc.req)
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
			"invalid public position id",
			&types.QueryRewardsRequest{
				PublicPositionId: 0,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"happy case",
			&types.QueryRewardsRequest{
				PublicPositionId: 1,
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

func (s *KeeperTestSuite) TestQueryExchangeRate() {
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
			"public position not found",
			&types.QueryExchangeRateRequest{
				PublicPositionId: 10,
			},
			"rpc error: code = InvalidArgument desc = public position id must not be 0",
			nil,
		},
		{
			"zero share supply",
			&types.QueryExchangeRateRequest{
				PublicPositionId: 2,
			},
			"",
			func(resp *types.QueryExchangeRateResponse) {
				s.AssertEqual(sdk.ZeroDec(), resp.MintRate)
				s.AssertEqual(sdk.ZeroDec(), resp.BurnRate)
			},
		},
		{
			"happy case",
			&types.QueryExchangeRateRequest{
				PublicPositionId: 1,
			},
			"",
			func(resp *types.QueryExchangeRateResponse) {
				s.Require().Equal("0.934782608695148231", resp.MintRate.String())
				s.Require().Equal("1.036177474410059339", resp.BurnRate.String())
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
