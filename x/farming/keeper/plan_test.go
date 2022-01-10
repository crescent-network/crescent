package keeper_test

import (
	"github.com/crescent-network/crescent/x/farming/types"
)

func (suite *KeeperTestSuite) TestGlobalPlanId() {
	globalPlanId := suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(0), globalPlanId)

	cacheCtx, _ := suite.ctx.CacheContext()
	nextPlanId := suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(1), nextPlanId)

	sampleFixedPlan := suite.sampleFixedAmtPlans[0].(*types.FixedAmountPlan)
	poolAcc, err := suite.keeper.DerivePrivatePlanFarmingPoolAcc(suite.ctx, sampleFixedPlan.Name)
	suite.Require().NoError(err)
	_, err = suite.keeper.CreateFixedAmountPlan(suite.ctx, &types.MsgCreateFixedAmountPlan{
		Name:               sampleFixedPlan.Name,
		Creator:            suite.addrs[0].String(),
		StakingCoinWeights: sampleFixedPlan.GetStakingCoinWeights(),
		StartTime:          sampleFixedPlan.GetStartTime(),
		EndTime:            sampleFixedPlan.GetEndTime(),
		EpochAmount:        sampleFixedPlan.EpochAmount,
	}, poolAcc, suite.addrs[0], types.PlanTypePublic)
	suite.Require().NoError(err)

	globalPlanId = suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(1), globalPlanId)

	plans := suite.keeper.GetPlans(suite.ctx)
	suite.Require().Len(plans, 1)
	suite.Require().Equal(uint64(len(plans)), globalPlanId)

	cacheCtx, _ = suite.ctx.CacheContext()
	nextPlanId = suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(2), nextPlanId)

	sampleRatioPlan := suite.sampleRatioPlans[0].(*types.RatioPlan)
	poolAcc, err = suite.keeper.DerivePrivatePlanFarmingPoolAcc(suite.ctx, sampleRatioPlan.Name)
	suite.Require().NoError(err)
	_, err = suite.keeper.CreateRatioPlan(suite.ctx, &types.MsgCreateRatioPlan{
		Name:               sampleRatioPlan.Name,
		Creator:            suite.addrs[1].String(),
		StakingCoinWeights: sampleRatioPlan.GetStakingCoinWeights(),
		StartTime:          sampleRatioPlan.GetStartTime(),
		EndTime:            sampleRatioPlan.GetEndTime(),
		EpochRatio:         sampleRatioPlan.EpochRatio,
	}, poolAcc, suite.addrs[1], types.PlanTypePrivate)
	suite.Require().NoError(err)

	globalPlanId = suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(2), globalPlanId)

	plans = suite.keeper.GetPlans(suite.ctx)
	suite.Require().Len(plans, 2)
	suite.Require().Equal(uint64(len(plans)), globalPlanId)

	cacheCtx, _ = suite.ctx.CacheContext()
	nextPlanId = suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(3), nextPlanId)
}
