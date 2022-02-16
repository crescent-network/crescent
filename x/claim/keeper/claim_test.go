package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/claim"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestClaim() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		squadtypes.MustParseRFC3339("2022-02-01T00:00:00Z"),
		squadtypes.MustParseRFC3339("2022-06-01T00:00:00Z"),
		true,
	)

	record := s.createClaimRecord(airdrop.AirdropId, s.addr(0), parseCoins("100000000denom1"), parseCoins("100000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false}},
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

	// Claim farming action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeDeposit,
	})
	s.Require().ErrorIs(err, types.ErrAlreadyClaimedAll)

	// Verify
	r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdrop.AirdropId, record.GetRecipient())
	s.Require().True(found)
	s.Require().True(r.ClaimableCoins.IsZero())
	s.Require().True(coinsEq(r.InitialClaimableCoins, sdk.NewCoins(s.getBalance(record.GetRecipient(), "denom1"))))
	s.Require().True(r.Actions[0].Claimed)
	s.Require().True(r.Actions[1].Claimed)
	s.Require().True(r.Actions[2].Claimed)
}

func (s *KeeperTestSuite) TestClaimExecuteSameAction() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		squadtypes.MustParseRFC3339("2022-02-01T00:00:00Z"),
		squadtypes.MustParseRFC3339("2022-06-01T00:00:00Z"),
		true,
	)

	record := s.createClaimRecord(airdrop.AirdropId, s.addr(0), parseCoins("100000000denom1"), parseCoins("100000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false}},
	)

	// Claim deposit action
	_, err := s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeDeposit,
	})
	s.Require().NoError(err)

	// Claim the already completed deposit action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeDeposit,
	})
	s.Require().ErrorIs(err, types.ErrAlreadyClaimed)
}

func (s *KeeperTestSuite) TestClaimAirdropTerminated() {
	airdrop := s.createAirdrop(
		1,
		parseCoins("1000000000denom1"),
		squadtypes.MustParseRFC3339("2022-02-01T00:00:00Z"),
		squadtypes.MustParseRFC3339("2022-06-01T00:00:00Z"),
		true,
	)

	record := s.createClaimRecord(airdrop.AirdropId, s.addr(0), parseCoins("100000000denom1"), parseCoins("100000000denom1"),
		[]types.Action{
			{ActionType: types.ActionTypeDeposit, Claimed: false},
			{ActionType: types.ActionTypeSwap, Claimed: false},
			{ActionType: types.ActionTypeFarming, Claimed: false}},
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

	// Terminate the airdrop
	s.ctx = s.ctx.WithBlockTime(airdrop.EndTime.AddDate(0, 0, 1))
	claim.EndBlocker(s.ctx, s.keeper)

	// Claim farming action must fail
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:  airdrop.AirdropId,
		Recipient:  record.Recipient,
		ActionType: types.ActionTypeFarming,
	})
	s.Require().ErrorIs(err, types.ErrTerminatedAirdrop)

	t1 := s.getAllBalances(airdrop.GetSourceAddress())
	t2 := s.getAllBalances(airdrop.GetTerminationAddress())
	fmt.Println("t1: ", t1)
	fmt.Println("t2: ", t2)
}
