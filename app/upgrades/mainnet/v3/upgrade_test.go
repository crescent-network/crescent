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
				suite.Require().EqualValues(marketmakerparams.DepositAmount, sdk.NewCoins(sdk.NewCoin("ucre", sdk.NewInt(1000000000))))
				suite.Require().EqualValues(marketmakerparams.IncentiveBudgetAddress, "cre1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sq3qhhde")
				suite.Require().EqualValues(marketmakerparams.Common, marketmakertypes.Common{
					MinOpenRatio:      sdk.MustNewDecFromStr("0.5"),
					MinOpenDepthRatio: sdk.MustNewDecFromStr("0.1"),
					MaxDowntime:       uint32(20),
					MaxTotalDowntime:  uint32(100),
					MinHours:          uint32(16),
					MinDays:           uint32(22),
				})

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
