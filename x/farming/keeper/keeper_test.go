package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
		sdk.NewInt64Coin(denom1, 1_000_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000_000),
		sdk.NewInt64Coin(denom3, 1_000_000_000))
)

type KeeperTestSuite struct {
	suite.Suite

	app                 *simapp.FarmingApp
	ctx                 sdk.Context
	keeper              keeper.Keeper
	querier             keeper.Querier
	addrs               []sdk.AccAddress
	sampleFixedAmtPlans []types.PlanI
	sampleRatioPlans    []types.PlanI
	samplePlans         []types.PlanI
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.keeper = suite.app.FarmingKeeper
	suite.querier = keeper.Querier{Keeper: suite.keeper}
	suite.addrs = simapp.AddTestAddrs(suite.app, suite.ctx, 6, sdk.ZeroInt())
	for _, addr := range suite.addrs {
		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
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
				mustParseRFC3339("2021-08-02T00:00:00Z"),
				mustParseRFC3339("2021-08-10T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)),
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
				mustParseRFC3339("2021-08-04T00:00:00Z"),
				mustParseRFC3339("2021-08-12T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 2000000)),
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
				mustParseRFC3339("2021-08-01T00:00:00Z"),
				mustParseRFC3339("2021-08-09T00:00:00Z"),
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
				mustParseRFC3339("2021-08-03T00:00:00Z"),
				mustParseRFC3339("2021-08-07T00:00:00Z"),
			),
			sdk.NewDecWithPrec(3, 2), // 3%
		),
	}
	suite.samplePlans = append(suite.sampleFixedAmtPlans, suite.sampleRatioPlans...)
}

// Stake is a convenient method to test Keeper.Stake.
func (suite *KeeperTestSuite) Stake(farmerAcc sdk.AccAddress, amt sdk.Coins) {
	err := suite.keeper.Stake(suite.ctx, farmerAcc, amt)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) StakedCoins(farmerAcc sdk.AccAddress) sdk.Coins {
	stakedCoins := sdk.NewCoins()
	suite.keeper.IterateStakingsByFarmer(suite.ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		stakedCoins = stakedCoins.Add(sdk.NewCoin(stakingCoinDenom, staking.Amount))
		return false
	})
	return stakedCoins
}

func (suite *KeeperTestSuite) QueuedCoins(farmerAcc sdk.AccAddress) sdk.Coins {
	queuedCoins := sdk.NewCoins()
	suite.keeper.IterateQueuedStakingsByFarmer(suite.ctx, farmerAcc, func(stakingCoinDenom string, queuedStaking types.QueuedStaking) (stop bool) {
		queuedCoins = queuedCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
		return false
	})
	return queuedCoins
}

func (suite *KeeperTestSuite) Rewards(farmerAcc sdk.AccAddress) sdk.Coins {
	cacheCtx, _ := suite.ctx.CacheContext()
	rewards, err := suite.keeper.WithdrawAllRewards(cacheCtx, farmerAcc)
	suite.Require().NoError(err)
	return rewards
}

func intEq(exp, got sdk.Int) (bool, string, string, string) {
	return exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func mustParseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
