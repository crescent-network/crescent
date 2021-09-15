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
	// The first epoch may run very quickly depending on when
	// the farming module is deployed,
	// meaning that (block time) - (last epoch time) may be smaller
	// than epoch_days parameter on the first epoch.

	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(uint32(1), params.EpochDays)

	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-08-11T23:59:59Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	lastEpochTime, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().True(found)

	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-08-12T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	t, _ := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().True(t.After(lastEpochTime)) // Indicating that the epoch advanced.
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

func (suite *KeeperTestSuite) TestDelayedBlockTime() {
	// Entire network can be down for several days,
	// and the epoch should be advanced after the downtime.
	suite.keeper.SetLastEpochTime(suite.ctx, mustParseRFC3339("2021-09-23T00:00:05Z"))

	t := mustParseRFC3339("2021-10-03T00:00:04Z")
	suite.ctx = suite.ctx.WithBlockTime(t)
	farming.EndBlocker(suite.ctx, suite.keeper)

	lastEpochTime, found := suite.keeper.GetLastEpochTime(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(t, lastEpochTime)
}
