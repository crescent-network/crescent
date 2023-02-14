package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/lpfarm/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCPairs() {
	privPlan, pubPlan := s.createSamplePlans()

	for _, tc := range []struct {
		name        string
		req         *types.QueryPlansRequest
		expectedErr string
		postRun     func(resp *types.QueryPlansResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query all plans",
			&types.QueryPlansRequest{},
			"",
			func(resp *types.QueryPlansResponse) {
				s.Require().Len(resp.Plans, 2)
				s.Require().Equal(privPlan, resp.Plans[0])
				s.Require().Equal(pubPlan, resp.Plans[1])
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Plans(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPair() {
	privPlan, pubPlan := s.createSamplePlans()

	for _, tc := range []struct {
		name        string
		req         *types.QueryPlanRequest
		expectedErr string
		postRun     func(resp *types.QueryPlanResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query a plan #1",
			&types.QueryPlanRequest{PlanId: privPlan.Id},
			"",
			func(resp *types.QueryPlanResponse) {
				s.Require().Equal(privPlan, resp.Plan)
			},
		},
		{
			"query a plan #2",
			&types.QueryPlanRequest{PlanId: pubPlan.Id},
			"",
			func(resp *types.QueryPlanResponse) {
				s.Require().Equal(pubPlan, resp.Plan)
			},
		},
		{
			"plan not found",
			&types.QueryPlanRequest{PlanId: 3},
			"rpc error: code = NotFound desc = plan not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Plan(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCFarm() {
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	farm, _ := s.keeper.GetFarm(s.ctx, "pool1")

	for _, tc := range []struct {
		name        string
		req         *types.QueryFarmRequest
		expectedErr string
		postRun     func(resp *types.QueryFarmResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query a farm",
			&types.QueryFarmRequest{Denom: "pool1"},
			"",
			func(resp *types.QueryFarmResponse) {
				s.Require().Equal(farm, resp.Farm)
			},
		},
		{
			"farm not found",
			&types.QueryFarmRequest{Denom: "pool2"},
			"rpc error: code = NotFound desc = farm not found",
			nil,
		},
		{
			"invalid denom",
			&types.QueryFarmRequest{Denom: "invalid!"},
			"rpc error: code = InvalidArgument desc = invalid denom: invalid denom: invalid!",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Farm(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPositions() {
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1,1_000000pool2"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	position1, _ := s.keeper.GetPosition(s.ctx, farmerAddr, "pool1")
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool2"))
	position2, _ := s.keeper.GetPosition(s.ctx, farmerAddr, "pool2")

	for _, tc := range []struct {
		name        string
		req         *types.QueryPositionsRequest
		expectedErr string
		postRun     func(resp *types.QueryPositionsResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query all positions by farmer",
			&types.QueryPositionsRequest{Farmer: farmerAddr.String()},
			"",
			func(resp *types.QueryPositionsResponse) {
				s.Require().Len(resp.Positions, 2)
				s.Require().Equal(position1, resp.Positions[0])
				s.Require().Equal(position2, resp.Positions[1])
			},
		},
		{
			"query all positions by unknown farmer",
			&types.QueryPositionsRequest{Farmer: utils.TestAddress(1).String()},
			"",
			func(resp *types.QueryPositionsResponse) {
				s.Require().Empty(resp.Positions)
			},
		},
		{
			"invalid farmer address",
			&types.QueryPositionsRequest{Farmer: "invalidaddr"},
			"rpc error: code = InvalidArgument desc = invalid farmer address: decoding bech32 failed: invalid separator index -1",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Positions(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPosition() {
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1,1_000000pool2"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	position1, _ := s.keeper.GetPosition(s.ctx, farmerAddr, "pool1")
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool2"))
	position2, _ := s.keeper.GetPosition(s.ctx, farmerAddr, "pool2")

	for _, tc := range []struct {
		name        string
		req         *types.QueryPositionRequest
		expectedErr string
		postRun     func(resp *types.QueryPositionResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query a position #1",
			&types.QueryPositionRequest{
				Farmer: farmerAddr.String(),
				Denom:  position1.Denom,
			},
			"",
			func(resp *types.QueryPositionResponse) {
				s.Require().Equal(position1, resp.Position)
			},
		},
		{
			"query a position #2",
			&types.QueryPositionRequest{
				Farmer: farmerAddr.String(),
				Denom:  position2.Denom,
			},
			"",
			func(resp *types.QueryPositionResponse) {
				s.Require().Equal(position2, resp.Position)
			},
		},
		{
			"position not found",
			&types.QueryPositionRequest{
				Farmer: farmerAddr.String(),
				Denom:  "pool3",
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
		{
			"invalid farmer address",
			&types.QueryPositionRequest{
				Farmer: "invalidaddr",
				Denom:  position1.Denom,
			},
			"rpc error: code = InvalidArgument desc = invalid farmer address: decoding bech32 failed: invalid separator index -1",
			nil,
		},
		{
			"invalid denom",
			&types.QueryPositionRequest{
				Farmer: farmerAddr.String(),
				Denom:  "invalid!",
			},
			"rpc error: code = InvalidArgument desc = invalid denom: invalid denom: invalid!",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Position(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCHistoricalRewards() {
	s.createSamplePlans()
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1,1_000000pool2"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool2"))
	s.nextBlock()
	_, _ = s.keeper.Unfarm(s.ctx, farmerAddr, utils.ParseCoin("500000pool1"))
	s.nextBlock()

	for _, tc := range []struct {
		name        string
		req         *types.QueryHistoricalRewardsRequest
		expectedErr string
		postRun     func(resp *types.QueryHistoricalRewardsResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query historical rewards #1",
			&types.QueryHistoricalRewardsRequest{
				Denom: "pool1",
			},
			"",
			func(resp *types.QueryHistoricalRewardsResponse) {
				s.Require().Len(resp.HistoricalRewards, 1)
				hist, found := s.keeper.GetHistoricalRewards(s.ctx, "pool1", 2)
				s.Require().True(found)
				s.Require().EqualValues(2, resp.HistoricalRewards[0].Period)
				s.Require().Equal(
					hist.CumulativeUnitRewards,
					resp.HistoricalRewards[0].CumulativeUnitRewards)
				s.Require().Equal(
					hist.ReferenceCount, resp.HistoricalRewards[0].ReferenceCount)
			},
		},
		{
			"query historical rewards #2",
			&types.QueryHistoricalRewardsRequest{
				Denom: "pool2",
			},
			"",
			func(resp *types.QueryHistoricalRewardsResponse) {
				s.Require().Len(resp.HistoricalRewards, 1)
				hist, found := s.keeper.GetHistoricalRewards(s.ctx, "pool2", 1)
				s.Require().True(found)
				s.Require().EqualValues(1, resp.HistoricalRewards[0].Period)
				s.Require().Equal(
					hist.CumulativeUnitRewards,
					resp.HistoricalRewards[0].CumulativeUnitRewards)
				s.Require().Equal(
					hist.ReferenceCount, resp.HistoricalRewards[0].ReferenceCount)
			},
		},
		{
			"query unknown historical rewards",
			&types.QueryHistoricalRewardsRequest{
				Denom: "pool3",
			},
			"",
			func(resp *types.QueryHistoricalRewardsResponse) {
				s.Require().Empty(resp.HistoricalRewards)
			},
		},
		{
			"invalid denom",
			&types.QueryHistoricalRewardsRequest{
				Denom: "invalid!",
			},
			"rpc error: code = InvalidArgument desc = invalid denom: invalid denom: invalid!",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.HistoricalRewards(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCTotalRewards() {
	s.createSamplePlans()
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1,1_000000pool2"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool2"))
	s.nextBlock()
	_, _ = s.keeper.Unfarm(s.ctx, farmerAddr, utils.ParseCoin("500000pool1"))
	s.nextBlock()

	for _, tc := range []struct {
		name        string
		req         *types.QueryTotalRewardsRequest
		expectedErr string
		postRun     func(resp *types.QueryTotalRewardsResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query total rewards",
			&types.QueryTotalRewardsRequest{
				Farmer: farmerAddr.String(),
			},
			"",
			func(resp *types.QueryTotalRewardsResponse) {
				s.assertEq(utils.ParseDecCoins("28935stake"), resp.Rewards)
			},
		},
		{
			"query unknown farmer",
			&types.QueryTotalRewardsRequest{
				Farmer: utils.TestAddress(1).String(),
			},
			"",
			func(resp *types.QueryTotalRewardsResponse) {
				s.assertEq(sdk.DecCoins{}, resp.Rewards)
			},
		},
		{
			"invalid farmer address",
			&types.QueryTotalRewardsRequest{
				Farmer: "invalidaddr",
			},
			"rpc error: code = InvalidArgument desc = invalid farmer address: decoding bech32 failed: invalid separator index -1",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.TotalRewards(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCRewards() {
	s.createSamplePlans()
	farmerAddr := utils.TestAddress(0)
	s.fundAddr(farmerAddr, utils.ParseCoins("1_000000pool1,1_000000pool2"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	_, _ = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool2"))
	s.nextBlock()
	_, _ = s.keeper.Unfarm(s.ctx, farmerAddr, utils.ParseCoin("500000pool1"))
	s.nextBlock()

	for _, tc := range []struct {
		name        string
		req         *types.QueryRewardsRequest
		expectedErr string
		postRun     func(resp *types.QueryRewardsResponse)
	}{
		{
			"nil request",
			nil,
			"rpc error: code = InvalidArgument desc = empty request",
			nil,
		},
		{
			"query rewards #1",
			&types.QueryRewardsRequest{
				Farmer: farmerAddr.String(),
				Denom:  "pool1",
			},
			"",
			func(resp *types.QueryRewardsResponse) {
				s.assertEq(utils.ParseDecCoins("5787stake"), resp.Rewards)
			},
		},
		{
			"query rewards #2",
			&types.QueryRewardsRequest{
				Farmer: farmerAddr.String(),
				Denom:  "pool2",
			},
			"",
			func(resp *types.QueryRewardsResponse) {
				s.assertEq(utils.ParseDecCoins("23148stake"), resp.Rewards)
			},
		},
		{
			"query unknown farmer",
			&types.QueryRewardsRequest{
				Farmer: utils.TestAddress(1).String(),
				Denom:  "pool1",
			},
			"",
			func(resp *types.QueryRewardsResponse) {
				s.assertEq(sdk.DecCoins{}, resp.Rewards)
			},
		},
		{
			"query unknown denom",
			&types.QueryRewardsRequest{
				Farmer: farmerAddr.String(),
				Denom:  "pool3",
			},
			"",
			func(resp *types.QueryRewardsResponse) {
				s.assertEq(sdk.DecCoins{}, resp.Rewards)
			},
		},
		{
			"invalid farmer address",
			&types.QueryRewardsRequest{
				Farmer: "invalidaddr",
				Denom:  "pool1",
			},
			"rpc error: code = InvalidArgument desc = invalid farmer address: decoding bech32 failed: invalid separator index -1",
			nil,
		},
		{
			"invalid denom",
			&types.QueryRewardsRequest{
				Farmer: farmerAddr.String(),
				Denom:  "invalid!",
			},
			"rpc error: code = InvalidArgument desc = invalid denom: invalid denom: invalid!",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Rewards(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}
