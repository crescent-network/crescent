package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (suite *KeeperTestSuite) TestDistributionInfos() {
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
				false,
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
				false,
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
						false,
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
			distrInfos := suite.keeper.DistributionInfos(suite.ctx)
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

func (suite *KeeperTestSuite) TestDistributeRewards() {
	plans := []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1))),
				mustParseRFC3339("2021-07-30T00:00:00Z"),
				mustParseRFC3339("2021-08-30T00:00:00Z"),
				false,
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000))),
	}
	for _, plan := range plans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.Stake(suite.addrs[1], sdk.NewCoins(
		sdk.NewInt64Coin(denom1, 1_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-07-31T00:00:00Z"))
	err := suite.keeper.DistributeRewards(suite.ctx)
	suite.Require().NoError(err)

	rewards := suite.keeper.GetRewardsByFarmer(suite.ctx, suite.addrs[1])
	suite.Require().Len(rewards, 2)
	for _, reward := range rewards {
		switch reward.StakingCoinDenom {
		case denom1:
			suite.Require().True(intEq(sdk.NewInt(300_000), reward.RewardCoins.AmountOf(denom3)))
		case denom2:
			suite.Require().True(intEq(sdk.NewInt(700_000), reward.RewardCoins.AmountOf(denom3)))
		}
	}

	totalRewardCoins := suite.keeper.GetTotalDistributedRewardCoins(suite.ctx, 1)
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000)), totalRewardCoins))
	lastDistributedAt, found := suite.keeper.GetLastDistributedTime(suite.ctx, 1)
	suite.Require().True(found)
	suite.Require().Equal(suite.ctx.BlockTime(), lastDistributedAt)
}
