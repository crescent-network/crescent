package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (suite *KeeperTestSuite) TestAllocationInfos() {
	normalPlans := []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom1, sdk.NewDec(1))),
				mustParseRFC3339("2021-07-27T00:00:00Z"),
				mustParseRFC3339("2021-07-28T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))),
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				2,
				"",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom1, sdk.NewDec(1))),
				mustParseRFC3339("2021-07-27T12:00:00Z"),
				mustParseRFC3339("2021-07-28T12:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))),
	}

	for _, tc := range []struct {
		name      string
		plans     []types.PlanI
		t         time.Time
		distrAmts map[uint64]sdk.Coins // planID => sdk.Coins
	}{
		{
			"insufficient farming pool balances",
			[]types.PlanI{
				types.NewFixedAmountPlan(
					types.NewBasePlan(
						1,
						"",
						types.PlanTypePrivate,
						suite.addrs[0].String(),
						suite.addrs[0].String(),
						sdk.NewDecCoins(sdk.NewDecCoinFromDec(denom1, sdk.NewDec(1))),
						mustParseRFC3339("2021-07-27T00:00:00Z"),
						mustParseRFC3339("2021-07-30T00:00:00Z"),
					),
					sdk.NewCoins(sdk.NewInt64Coin(denom3, 10_000_000_000))),
			},
			mustParseRFC3339("2021-07-28T00:00:00Z"),
			nil,
		},
		{
			"start time & end time edgecase #1",
			normalPlans,
			mustParseRFC3339("2021-07-26T23:59:59Z"),
			nil,
		},
		{
			"start time & end time edgecase #2",
			normalPlans,
			mustParseRFC3339("2021-07-27T00:00:00Z"),
			map[uint64]sdk.Coins{1: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #3",
			normalPlans,
			mustParseRFC3339("2021-07-27T11:59:59Z"),
			map[uint64]sdk.Coins{1: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #4",
			normalPlans,
			mustParseRFC3339("2021-07-27T12:00:00Z"),
			map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)),
				2: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #5",
			normalPlans,
			mustParseRFC3339("2021-07-27T23:59:59Z"),
			map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)),
				2: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #6",
			normalPlans,
			mustParseRFC3339("2021-07-28T00:00:00Z"),
			map[uint64]sdk.Coins{2: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #7",
			normalPlans,
			mustParseRFC3339("2021-07-28T11:59:59Z"),
			map[uint64]sdk.Coins{2: sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000))},
		},
		{
			"start time & end time edgecase #8",
			normalPlans,
			mustParseRFC3339("2021-07-28T12:00:00Z"),
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			for _, plan := range tc.plans {
				suite.keeper.SetPlan(suite.ctx, plan)
			}

			suite.ctx = suite.ctx.WithBlockTime(tc.t)
			distrInfos := suite.keeper.AllocationInfos(suite.ctx)
			if suite.Len(distrInfos, len(tc.distrAmts)) {
				for _, distrInfo := range distrInfos {
					distrAmt, ok := tc.distrAmts[distrInfo.Plan.GetId()]
					if suite.True(ok) {
						suite.True(coinsEq(distrAmt, distrInfo.Amount))
					}
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAllocateRewards() {
	for _, plan := range suite.sampleFixedAmtPlans {
		_ = plan.SetStartTime(mustParseRFC3339("0001-01-01T00:00:00Z"))
		_ = plan.SetEndTime(mustParseRFC3339("9999-12-31T00:00:00Z"))
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	prevDistrCoins := map[uint64]sdk.Coins{}

	t := mustParseRFC3339("2021-09-01T00:00:00Z")
	for i := 0; i < 365; i++ {
		suite.ctx = suite.ctx.WithBlockTime(t)

		err := suite.keeper.AllocateRewards(suite.ctx)
		suite.Require().NoError(err)

		for _, plan := range suite.sampleFixedAmtPlans {
			plan, _ := suite.keeper.GetPlan(suite.ctx, plan.GetId())
			fixedAmtPlan := plan.(*types.FixedAmountPlan)

			dist := plan.GetDistributedCoins()
			suite.Require().True(coinsEq(prevDistrCoins[plan.GetId()].Add(fixedAmtPlan.EpochAmount...), dist))
			prevDistrCoins[plan.GetId()] = dist

			t2 := plan.GetLastDistributionTime()
			suite.Require().NotNil(t2)
			suite.Require().Equal(t, *t2)
		}

		t = t.AddDate(0, 0, 1)
	}
}

func (suite *KeeperTestSuite) TestHarvest() {
	for _, plan := range suite.samplePlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-08-05T00:00:00Z"))
	err := suite.keeper.AllocateRewards(suite.ctx)
	suite.Require().NoError(err)

	rewards := suite.Rewards(suite.addrs[0])

	err = suite.keeper.Harvest(suite.ctx, suite.addrs[0], []string{denom1})
	suite.Require().NoError(err)

	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(balancesBefore.Add(rewards...), balancesAfter))
	suite.Require().True(suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.keeper.GetRewardsReservePoolAcc(suite.ctx)).IsZero())
	suite.Require().True(suite.Rewards(suite.addrs[0]).IsZero())
}
