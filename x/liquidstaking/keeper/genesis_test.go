package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
)

//func (suite *KeeperTestSuite) TestInitGenesis() {
//	suite.SetupTest()
//	params := suite.keeper.GetParams(suite.ctx)
//	//params.WhitelistedValidators =
//	suite.keeper.SetParams(suite.ctx, params)
//
//	emptyGenState := suite.keeper.ExportGenesis(suite.ctx)
//	suite.Require().NotPanics(func() {
//		suite.keeper.InitGenesis(suite.ctx, *emptyGenState)
//	})
//	suite.Require().Equal(emptyGenState, suite.keeper.ExportGenesis(suite.ctx))
//	suite.Require().Nil(emptyGenState.LiquidValidators)
//
//	var genState *types.GenesisState
//	suite.Require().NotPanics(func() {
//		genState = suite.keeper.ExportGenesis(suite.ctx)
//	})
//	err := types.ValidateGenesis(*genState)
//	suite.Require().NoError(err)
//
//	suite.Require().NotNil(genState.LiquidValidators)
//	suite.Require().NotPanics(func() {
//		suite.keeper.InitGenesis(suite.ctx, *genState)
//	})
//	suite.Require().Equal(genState, suite.keeper.ExportGenesis(suite.ctx))
//}
