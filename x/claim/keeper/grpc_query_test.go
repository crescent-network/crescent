package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCClaimRecord() {
	cr := types.ClaimRecord{
		Address:               s.addr(0).String(),
		InitialClaimableCoins: parseCoins("10000000denom1"),
		DepositActionClaimed:  true,
		SwapActionClaimed:     false,
		FarmingActionClaimed:  false,
	}
	s.keeper.SetClaimRecord(s.ctx, cr)

	for _, tc := range []struct {
		name      string
		req       *types.QueryClaimRecordRequest
		expectErr bool
		postRun   func(*types.QueryClaimRecordResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query with not eligible address",
			&types.QueryClaimRecordRequest{
				Address: s.addr(5).String(),
			},
			true,
			nil,
		},
		{
			"query by address",
			&types.QueryClaimRecordRequest{
				Address: s.addr(0).String(),
			},
			false,
			func(resp *types.QueryClaimRecordResponse) {
				s.Require().Equal(s.addr(0).String(), resp.ClaimRecord.Address)
				s.Require().True(coinsEq(parseCoins("10000000denom1"), resp.ClaimRecord.InitialClaimableCoins))
				s.Require().True(resp.ClaimRecord.DepositActionClaimed)
				s.Require().False(resp.ClaimRecord.SwapActionClaimed)
				s.Require().False(resp.ClaimRecord.FarmingActionClaimed)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.ClaimRecord(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
