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

func (suite *KeeperTestSuite) TestGRPCStaking() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))

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
			"query by farmer addr",
			&types.QueryStakingRequest{Farmer: suite.addrs[0].String()},
			false,
			func(resp *types.QueryStakingResponse) {
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
			&types.QueryStakingRequest{Farmer: "invalid"},
			true,
			nil,
		},
		{
			"query with staking coin denom",
			&types.QueryStakingRequest{Farmer: suite.addrs[1].String(), StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryStakingResponse) {
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
			&types.QueryStakingRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
		{
			"query with staking coin denom, without farmer addr",
			&types.QueryStakingRequest{StakingCoinDenom: denom1},
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

func (suite *KeeperTestSuite) TestGRPCTotalStaking() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1500)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500), sdk.NewInt64Coin(denom2, 2000)))

	for _, tc := range []struct {
		name      string
		req       *types.QueryTotalStakingRequest
		expectErr bool
		postRun   func(*types.QueryTotalStakingResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"empty request",
			&types.QueryTotalStakingRequest{},
			true,
			nil,
		},
		{
			"query by staking coin denom #1",
			&types.QueryTotalStakingRequest{StakingCoinDenom: denom1},
			false,
			func(resp *types.QueryTotalStakingResponse) {
				suite.Require().True(intEq(sdk.NewInt(1500), resp.Amount))
			},
		},
		{
			"query by staking coin denom #1",
			&types.QueryTotalStakingRequest{StakingCoinDenom: denom2},
			false,
			func(resp *types.QueryTotalStakingResponse) {
				suite.Require().True(intEq(sdk.NewInt(3500), resp.Amount))
			},
		},
		{
			"invalid staking coin denom",
			&types.QueryTotalStakingRequest{StakingCoinDenom: "!"},
			true,
			nil,
		},
	} {
		resp, err := suite.querier.TotalStaking(sdk.WrapSDKContext(suite.ctx), tc.req)
		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
			tc.postRun(resp)
		}
	}
}
