package keeper_test

import (
	"fmt"
	"time"

	"github.com/tendermint/farming/x/farming"
)

func (suite *KeeperTestSuite) TestLastEpochTime() {
	_, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().False(found)

	t := mustParseRFC3339("2021-07-23T05:01:02Z")
	suite.keeper.SetLastEpochTime(suite.ctx, t)

	t2, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(t, t2)
}

func (suite *KeeperTestSuite) TestFirstEpoch() {
	// TODO: write test
}

func (suite *KeeperTestSuite) TestEpochDays() {
	for _, epochDays := range []uint32{1, 2, 3} {
		suite.Run(fmt.Sprintf("epoch days = %d", epochDays), func() {
			suite.SetupTest()

			params := suite.keeper.GetParams(suite.ctx)
			params.EpochDays = epochDays
			suite.keeper.SetParams(suite.ctx, params)

			t := mustParseRFC3339("2021-08-11T00:00:00Z")
			suite.ctx = suite.ctx.WithBlockTime(t)
			farming.EndBlocker(suite.ctx, suite.keeper)

			lastEpochTime, _ := suite.keeper.GetLastEpochTime(suite.ctx)

			for i := 0; i < 10000; i++ {
				t = t.Add(5 * time.Minute)
				suite.ctx = suite.ctx.WithBlockTime(t)
				farming.EndBlocker(suite.ctx, suite.keeper)

				t2, _ := suite.keeper.GetLastEpochTime(suite.ctx)
				if t2.After(lastEpochTime) {
					suite.Require().GreaterOrEqual(t2.Sub(lastEpochTime).Hours(), float64(epochDays*24))
					lastEpochTime = t2
				}
			}
		})
	}
}
