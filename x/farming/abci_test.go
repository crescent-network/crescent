package farming_test

import (
	"time"

	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *ModuleTestSuite) TestEndBlockerEpochDaysTest() {
	epochDaysTest := func(formerEpochDays, targetNextEpochDays uint32) {
		suite.SetupTest()

		params := suite.keeper.GetParams(suite.ctx)
		params.NextEpochDays = formerEpochDays
		suite.keeper.SetParams(suite.ctx, params)
		suite.keeper.SetCurrentEpochDays(suite.ctx, formerEpochDays)

		t := types.ParseTime("2021-08-01T00:00:00Z")
		suite.ctx = suite.ctx.WithBlockTime(t)
		farming.EndBlocker(suite.ctx, suite.keeper)

		lastEpochTime, _ := suite.keeper.GetLastEpochTime(suite.ctx)

		for i := 1; i < 200; i++ {
			t = t.Add(1 * time.Hour)
			suite.ctx = suite.ctx.WithBlockTime(t)
			farming.EndBlocker(suite.ctx, suite.keeper)

			if i == 10 { // 10 hours passed
				params := suite.keeper.GetParams(suite.ctx)
				params.NextEpochDays = targetNextEpochDays
				suite.keeper.SetParams(suite.ctx, params)
			}

			currentEpochDays := suite.keeper.GetCurrentEpochDays(suite.ctx)
			t2, _ := suite.keeper.GetLastEpochTime(suite.ctx)

			if uint32(i) == formerEpochDays*24 {
				suite.Require().True(t2.After(lastEpochTime))
				suite.Require().Equal(t2.Sub(lastEpochTime).Hours(), float64(formerEpochDays*24))
				suite.Require().Equal(targetNextEpochDays, currentEpochDays)
			}

			if uint32(i) == formerEpochDays*24+targetNextEpochDays*24 {
				suite.Require().Equal(t2.Sub(lastEpochTime).Hours(), float64(currentEpochDays*24))
				suite.Require().Equal(targetNextEpochDays, currentEpochDays)
			}

			lastEpochTime = t2
		}
	}

	// increasing case
	epochDaysTest(1, 7)

	// decreasing case
	epochDaysTest(7, 1)

	// stay case
	epochDaysTest(1, 1)
}
