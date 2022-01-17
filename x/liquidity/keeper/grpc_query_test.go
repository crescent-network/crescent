package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)

	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCPools() {
	creator := s.addr(0)
	s.createPool(creator, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	s.createPool(creator, parseCoin("5000000denom1"), parseCoin("5000000denom3"), true)
	s.createPool(creator, parseCoin("3000000denom2"), parseCoin("3000000denom3"), true)
	s.createPool(creator, parseCoin("3000000denom3"), parseCoin("3000000denom4"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolsRequest
		expectErr bool
		postRun   func(*types.QueryPoolsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryPoolsRequest{},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 4)
			},
		},
		{
			"query all with query string XDenom",
			&types.QueryPoolsRequest{
				XDenom: "denom1",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 2)
			},
		},
		{
			"query all with query string YDenom",
			&types.QueryPoolsRequest{
				YDenom: "denom2",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
		{
			"query all with query string XDenom and YDenom",
			&types.QueryPoolsRequest{
				XDenom: "denom1",
				YDenom: "denom3",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pools(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPoolsByPair() {
	creator := s.addr(0)
	s.createPool(creator, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	s.createPool(creator, parseCoin("5000000denom1"), parseCoin("5000000denom3"), true)
	s.createPool(creator, parseCoin("3000000denom2"), parseCoin("3000000denom3"), true)
	s.createPool(creator, parseCoin("3000000denom3"), parseCoin("3000000denom4"), true)

	s.Require().Len(s.keeper.GetAllPairs(s.ctx), 4)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolsByPairRequest
		expectErr bool
		postRun   func(*types.QueryPoolsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request all",
			&types.QueryPoolsByPairRequest{},
			true,
			nil,
		},
		{
			"query all pool with pair id",
			&types.QueryPoolsByPairRequest{
				PairId: 1,
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PoolsByPair(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPool() {

}

func (s *KeeperTestSuite) TestGRPCPoolByReserveAcc() {

}

func (s *KeeperTestSuite) TestGRPCPoolByPoolCoinDenom() {

}

func (s *KeeperTestSuite) TestGRPCPairs() {

}

func (s *KeeperTestSuite) TestGRPCPair() {

}

func (s *KeeperTestSuite) TestGRPCDepositRequests() {

}

func (s *KeeperTestSuite) TestGRPCDepositRequest() {

}

func (s *KeeperTestSuite) TestGRPCWithdrawRequests() {

}

func (s *KeeperTestSuite) TestGRPCWithdrawRequest() {

}

func (s *KeeperTestSuite) TestGRPCSwapRequests() {

}
func (s *KeeperTestSuite) TestGRPCSwapRequest() {

}
