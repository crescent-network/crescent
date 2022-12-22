package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/marketmaker/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesisState()

	suite.keeper.InitGenesis(suite.ctx, genState)
	got := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().Equal(genState, *got)
}

func (suite *KeeperTestSuite) TestImportExportGenesisEmpty() {
	k, ctx := suite.keeper, suite.ctx
	genState := k.ExportGenesis(ctx)

	var genState2 types.GenesisState
	bz := suite.app.AppCodec().MustMarshalJSON(genState)
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)

	genState3 := k.ExportGenesis(ctx)
	suite.Require().Equal(*genState, genState2, *genState3)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[5].String()
	k.SetParams(ctx, params)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2, 3, 4, 5, 6})
	suite.NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{2, 3, 4, 5, 6, 7})
	suite.NoError(err)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description",
		[]types.MarketMakerHandle{
			{Address: mmAddr.String(), PairId: 1},
			{Address: mmAddr2.String(), PairId: 3}},
		nil, nil, nil)
	suite.handleProposal(proposal)

	// distribute incentive
	incentiveAmount := sdk.NewInt(500000000)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
			{
				Address: mmAddr2.String(),
				PairId:  3,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	mms := k.GetAllMarketMakers(ctx)
	suite.Require().Len(mms, 12)

	incentives := k.GetAllIncentives(ctx)
	suite.Require().Len(incentives, 2)

	var genState *types.GenesisState
	suite.Require().NotPanics(func() {
		genState = suite.keeper.ExportGenesis(suite.ctx)
	})

	err = types.ValidateGenesis(*genState)
	suite.Require().NoError(err)

	suite.Require().NotPanics(func() {
		suite.keeper.InitGenesis(suite.ctx, *genState)
	})
	suite.Require().Equal(genState, suite.keeper.ExportGenesis(suite.ctx))

	mms = suite.keeper.GetAllMarketMakers(ctx)
	suite.Require().Len(mms, 12)

	incentives = suite.keeper.GetAllIncentives(ctx)
	suite.Require().Len(incentives, 2)
}
