package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCClaimRecord() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		s.ctx.BlockTime(),
		squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"),
		true,
	)

	s.createClaimRecord(
		airdrop.AirdropId,
		s.addr(0),
		parseCoins("100000000denom1"),
		parseCoins("100000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: true},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false},
		},
	)

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
				AirdropId: airdrop.AirdropId,
				Address:   s.addr(0).String(),
			},
			false,
			func(resp *types.QueryClaimRecordResponse) {
				s.Require().Equal(s.addr(0).String(), resp.ClaimRecord.Recipient)
				s.Require().True(coinsEq(parseCoins("100000000denom1"), resp.ClaimRecord.InitialClaimableCoins))
				s.Require().True(resp.ClaimRecord.Actions[0].Claimed)
				s.Require().False(resp.ClaimRecord.Actions[1].Claimed)
				s.Require().False(resp.ClaimRecord.Actions[2].Claimed)
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
