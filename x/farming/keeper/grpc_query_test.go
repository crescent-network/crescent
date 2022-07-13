package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	resp, err := suite.querier.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	suite.Require().Equal(suite.keeper.GetParams(suite.ctx), resp.Params)
}

func (suite *KeeperTestSuite) TestGRPCPlans() {
	for i, plan := range suite.samplePlans {
		if i == 1 || i == 3 { // Mark 2nd and 4th plans as terminated. This is just for testing query.
			_ = plan.SetTerminated(true)
		}
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	for _, tc := range []struct {
		name      string
		req       *types.QueryPlansRequest
		expectErr bool
		postRun   func(*types.QueryPlansResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryPlansRequest{},
			false,
			func(resp *types.QueryPlansResponse) {
				suite.Require().Len(resp.Plans, 4)
			},
		},
		{
			"invalid type",
			&types.QueryPlansRequest{
				Type: "invalid",
			},
			true,
			nil,
		},
		{
			"query by type",
			&types.QueryPlansRequest{Type: types.PlanTypePrivate.String()},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 2)
				for _, plan := range plans {
					suite.Require().Equal(types.PlanTypePrivate, plan.GetType())
				}
			},
		},
		{
			"invalid farming pool addr",
			&types.QueryPlansRequest{FarmingPoolAddress: "invalid"},
			true,
			nil,
		},
		{
			"query by farming pool addr",
			&types.QueryPlansRequest{FarmingPoolAddress: suite.addrs[4].String()},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 2)
				for _, plan := range plans {
					suite.Require().True(plan.GetFarmingPoolAddress().Equals(suite.addrs[4]))
				}
			},
		},
		{
			"invalid termination addr",
			&types.QueryPlansRequest{TerminationAddress: "invalid"},
			true,
			nil,
		},
		{
			"query by termination addr",
			&types.QueryPlansRequest{TerminationAddress: suite.addrs[4].String()},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 2)
				for _, plan := range plans {
					suite.Require().True(plan.GetTerminationAddress().Equals(suite.addrs[4]))
				}
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryPlansRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
		{
			"query by staking coin denom",
			&types.QueryPlansRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 3)
				for _, plan := range plans {
					found := false
					for _, coinWeight := range plan.GetStakingCoinWeights() {
						if coinWeight.Denom == denom1 {
							found = true
							break
						}
					}
					suite.Require().True(found)
				}
			},
		},
		{
			"invalid terminated",
			&types.QueryPlansRequest{Terminated: "invalid"},
			true,
			nil,
		},
		{
			"query by terminated(true)",
			&types.QueryPlansRequest{Terminated: "true"},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 2)
				for _, plan := range plans {
					suite.Require().True(plan.IsTerminated())
				}
			},
		},
		{
			"query by terminated(false)",
			&types.QueryPlansRequest{Terminated: "false"},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 2)
				for _, plan := range plans {
					suite.Require().False(plan.IsTerminated())
				}
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Plans(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCPlan() {
	for _, plan := range suite.samplePlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	for _, tc := range []struct {
		name      string
		req       *types.QueryPlanRequest
		expectErr bool
		postRun   func(*types.QueryPlanResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by id",
			&types.QueryPlanRequest{PlanId: 1},
			false,
			func(resp *types.QueryPlanResponse) {
				plan, err := types.UnpackPlan(resp.Plan)
				suite.Require().NoError(err)
				suite.Require().Equal(plan.GetId(), uint64(1))
			},
		},
		{
			"id not found",
			&types.QueryPlanRequest{PlanId: 5},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Plan(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCPosition() {
	for _, plan := range suite.sampleFixedAmtPlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2021-08-05T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.advanceEpochDays()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))

	for _, tc := range []struct {
		name      string
		req       *types.QueryPositionRequest
		expectErr bool
		postRun   func(*types.QueryPositionResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by farmer addr",
			&types.QueryPositionRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryPositionResponse) {
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)),
					resp.StakedCoins))
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)),
					resp.QueuedCoins))
				suite.Require().True(coinsEq(utils.ParseCoins("1833333denom3"), resp.Rewards))
			},
		},
		{
			"invalid farmer addr",
			&types.QueryPositionRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryPositionRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryPositionResponse) {
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom2, 2000)),
					resp.StakedCoins))
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom2, 2000)),
					resp.QueuedCoins))
				suite.Require().True(coinsEq(utils.ParseCoins("400000denom3"), resp.Rewards))
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryPositionRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryPositionRequest{StakingCoinDenom: denom1},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Position(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCStakings() {
	suite.executeBlock(utils.ParseTime("2022-04-01T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	})
	suite.executeBlock(utils.ParseTime("2022-04-02T00:00:00Z"), nil)
	suite.executeBlock(utils.ParseTime("2022-04-02T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryStakingsRequest
		expectErr bool
		postRun   func(*types.QueryStakingsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by farmer addr",
			&types.QueryStakingsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 2)
				for _, staking := range resp.Stakings {
					switch staking.StakingCoinDenom {
					case denom1:
						suite.Require().True(intEq(sdk.NewInt(1000), staking.Amount))
						suite.Require().EqualValues(1, staking.StartingEpoch)
					case denom2:
						suite.Require().True(intEq(sdk.NewInt(1500), staking.Amount))
						suite.Require().EqualValues(1, staking.StartingEpoch)
					default:
						suite.FailNowf("invalid staking coin denom: %s", staking.StakingCoinDenom)
					}
				}
			},
		},
		{
			"invalid farmer addr",
			&types.QueryStakingsRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryStakingsRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 1)
				suite.Require().Equal(denom2, resp.Stakings[0].StakingCoinDenom)
				suite.Require().True(intEq(sdk.NewInt(2000), resp.Stakings[0].Amount))
				suite.Require().EqualValues(1, resp.Stakings[0].StartingEpoch)
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryStakingsRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryStakingsRequest{StakingCoinDenom: denom1},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Stakings(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueuedStakings() {
	suite.executeBlock(utils.ParseTime("2022-04-01T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	})
	suite.executeBlock(utils.ParseTime("2022-04-02T00:00:00Z"), nil)
	suite.executeBlock(utils.ParseTime("2022-04-02T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryQueuedStakingsRequest
		expectErr bool
		postRun   func(*types.QueryQueuedStakingsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by farmer addr",
			&types.QueryQueuedStakingsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryQueuedStakingsResponse) {
				suite.Require().Len(resp.QueuedStakings, 2)
				for _, queuedStaking := range resp.QueuedStakings {
					switch queuedStaking.StakingCoinDenom {
					case denom1:
						suite.Require().True(intEq(sdk.NewInt(1000), queuedStaking.Amount))
						suite.Require().Equal(utils.ParseTime("2022-04-03T23:00:00Z"), queuedStaking.EndTime)
					case denom2:
						suite.Require().True(intEq(sdk.NewInt(1500), queuedStaking.Amount))
						suite.Require().Equal(utils.ParseTime("2022-04-03T23:00:00Z"), queuedStaking.EndTime)
					default:
						suite.FailNowf("invalid staking coin denom: %s", queuedStaking.StakingCoinDenom)
					}
				}
			},
		},
		{
			"invalid farmer addr",
			&types.QueryQueuedStakingsRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryQueuedStakingsRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryQueuedStakingsResponse) {
				suite.Require().Len(resp.QueuedStakings, 1)
				suite.Require().Equal(denom2, resp.QueuedStakings[0].StakingCoinDenom)
				suite.Require().True(intEq(sdk.NewInt(2000), resp.QueuedStakings[0].Amount))
				suite.Require().Equal(utils.ParseTime("2022-04-03T23:00:00Z"), resp.QueuedStakings[0].EndTime)
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryQueuedStakingsRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryQueuedStakingsRequest{StakingCoinDenom: denom1},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.QueuedStakings(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCTotalStakings() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.advanceEpochDays()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))

	for _, tc := range []struct {
		name      string
		req       *types.QueryTotalStakingsRequest
		expectErr bool
		postRun   func(*types.QueryTotalStakingsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"empty request",
			&types.QueryTotalStakingsRequest{},
			true,
			nil,
		},
		{
			"query by staking coin denom #1",
			&types.QueryTotalStakingsRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryTotalStakingsResponse) {
				suite.Require().True(intEq(sdk.NewInt(1500), resp.Amount))
			},
		},
		{
			"query by staking coin denom #1",
			&types.QueryTotalStakingsRequest{StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryTotalStakingsResponse) {
				suite.Require().True(intEq(sdk.NewInt(3500), resp.Amount))
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryTotalStakingsRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
	} {
		resp, err := suite.querier.TotalStakings(sdk.WrapSDKContext(suite.ctx), tc.req)
		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			tc.postRun(resp)
		}
	}
}

func (suite *KeeperTestSuite) TestGRPCRewards() {
	for _, plan := range suite.sampleFixedAmtPlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-05T12:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-06T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // next epoch

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-06T12:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // queued -> staked

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-07T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // rewards distribution

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsRequest
		expectErr bool
		postRun   func(*types.QueryRewardsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"empty request",
			&types.QueryRewardsRequest{},
			true,
			nil,
		},
		{
			"query by farmer addr",
			&types.QueryRewardsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryRewardsResponse) {
				// 0.3 * 1000000 * 1/2
				// + 0.7 * 1000000 * 1/1
				// + 1.0 * 2000000 * 1/2
				// ~= 1850000
				suite.Require().Len(resp.Rewards, 2)
				for _, r := range resp.Rewards {
					switch r.StakingCoinDenom {
					case denom1:
						suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1150000)), r.Rewards))
					case denom2:
						suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 699999)), r.Rewards))
					}
				}
			},
		},
		{
			"invalid farmer addr",
			&types.QueryRewardsRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryRewardsRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryRewardsResponse) {
				// 0.3 * 1000000 * 1/2
				// + 1.0 * 2000000 * 1/2
				// ~= 1150000
				suite.Require().Len(resp.Rewards, 1)
				suite.Require().Equal(denom1, resp.Rewards[0].StakingCoinDenom)
				suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1150000)), resp.Rewards[0].Rewards))
			},
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryRewardsRequest{StakingCoinDenom: denom2},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Rewards(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCUnharvestedRewards() {
	for _, plan := range suite.sampleFixedAmtPlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.executeBlock(utils.ParseTime("2021-08-04T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
	})

	suite.executeBlock(utils.ParseTime("2021-08-05T00:00:00Z"), nil) // next epoch
	suite.executeBlock(utils.ParseTime("2021-08-05T23:00:00Z"), nil) // queued -> staked

	// Stake more.
	suite.executeBlock(utils.ParseTime("2021-08-05T23:30:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
	})

	suite.executeBlock(utils.ParseTime("2021-08-06T00:00:00Z"), nil) // rewards distribution
	suite.executeBlock(utils.ParseTime("2021-08-06T23:30:00Z"), nil) // queued -> staked

	for _, tc := range []struct {
		name      string
		req       *types.QueryUnharvestedRewardsRequest
		expectErr bool
		postRun   func(*types.QueryUnharvestedRewardsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"empty request",
			&types.QueryUnharvestedRewardsRequest{},
			true,
			nil,
		},
		{
			"query by farmer addr",
			&types.QueryUnharvestedRewardsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryUnharvestedRewardsResponse) {
				suite.Require().Len(resp.UnharvestedRewards, 2)
				for _, unharvestedRewards := range resp.UnharvestedRewards {
					switch unharvestedRewards.StakingCoinDenom {
					case denom1:
						// 0.3 * 1000000 * 1/2
						// + 1.0 * 2000000 * 1/2
						// ~= 1150000
						suite.Require().True(coinsEq(utils.ParseCoins("1150000denom3"), unharvestedRewards.Rewards))
					case denom2:
						// 0.7 * 1000000 * 1/1
						// ~= 700000
						suite.Require().True(coinsEq(utils.ParseCoins("699999denom3"), unharvestedRewards.Rewards))
					default:
						suite.FailNowf("invalid staking coin denom: %s", unharvestedRewards.StakingCoinDenom)
					}
				}
			},
		},
		{
			"invalid farmer addr",
			&types.QueryUnharvestedRewardsRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryUnharvestedRewardsRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryUnharvestedRewardsResponse) {
				// 0.3 * 1000000 * 1/2
				// + 1.0 * 2000000 * 1/2
				// ~= 1150000
				suite.Require().Len(resp.UnharvestedRewards, 1)
				suite.Require().True(coinsEq(utils.ParseCoins("1150000denom3"), resp.UnharvestedRewards[0].Rewards))
			},
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryUnharvestedRewardsRequest{StakingCoinDenom: denom2},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.UnharvestedRewards(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCHistoricalRewards() {
	for _, plan := range suite.sampleFixedAmtPlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.executeBlock(utils.ParseTime("2021-08-04T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
	})

	suite.executeBlock(utils.ParseTime("2021-08-05T00:00:00Z"), nil) // next epoch
	suite.executeBlock(utils.ParseTime("2021-08-05T23:00:00Z"), nil) // queued -> staked

	// Stake more.
	suite.executeBlock(utils.ParseTime("2021-08-05T23:30:00Z"), func() {
		suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
		suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
	})

	suite.executeBlock(utils.ParseTime("2021-08-06T00:00:00Z"), nil) // rewards distribution
	suite.executeBlock(utils.ParseTime("2021-08-06T23:30:00Z"), nil) // queued -> staked

	for _, tc := range []struct {
		name      string
		req       *types.QueryHistoricalRewardsRequest
		expectErr bool
		postRun   func(response *types.QueryHistoricalRewardsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"empty request",
			&types.QueryHistoricalRewardsRequest{},
			true,
			nil,
		},
		{
			"query for a staking coin denom",
			&types.QueryHistoricalRewardsRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryHistoricalRewardsResponse) {
				suite.Require().Len(resp.HistoricalRewards, 2)
				suite.Require().EqualValues(0, resp.HistoricalRewards[0].Epoch)
				suite.Require().True(decCoinsEq(sdk.DecCoins{}, resp.HistoricalRewards[0].CumulativeUnitRewards))
				suite.Require().EqualValues(1, resp.HistoricalRewards[1].Epoch)
				suite.Require().True(decCoinsEq(utils.ParseDecCoins("1150denom3"), resp.HistoricalRewards[1].CumulativeUnitRewards))
			},
		},
		{
			"query for a staking coin denom 2",
			&types.QueryHistoricalRewardsRequest{StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryHistoricalRewardsResponse) {
				suite.Require().Len(resp.HistoricalRewards, 2)
				suite.Require().EqualValues(0, resp.HistoricalRewards[0].Epoch)
				suite.Require().True(decCoinsEq(sdk.DecCoins{}, resp.HistoricalRewards[0].CumulativeUnitRewards))
				suite.Require().EqualValues(1, resp.HistoricalRewards[1].Epoch)
				suite.Require().True(decCoinsEq(utils.ParseDecCoins("466.666666666666666666denom3"), resp.HistoricalRewards[1].CumulativeUnitRewards))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.HistoricalRewards(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
