package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
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
			"invalid reward pool addr",
			&types.QueryPlansRequest{RewardPoolAddress: "invalid"},
			true,
			nil,
		},
		{
			"query by reward pool addr",
			&types.QueryPlansRequest{
				RewardPoolAddress: suite.samplePlans[0].GetRewardPoolAddress().String(),
			},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				suite.Require().NoError(err)
				suite.Require().Len(plans, 1)
				suite.Require().True(plans[0].GetRewardPoolAddress().Equals(suite.samplePlans[0].GetRewardPoolAddress()))
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
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)))
	suite.Stake(suite.addrs[2], sdk.NewCoins(sdk.NewInt64Coin(denom2, 800)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	for _, tc := range []struct {
		name      string
		req       *types.QueryStakingsRequest
		expectErr bool
		postRun   func(response *types.QueryStakingsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryStakingsRequest{},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 3)
			},
		},
		{
			"query by farmer addr",
			&types.QueryStakingsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 1)
				suite.Require().Equal(suite.addrs[0], resp.Stakings[0].GetFarmer())
			},
		},
		{
			"invalid farmer addr",
			&types.QueryStakingsRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query by staking coin denom",
			&types.QueryStakingsRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 2)
				for _, staking := range resp.Stakings {
					suite.Require().True(staking.StakedCoins.Add(staking.QueuedCoins...).AmountOf(denom1).IsPositive())
				}
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryStakingsRequest{Farmer: "!!!"},
			true,
			nil,
		},
		{
			"query by farmer addr and staking coin denom",
			&types.QueryStakingsRequest{
				Farmer:           suite.addrs[0].String(),
				StakingCoinDenom: denom1,
			},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 1)
				suite.Require().Equal(suite.addrs[0], resp.Stakings[0].GetFarmer())
				suite.Require().True(
					resp.Stakings[0].StakedCoins.Add(resp.Stakings[0].QueuedCoins...).AmountOf(denom1).IsPositive())
			},
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

func (suite *KeeperTestSuite) TestGRPCStaking() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)))
	suite.Stake(suite.addrs[2], sdk.NewCoins(sdk.NewInt64Coin(denom2, 800)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	for _, tc := range []struct {
		name      string
		req       *types.QueryStakingRequest
		expectErr bool
		postRun   func(*types.QueryStakingResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by id",
			&types.QueryStakingRequest{StakingId: 1},
			false,
			func(resp *types.QueryStakingResponse) {
				suite.Require().Equal(uint64(1), resp.Staking.Id)
			},
		},
		{
			"id not found",
			&types.QueryStakingRequest{StakingId: 10},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Staking(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCRewards() {
	// Set rewards manually for testing.
	// Actual reward distribution doesn't work like this.
	suite.keeper.SetReward(suite.ctx, denom1, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)))
	suite.keeper.SetReward(suite.ctx, denom2, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)))
	suite.keeper.SetReward(suite.ctx, denom1, suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)))
	suite.keeper.SetReward(suite.ctx, denom2, suite.addrs[2], sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000)))

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsRequest
		expectErr bool
		postRun   func(response *types.QueryRewardsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryRewardsRequest{},
			false,
			func(resp *types.QueryRewardsResponse) {
				suite.Require().Len(resp.Rewards, 4)
			},
		},
		{
			"query by farmer addr",
			&types.QueryRewardsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryRewardsResponse) {
				suite.Require().Len(resp.Rewards, 2)
				for _, reward := range resp.Rewards {
					suite.Require().Equal(suite.addrs[0].String(), reward.Farmer)
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
			"query by staking coin denom",
			&types.QueryRewardsRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryRewardsResponse) {
				suite.Require().Len(resp.Rewards, 2)
				for _, reward := range resp.Rewards {
					suite.Require().Equal(denom1, reward.StakingCoinDenom)
				}
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryRewardsRequest{StakingCoinDenom: "!!!"},
			true,
			nil,
		},
		{
			"query by farmer addr and staking coin denom",
			&types.QueryRewardsRequest{
				Farmer:           suite.addrs[0].String(),
				StakingCoinDenom: denom1,
			},
			false,
			func(resp *types.QueryRewardsResponse) {
				suite.Require().Len(resp.Rewards, 1)
				suite.Require().Equal(suite.addrs[0], resp.Rewards[0].GetFarmer())
				suite.Require().Equal(denom1, resp.Rewards[0].StakingCoinDenom)
			},
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
