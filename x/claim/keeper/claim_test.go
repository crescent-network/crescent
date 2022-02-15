package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGetAllAirdrops() {
	// TODO: not implemented yet
}

func (s *KeeperTestSuite) TestGetAllClaimRecords() {
	// TODO: not implemented yet
}

func (s *KeeperTestSuite) TestDistributeByDivisor() {
	airdrop := s.createAirdrop(
		parseCoins("1000000000denom1"),
		s.ctx.BlockTime(),
		squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"),
		true,
	)

	record := s.createClaimRecord(
		airdrop.AirdropId,
		s.addr(0),
		parseCoins("100000000denom1"),
		parseCoins("100000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false},
		},
	)

	// Claim deposit action
	_, err := s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeDeposit,
	})
	s.Require().NoError(err)

	// Claim swap action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeSwap,
	})
	s.Require().NoError(err)

	// Claim farming action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeFarming,
	})
	s.Require().NoError(err)

	// Verify
	r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdrop.AirdropId, record.GetRecipient())
	s.Require().True(found)
	s.Require().True(r.ClaimableCoins.IsZero())
	s.Require().True(r.InitialClaimableCoins.IsEqual(sdk.NewCoins(s.getBalance(record.GetRecipient(), "denom1"))))
	for _, action := range r.Actions {
		s.Require().True(action.Claimed)
	}
}
