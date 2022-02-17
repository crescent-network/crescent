package keeper_test

import (
	"github.com/cosmosquad-labs/squad/x/claim"
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestClaim() {
	// airdrop := s.createAirdrop(
	// 	1,
	// 	s.addr(0),
	// 	parseCoins("1000000000denom1"),
	// 	[]types.ConditionType{
	// 		types.ConditionTypeDeposit,
	// 		types.ConditionTypeSwap,
	// 		types.ConditionTypeFarming,
	// 	},
	// 	s.ctx.BlockTime(),
	// 	s.ctx.BlockTime().AddDate(0, 1, 0),
	// 	true,
	// )

	// record := s.createClaimRecord(
	// 	airdrop.Id,
	// 	s.addr(1),
	// 	parseCoins("666666667denom1"),
	// 	parseCoins("666666667denom1"),
	// 	[]bool{false, false, false},
	// )

	// // Claim deposit action
	// _, err := s.keeper.Claim(s.ctx, &types.MsgClaim{
	// 	AirdropId:     airdrop.Id,
	// 	Recipient:     record.Recipient,
	// 	ConditionType: types.ConditionTypeDeposit,
	// })
	// s.Require().NoError(err)

	// // Claim swap action
	// _, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
	// 	AirdropId:     airdrop.Id,
	// 	Recipient:     record.Recipient,
	// 	ConditionType: types.ConditionTypeSwap,
	// })
	// s.Require().NoError(err)

	// // Claim farming action
	// _, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
	// 	AirdropId:     airdrop.Id,
	// 	Recipient:     record.Recipient,
	// 	ConditionType: types.ConditionTypeFarming,
	// })
	// s.Require().NoError(err)

	// Claim already claimed action
	// _, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
	// 	AirdropId:     airdrop.Id,
	// 	Recipient:     record.Recipient,
	// 	ConditionType: types.ConditionTypeDeposit,
	// })
	// s.Require().ErrorIs(err, types.ErrAlreadyClaimed)

	// // Verify
	// r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdrop.Id, record.GetRecipient())
	// s.Require().True(found)
	// s.Require().True(r.ClaimableCoins.IsZero())
	// s.Require().True(coinsEq(r.InitialClaimableCoins, sdk.NewCoins(s.getBalance(record.GetRecipient(), "denom1"))))
	// s.Require().True(r.ClaimedConditions[0])
	// s.Require().True(r.ClaimedConditions[1])
	// s.Require().True(r.ClaimedConditions[2])
}

func (s *KeeperTestSuite) TestClaimExecuteSameAction() {
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		parseCoins("1000000000denom1"),
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
		s.addr(0),
		parseCoins("100000000denom1"),
		parseCoins("100000000denom1"),
		[]bool{false, false, false},
	)

	// Claim deposit action
	_, err := s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:     airdrop.Id,
		Recipient:     record.Recipient,
		ConditionType: types.ConditionTypeDeposit,
	})
	s.Require().NoError(err)

	// Claim the already completed deposit action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:     airdrop.Id,
		Recipient:     record.Recipient,
		ConditionType: types.ConditionTypeDeposit,
	})
	s.Require().ErrorIs(err, types.ErrAlreadyClaimed)
}
func (s *KeeperTestSuite) TestTerminateAidropClaimAll() {
	// airdrop := s.createAirdrop(
	// 	1,
	// 	s.addr(0),
	// 	parseCoins("1000000000denom1"),
	// 	[]types.ConditionType{
	// 		types.ConditionTypeDeposit,
	// 		types.ConditionTypeSwap,
	// 		types.ConditionTypeFarming,
	// 	},
	// 	s.ctx.BlockTime(),
	// 	s.ctx.BlockTime().AddDate(0, 1, 0),
	// 	true,
	// )

	// record := s.createClaimRecord(
	// 	airdrop.Id,
	// 	s.addr(0),
	// 	parseCoins("100000000denom1"),
	// 	parseCoins("100000000denom1"),
	// 	[]bool{false, false, false},
	// )

	// // Claim deposit action
	// _, err := s.keeper.Claim(s.ctx, types.NewMsgClaim(airdrop.Id, record.GetRecipient(), types.ConditionTypeDeposit))
	// s.Require().NoError(err)

	// // Claim swap action
	// _, err = s.keeper.Claim(s.ctx, types.NewMsgClaim(airdrop.Id, record.GetRecipient(), types.ConditionTypeSwap))
	// s.Require().NoError(err)

	// // Claim farming action
	// _, err = s.keeper.Claim(s.ctx, types.NewMsgClaim(airdrop.Id, record.GetRecipient(), types.ConditionTypeFarming))
	// s.Require().NoError(err)

	// err = s.keeper.TerminateAirdrop(s.ctx, airdrop)
	// s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestTerminatAirdropClaimPartial() {
	airdrop := s.createAirdrop(
		1,
		s.addr(0),
		parseCoins("1000000000denom1"),
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
		s.addr(0),
		parseCoins("100000000denom1"),
		parseCoins("100000000denom1"),
		[]bool{false, false, false},
	)

	// Claim deposit action
	_, err := s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:     airdrop.Id,
		Recipient:     record.Recipient,
		ConditionType: types.ConditionTypeDeposit,
	})
	s.Require().NoError(err)

	// Claim swap action
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:     airdrop.Id,
		Recipient:     record.Recipient,
		ConditionType: types.ConditionTypeSwap,
	})
	s.Require().NoError(err)

	// Terminate the airdrop
	s.ctx = s.ctx.WithBlockTime(airdrop.EndTime.AddDate(0, 0, 1))
	claim.EndBlocker(s.ctx, s.keeper)

	// Claim farming action must fail
	_, err = s.keeper.Claim(s.ctx, &types.MsgClaim{
		AirdropId:     airdrop.Id,
		Recipient:     record.Recipient,
		ConditionType: types.ConditionTypeFarming,
	})
	s.Require().ErrorIs(err, types.ErrTerminatedAirdrop)

	sourceBalanaces := s.getAllBalances(airdrop.GetSourceAddress())
	s.Require().True(sourceBalanaces.IsZero())
}
