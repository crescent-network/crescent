package keeper_test

import (
	"github.com/crescent-network/crescent/v5/x/budget/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	suite.SetupTest()
	params := suite.keeper.GetParams(suite.ctx)
	params.Budgets = suite.budgets[:4]
	suite.keeper.SetParams(suite.ctx, params)

	emptyGenState := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NotPanics(func() {
		suite.keeper.InitGenesis(suite.ctx, *emptyGenState)
	})
	suite.Require().Equal(emptyGenState, suite.keeper.ExportGenesis(suite.ctx))
	suite.Require().EqualValues(emptyGenState.BudgetRecords, []types.BudgetRecord{})

	err := suite.keeper.CollectBudgets(suite.ctx)
	suite.Require().NoError(err)

	var genState *types.GenesisState
	suite.Require().NotPanics(func() {
		genState = suite.keeper.ExportGenesis(suite.ctx)
	})
	err = types.ValidateGenesis(*genState)
	suite.Require().NoError(err)

	suite.Require().NotNil(genState.BudgetRecords)
	suite.Require().NotPanics(func() {
		suite.keeper.InitGenesis(suite.ctx, *genState)
	})
	suite.Require().Equal(genState, suite.keeper.ExportGenesis(suite.ctx))
}

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesisState()

	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	k, ctx := s.keeper, s.ctx
	k.SetParams(ctx, types.DefaultParams())
	genState := k.ExportGenesis(ctx)

	bz := s.app.AppCodec().MustMarshalJSON(genState)

	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)
	genState3 := k.ExportGenesis(ctx)

	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
