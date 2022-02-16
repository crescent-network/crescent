package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCAirdrops() {
	s.createAirdrop(1, parseCoins("1000000000denom1"), s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(2, parseCoins("1000000000denom1"), s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(3, parseCoins("1000000000denom1"), s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryAirdropsRequest
		expectErr bool
		postRun   func(*types.QueryAirdropsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all airdrops",
			&types.QueryAirdropsRequest{},
			false,
			func(resp *types.QueryAirdropsResponse) {
				s.Require().Len(resp.Airdrops, 3)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Airdrops(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCAirdrop() {
	airdrop := s.createAirdrop(1, parseCoins("1000000000denom1"), s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryAirdropRequest
		expectErr bool
		postRun   func(*types.QueryAirdropResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"airdrop not found",
			&types.QueryAirdropRequest{
				AirdropId: 5,
			},
			true,
			nil,
		},
		{
			"airdrop not found",
			&types.QueryAirdropRequest{
				AirdropId: 1,
			},
			false,
			func(resp *types.QueryAirdropResponse) {
				s.Require().Equal(airdrop.SourceAddress, resp.Airdrop.SourceAddress)
				s.Require().Equal(airdrop.TerminationAddress, resp.Airdrop.TerminationAddress)
				s.Require().Equal(airdrop.StartTime, resp.Airdrop.StartTime)
				s.Require().Equal(airdrop.EndTime, resp.Airdrop.EndTime)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Airdrop(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCClaimRecord() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

	s.createClaimRecord(
		airdrop.Id,
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
				Recipient: s.addr(5).String(),
			},
			true,
			nil,
		},
		{
			"query by address",
			&types.QueryClaimRecordRequest{
				AirdropId: airdrop.Id,
				Recipient: s.addr(0).String(),
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
