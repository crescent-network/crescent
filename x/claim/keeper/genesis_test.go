package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/claim/types"

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
				Id:            1,
				SourceAddress: s.addr(0).String(),
				Conditions: []types.ConditionType{
					types.ConditionTypeDeposit,
					types.ConditionTypeSwap,
					types.ConditionTypeLiquidStake,
					types.ConditionTypeVote,
				},
				StartTime: s.ctx.BlockTime(),
				EndTime:   s.ctx.BlockTime().AddDate(0, 1, 0),
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				AirdropId:             1,
				Recipient:             s.addr(1).String(),
				InitialClaimableCoins: utils.ParseCoins("50000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("50000000000denom1"),
				ClaimedConditions:     []types.ConditionType{},
			},
		},
	}

	s.fundAddr(sampleGenState.Airdrops[0].GetSourceAddress(), utils.ParseCoins("100000000000denom1"))
	s.Require().NotPanics(func() {
		s.keeper.InitGenesis(s.ctx, sampleGenState)
	})

	_, found := s.keeper.GetAirdrop(s.ctx, 1)
	s.Require().True(found)

	_, found = s.keeper.GetClaimRecordByRecipient(s.ctx, 1, sampleGenState.ClaimRecords[0].GetRecipient())
	s.Require().True(found)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	sampleGenState := types.GenesisState{
		Airdrops: []types.Airdrop{
			{
				Id:            1,
				SourceAddress: s.addr(0).String(),
				Conditions: []types.ConditionType{
					types.ConditionTypeDeposit,
					types.ConditionTypeSwap,
					types.ConditionTypeLiquidStake,
					types.ConditionTypeVote,
				},
				StartTime: s.ctx.BlockTime(),
				EndTime:   s.ctx.BlockTime().AddDate(0, 1, 0),
			},
			{
				Id:            2,
				SourceAddress: s.addr(1).String(),
				Conditions: []types.ConditionType{
					types.ConditionTypeDeposit,
					types.ConditionTypeSwap,
					types.ConditionTypeLiquidStake,
					types.ConditionTypeVote,
				},
				StartTime: s.ctx.BlockTime().AddDate(0, 5, 0),
				EndTime:   s.ctx.BlockTime().AddDate(0, 10, 0),
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				AirdropId:             1,
				Recipient:             s.addr(2).String(),
				InitialClaimableCoins: utils.ParseCoins("50000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("50000000000denom1"),
				ClaimedConditions:     []types.ConditionType{},
			},
			{
				AirdropId:             1,
				Recipient:             s.addr(3).String(),
				InitialClaimableCoins: utils.ParseCoins("50000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("50000000000denom1"),
				ClaimedConditions: []types.ConditionType{
					types.ConditionTypeLiquidStake,
				},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(3).String(),
				InitialClaimableCoins: utils.ParseCoins("100000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("100000000000denom1"),
				ClaimedConditions:     []types.ConditionType{},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(4).String(),
				InitialClaimableCoins: utils.ParseCoins("50000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("50000000000denom1"),
				ClaimedConditions: []types.ConditionType{
					types.ConditionTypeDeposit,
				},
			},
			{
				AirdropId:             2,
				Recipient:             s.addr(5).String(),
				InitialClaimableCoins: utils.ParseCoins("50000000000denom1"),
				ClaimableCoins:        utils.ParseCoins("50000000000denom1"),
				ClaimedConditions: []types.ConditionType{
					types.ConditionTypeDeposit,
					types.ConditionTypeSwap,
					types.ConditionTypeLiquidStake,
				},
			},
		},
	}
	s.fundAddr(sampleGenState.Airdrops[0].GetSourceAddress(), utils.ParseCoins("100000000000denom1"))
	s.fundAddr(sampleGenState.Airdrops[1].GetSourceAddress(), utils.ParseCoins("200000000000denom1"))

	// Initialize genesis state
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
		s.app = chain.Setup(false)
		s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
		s.keeper = s.app.ClaimKeeper
		s.keeper.InitGenesis(s.ctx, *genState)

		s.Require().Len(s.keeper.GetAllAirdrops(s.ctx), 2)
		s.Require().Len(s.keeper.GetAllClaimRecordsByAirdropId(s.ctx, 1), 2)
		s.Require().Len(s.keeper.GetAllClaimRecordsByAirdropId(s.ctx, 2), 3)
	})
	s.Require().Equal(genState, s.keeper.ExportGenesis(s.ctx))
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	k, ctx := s.keeper, s.ctx
	genState := k.ExportGenesis(ctx)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)

	genState3 := k.ExportGenesis(ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
