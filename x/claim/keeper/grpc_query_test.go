package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCAirdrops() {
	conditions := []types.ConditionType{
		types.ConditionTypeDeposit,
		types.ConditionTypeSwap,
		types.ConditionTypeFarming,
	}

	s.createAirdrop(1, s.addr(1), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(2, s.addr(2), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)
	s.createAirdrop(3, s.addr(3), squad.ParseCoins("1000000000denom1"), conditions,
		s.ctx.BlockTime(), s.ctx.BlockTime().AddDate(0, 1, 0), true)

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
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		squad.ParseCoins("1000000000denom1"),
		[]types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeFarming,
		},
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

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
			"query with airdrop id",
			&types.QueryAirdropRequest{
				AirdropId: airdrop.Id,
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
		s.addr(0),
		squad.ParseCoins("1000000000denom1"),
		[]types.ConditionType{
			types.ConditionTypeDeposit,
			types.ConditionTypeSwap,
			types.ConditionTypeFarming,
		},
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 1, 0),
		true,
	)

	record := s.createClaimRecord(
		airdrop.Id,
		s.addr(1),
		squad.ParseCoins("90000000denom1"),
		squad.ParseCoins("600000000denom1"),
		[]bool{true, false, false},
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
			"query with not eligible recipient address",
			&types.QueryClaimRecordRequest{
				Recipient: s.addr(5).String(),
			},
			true,
			nil,
		},
		{
			"query by airdrop id and recipient address",
			&types.QueryClaimRecordRequest{
				AirdropId: airdrop.Id,
				Recipient: record.Recipient,
			},
			false,
			func(resp *types.QueryClaimRecordResponse) {
				s.Require().Equal(record.Recipient, resp.ClaimRecord.Recipient)
				s.Require().True(coinsEq(squad.ParseCoins("90000000denom1"), record.InitialClaimableCoins))
				s.Require().True(resp.ClaimRecord.ClaimedConditions[0])
				s.Require().False(resp.ClaimRecord.ClaimedConditions[1])
				s.Require().False(resp.ClaimRecord.ClaimedConditions[2])
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
