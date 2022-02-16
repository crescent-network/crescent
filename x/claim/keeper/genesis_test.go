package keeper_test

import (
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, *genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, got)
}

func (s *KeeperTestSuite) TestInitExportGenesis() {
	sampleGenState := types.GenesisState{
		Airdrops: []types.Airdrop{
			{
				AirdropId:          1,
				SourceAddress:      s.addr(0).String(),
				SourceCoins:        parseCoins("100000000000denom1"),
				TerminationAddress: s.addr(5).String(),
				StartTime:          s.ctx.BlockTime(),
				EndTime:            s.ctx.BlockTime().AddDate(0, 1, 0),
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				AirdropId:             1,
				Recipient:             s.addr(1).String(),
				InitialClaimableCoins: parseCoins("50000000000denom1"),
				ClaimableCoins:        parseCoins("50000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: false},
					{ActionType: types.ActionTypeSwap, Claimed: false},
					{ActionType: types.ActionTypeFarming, Claimed: false},
				},
			},
		},
	}

	// Source account balances are empty; therefore it panics
	s.Require().Panics(func() {
		s.keeper.InitGenesis(s.ctx, sampleGenState)
	})

	s.fundAddr(sampleGenState.Airdrops[0].GetSourceAddress(), sampleGenState.Airdrops[0].SourceCoins)
	s.keeper.InitGenesis(s.ctx, sampleGenState)

	_, found := s.keeper.GetAirdrop(s.ctx, 1)
	s.Require().True(found)

	_, found = s.keeper.GetClaimRecordByRecipient(s.ctx, 1, sampleGenState.ClaimRecords[0].GetRecipient())
	s.Require().True(found)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	sampleGenState := types.GenesisState{
		Airdrops: []types.Airdrop{
			{
				AirdropId:          1,
				SourceAddress:      s.addr(0).String(),
				SourceCoins:        parseCoins("100000000000denom1"),
				TerminationAddress: s.addr(6).String(),
				StartTime:          s.ctx.BlockTime(),
				EndTime:            s.ctx.BlockTime().AddDate(0, 1, 0),
			},
			{
				AirdropId:          2,
				SourceAddress:      s.addr(1).String(),
				SourceCoins:        parseCoins("200000000000denom1"),
				TerminationAddress: s.addr(6).String(),
				StartTime:          s.ctx.BlockTime().AddDate(0, 5, 0),
				EndTime:            s.ctx.BlockTime().AddDate(0, 7, 0),
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				AirdropId:             1,
				Recipient:             s.addr(2).String(),
				InitialClaimableCoins: parseCoins("50000000000denom1"),
				ClaimableCoins:        parseCoins("50000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: true},
					{ActionType: types.ActionTypeSwap, Claimed: false},
					{ActionType: types.ActionTypeFarming, Claimed: false},
				},
			},
			{
				AirdropId:             1,
				Recipient:             s.addr(3).String(),
				InitialClaimableCoins: parseCoins("50000000000denom1"),
				ClaimableCoins:        parseCoins("50000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: false},
					{ActionType: types.ActionTypeSwap, Claimed: true},
					{ActionType: types.ActionTypeFarming, Claimed: true},
				},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(3).String(),
				InitialClaimableCoins: parseCoins("100000000000denom1"),
				ClaimableCoins:        parseCoins("100000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: false},
					{ActionType: types.ActionTypeSwap, Claimed: false},
					{ActionType: types.ActionTypeFarming, Claimed: false},
				},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(4).String(),
				InitialClaimableCoins: parseCoins("50000000000denom1"),
				ClaimableCoins:        parseCoins("50000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: false},
					{ActionType: types.ActionTypeSwap, Claimed: false},
					{ActionType: types.ActionTypeFarming, Claimed: false},
				},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(5).String(),
				InitialClaimableCoins: parseCoins("50000000000denom1"),
				ClaimableCoins:        parseCoins("50000000000denom1"),
				Actions: []types.Action{
					{ActionType: types.ActionTypeDeposit, Claimed: false},
					{ActionType: types.ActionTypeSwap, Claimed: false},
					{ActionType: types.ActionTypeFarming, Claimed: false},
				},
			},
		},
	}

	// Initialize genesis state
	s.fundAddr(sampleGenState.Airdrops[0].GetSourceAddress(), sampleGenState.Airdrops[0].SourceCoins)
	s.fundAddr(sampleGenState.Airdrops[1].GetSourceAddress(), sampleGenState.Airdrops[1].SourceCoins)
	s.Require().NotPanics(func() {
		s.keeper.InitGenesis(s.ctx, sampleGenState)
	})

	// Export genesis state
	var genState *types.GenesisState
	s.Require().NotPanics(func() {
		genState = s.keeper.ExportGenesis(s.ctx)
	})
	s.Require().Len(genState.Airdrops, 2)
	s.Require().Len(genState.ClaimRecords, 5)

	// Reinitialize exported genesis
	s.Require().NotPanics(func() {
		s.keeper.InitGenesis(s.ctx, *genState)
	})
	s.Require().Equal(genState, s.keeper.ExportGenesis(s.ctx))
}
