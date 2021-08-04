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
	for _, plan := range suite.samplePlans {
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
			"query by farmer address",
			&types.QueryStakingsRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryStakingsResponse) {
				suite.Require().Len(resp.Stakings, 1)
				suite.Require().Equal(suite.addrs[0], resp.Stakings[0].GetFarmer())
			},
		},
		{
			"invalid farmer address",
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

func (suite *KeeperTestSuite) TestGRPCRewards() {
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
