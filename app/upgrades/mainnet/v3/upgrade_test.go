package v3_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/crescent-network/crescent/v3/app"
	"github.com/crescent-network/crescent/v3/app/upgrades/mainnet/v3"
	"github.com/crescent-network/crescent/v3/cmd/crescentd/cmd"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	marketmakertypes "github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

type UpgradeTestSuite struct {
	suite.Suite
	ctx sdk.Context
	app *app.App
}

func (suite *UpgradeTestSuite) SetupTest() {
	cmd.GetConfig()
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const testUpgradeHeight = 10

func (suite *UpgradeTestSuite) TestUpgradeV2() {
	testCases := []struct {
		title   string
		before  func()
		after   func()
		expPass bool
	}{
		{
			"v3 upgrade liquidity, market maker params",
			func() {
			},
			func() {
				liquidityparams := suite.app.LiquidityKeeper.GetParams(suite.ctx)
				marketmakerparams := suite.app.MarketMakerKeeper.GetParams(suite.ctx)
				suite.Require().EqualValues(liquidityparams.MaxNumMarketMakingOrderTicks, liquiditytypes.DefaultMaxNumMarketMakingOrderTicks)
				suite.Require().EqualValues(marketmakerparams.IncentivePairs, []marketmakertypes.IncentivePair(nil))
				marketmakerparams.IncentivePairs = []marketmakertypes.IncentivePair{}
				suite.Require().EqualValues(marketmakerparams, marketmakertypes.DefaultParams())
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.title, func() {
			suite.SetupTest()

			tc.before()

			suite.ctx = suite.ctx.WithBlockHeight(testUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: v3.UpgradeName, Height: testUpgradeHeight}
			err := suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
			suite.Require().NoError(err)
			_, exists := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
			suite.Require().True(exists)

			suite.ctx = suite.ctx.WithBlockHeight(testUpgradeHeight)
			suite.Require().NotPanics(func() {
				suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})
			})

			tc.after()
		})
	}
}
