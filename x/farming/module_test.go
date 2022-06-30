package farming_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

const (
	denom1 = "denom1" // staking coin denom 1
	denom2 = "denom2" // staking coin denom 2
	denom3 = "denom3" // epoch amount for a fixed amount plan
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
		sdk.NewInt64Coin(denom1, 1_000_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000_000),
		sdk.NewInt64Coin(denom3, 1_000_000_000))
)

type ModuleTestSuite struct {
	suite.Suite

	app                 *chain.App
	ctx                 sdk.Context
	keeper              keeper.Keeper
	querier             keeper.Querier
	addrs               []sdk.AccAddress
	sampleFixedAmtPlans []types.PlanI
	sampleRatioPlans    []types.PlanI
	samplePlans         []types.PlanI
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (suite *ModuleTestSuite) SetupTest() {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	keeper.EnableRatioPlan = true

	suite.app = app
	suite.ctx = ctx
	suite.keeper = suite.app.FarmingKeeper
	suite.querier = keeper.Querier{Keeper: suite.keeper}
	suite.addrs = chain.AddTestAddrs(suite.app, suite.ctx, 6, sdk.ZeroInt())
	for _, addr := range suite.addrs {
		err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
		suite.Require().NoError(err)
	}
	suite.sampleFixedAmtPlans = []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"testPlan1",
				types.PlanTypePrivate,
				suite.addrs[4].String(),
				suite.addrs[4].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
				),
				types.ParseTime("2021-08-02T00:00:00Z"),
				types.ParseTime("2021-09-02T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000)),
		),
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				2,
				"testPlan2",
				types.PlanTypePublic,
				suite.addrs[5].String(),
				suite.addrs[5].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.OneDec()), // 100%
				),
				types.ParseTime("2021-08-04T00:00:00Z"),
				types.ParseTime("2021-08-12T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 2_000_000)),
		),
	}
	suite.sampleRatioPlans = []types.PlanI{
		types.NewRatioPlan(
			types.NewBasePlan(
				3,
				"testPlan3",
				types.PlanTypePrivate,
				suite.addrs[4].String(),
				suite.addrs[4].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(5, 1)), // 50%
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(5, 1)), // 50%
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-09T00:00:00Z"),
			),
			sdk.NewDecWithPrec(4, 2), // 4%
		),
		types.NewRatioPlan(
			types.NewBasePlan(
				4,
				"testPlan4",
				types.PlanTypePublic,
				suite.addrs[5].String(),
				suite.addrs[5].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom2, sdk.OneDec()), // 100%
				),
				types.ParseTime("2021-08-03T00:00:00Z"),
				types.ParseTime("2021-08-07T00:00:00Z"),
			),
			sdk.NewDecWithPrec(3, 2), // 3%
		),
	}
	suite.samplePlans = append(suite.sampleFixedAmtPlans, suite.sampleRatioPlans...)
}

// Stake is a convenient method to test Keeper.Stake.
func (suite *ModuleTestSuite) Stake(farmerAcc sdk.AccAddress, amt sdk.Coins) {
	err := suite.keeper.Stake(suite.ctx, farmerAcc, amt)
	suite.Require().NoError(err)
}

// Rewards is a convenient method to test Keeper.WithdrawAllRewards.
func (suite *ModuleTestSuite) Rewards(farmerAcc sdk.AccAddress) sdk.Coins {
	cacheCtx, _ := suite.ctx.CacheContext()
	rewards, err := suite.keeper.WithdrawAllRewards(cacheCtx, farmerAcc)
	suite.Require().NoError(err)
	return rewards
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
