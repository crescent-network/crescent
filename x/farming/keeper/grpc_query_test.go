package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/farming"
	"github.com/crescent-network/crescent/x/farming/types"
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
					suite.Require().True(plan.GetTerminated())
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
					suite.Require().False(plan.GetTerminated())
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

func (suite *KeeperTestSuite) TestGRPCStakings() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))

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
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)),
					resp.StakedCoins))
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)),
					resp.QueuedCoins))
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
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom2, 2000)),
					resp.StakedCoins))
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom2, 2000)),
					resp.QueuedCoins))
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

func (suite *KeeperTestSuite) TestGRPCTotalStakings() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)
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

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-06T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-07T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-08T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)

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
				suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1849999)), resp.Rewards))
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
				suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1150000)), resp.Rewards))
			},
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryRewardsRequest{StakingCoinDenom: denom2},
			true,
			nil,
		},
	} {
		resp, err := suite.querier.Rewards(sdk.WrapSDKContext(suite.ctx), tc.req)
		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			tc.postRun(resp)
		}
	}
}
