package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

func (suite *KeeperTestSuite) TestStakingReservedAmountInvariant() {
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
				mustParseRFC3339("2021-08-30T00:00:00Z")),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000))),
	}
	for _, plan := range plans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.Stake(suite.addrs[1], sdk.NewCoins(
		sdk.NewInt64Coin(denom1, 1_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000)))

	// invariants was not broken
	keeper.AllInvariants(suite.keeper)
	invariant := keeper.AllInvariants(suite.keeper)
	_, broken := invariant(suite.ctx)
	suite.False(broken)

	// manipulate staking reserved amount
	stakings := suite.keeper.GetAllStakings(suite.ctx)
	stakings[0].QueuedCoins = stakings[0].QueuedCoins.Add(sdk.NewInt64Coin(denom1, 1))
	suite.keeper.SetStaking(suite.ctx, stakings[0])

	// invariants was broken
	keeper.AllInvariants(suite.keeper)
	invariant = keeper.AllInvariants(suite.keeper)
	_, broken = invariant(suite.ctx)
	suite.True(broken)
}

func (suite *KeeperTestSuite) TestRemainingRewardsAmountInvariant() {
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
				mustParseRFC3339("2021-08-30T00:00:00Z")),
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

	// invariants was not broken
	keeper.AllInvariants(suite.keeper)
	invariant := keeper.AllInvariants(suite.keeper)
	_, broken := invariant(suite.ctx)
	suite.False(broken)

	// manipulate reward coins
	rewards[0].RewardCoins = rewards[0].RewardCoins.Add(sdk.NewInt64Coin(denom3, 1))
	suite.keeper.SetReward(suite.ctx, rewards[0].StakingCoinDenom, suite.addrs[1], rewards[0].RewardCoins)

	// invariants was broken
	keeper.AllInvariants(suite.keeper)
	invariant = keeper.AllInvariants(suite.keeper)
	_, broken = invariant(suite.ctx)
	suite.True(broken)
}
