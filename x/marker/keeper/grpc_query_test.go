package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/marker/types"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCLastBlockTime() {
	resp, err := s.querier.LastBlockTime(sdk.WrapSDKContext(s.ctx), &types.QueryLastBlockTimeRequest{})
	s.Require().NoError(err)
	s.Require().Nil(resp.LastBlockTime)
	
	s.nextBlock()

	resp, err = s.querier.LastBlockTime(sdk.WrapSDKContext(s.ctx), &types.QueryLastBlockTimeRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp.LastBlockTime)
}
