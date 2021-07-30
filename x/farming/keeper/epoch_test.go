package keeper_test

func (suite *KeeperTestSuite) TestLastEpochTime() {
	_, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().False(found)

	t := mustParseRFC3339("2021-07-23T05:01:02Z")
	suite.keeper.SetLastEpochTime(suite.ctx, t)

	t2, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(t, t2)
}
