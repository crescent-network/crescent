package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.Ctx), resp.Params)
}

func (s *KeeperTestSuite) TestQueryAllPools() {
	s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10"))
	s.CreateMarketAndPool("uatom", "ucre", utils.ParseDec("2"))

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllPoolsRequest
		expectedErr string
		postRun     func(resp *types.QueryAllPoolsResponse)
	}{
		{
			"happy case",
			&types.QueryAllPoolsRequest{},
			"",
			func(resp *types.QueryAllPoolsResponse) {
				s.Require().Len(resp.Pools, 3)
				s.Require().Equal("0ucre", resp.Pools[0].Balance0.String())
				s.Require().Equal("0uusd", resp.Pools[0].Balance1.String())
			},
		},
		{
			"query by market id",
			&types.QueryAllPoolsRequest{
				MarketId: 2,
			},
			"",
			func(resp *types.QueryAllPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
				pool := resp.Pools[0]
				s.Require().EqualValues(2, pool.MarketId)
				s.Require().EqualValues(2, pool.Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllPools(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}
