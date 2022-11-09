package farming_test

import (
	_ "github.com/stretchr/testify/suite"
)

//func (suite *ModuleTestSuite) TestMsgCreateFixedAmountPlan() {
//	msg := types.NewMsgCreateFixedAmountPlan(
//		"handlerTestPlan1",
//		suite.addrs[0],
//		sdk.NewDecCoins(
//			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
//			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
//		),
//		types.ParseTime("2021-08-02T00:00:00Z"),
//		types.ParseTime("2021-08-10T00:00:00Z"),
//		sdk.NewCoins(sdk.NewInt64Coin(denom3, 10_000_000)),
//	)
//
//	handler := farming.NewHandler(suite.keeper)
//	_, err := handler(suite.ctx, msg)
//	suite.Require().NoError(err)
//
//	plan, found := suite.keeper.GetPlan(suite.ctx, 1)
//	suite.Require().Equal(true, found)
//
//	suite.Require().Equal(msg.Name, plan.GetName())
//	suite.Require().Equal(msg.Creator, plan.GetTerminationAddress().String())
//	suite.Require().Equal(msg.StakingCoinWeights, plan.GetStakingCoinWeights())
//	suite.Require().Equal(types.PrivatePlanFarmingPoolAcc(msg.Name, 1), plan.GetFarmingPoolAddress())
//	suite.Require().Equal(types.ParseTime("2021-08-02T00:00:00Z"), plan.GetStartTime())
//	suite.Require().Equal(types.ParseTime("2021-08-10T00:00:00Z"), plan.GetEndTime())
//	suite.Require().Equal(msg.EpochAmount, plan.(*types.FixedAmountPlan).EpochAmount)
//}
//
//func (suite *ModuleTestSuite) TestMsgCreateRatioPlan() {
//	msg := types.NewMsgCreateRatioPlan(
//		"handlerTestPlan2",
//		suite.addrs[0],
//		sdk.NewDecCoins(
//			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
//			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
//		),
//		types.ParseTime("2021-08-02T00:00:00Z"),
//		types.ParseTime("2021-08-10T00:00:00Z"),
//		sdk.NewDecWithPrec(4, 2), // 4%,
//	)
//
//	handler := farming.NewHandler(suite.keeper)
//	_, err := handler(suite.ctx, msg)
//	suite.Require().NoError(err)
//
//	plan, found := suite.keeper.GetPlan(suite.ctx, 1)
//	suite.Require().Equal(true, found)
//
//	suite.Require().Equal(msg.Name, plan.GetName())
//	suite.Require().Equal(msg.Creator, plan.GetTerminationAddress().String())
//	suite.Require().Equal(msg.StakingCoinWeights, plan.GetStakingCoinWeights())
//	suite.Require().Equal(types.PrivatePlanFarmingPoolAcc(msg.Name, 1), plan.GetFarmingPoolAddress())
//	suite.Require().Equal(types.ParseTime("2021-08-02T00:00:00Z"), plan.GetStartTime())
//	suite.Require().Equal(types.ParseTime("2021-08-10T00:00:00Z"), plan.GetEndTime())
//	suite.Require().Equal(msg.EpochRatio, plan.(*types.RatioPlan).EpochRatio)
//}
//
//func (suite *ModuleTestSuite) TestMsgStake() {
//	msg := types.NewMsgStake(
//		suite.addrs[0],
//		sdk.NewCoins(sdk.NewInt64Coin(denom1, 10_000_000)),
//	)
//
//	handler := farming.NewHandler(suite.keeper)
//	_, err := handler(suite.ctx, msg)
//	suite.Require().NoError(err)
//
//	queuedCoins := suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
//	suite.Require().Equal(msg.StakingCoins, queuedCoins)
//}
//
//func (suite *ModuleTestSuite) TestMsgUnstake() {
//	stakeCoin := sdk.NewInt64Coin(denom1, 10_000_000)
//	suite.Stake(suite.addrs[0], sdk.NewCoins(stakeCoin))
//
//	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//
//	unstakeCoin := sdk.NewInt64Coin(denom1, 5_000_000)
//	msg := types.NewMsgUnstake(suite.addrs[0], sdk.NewCoins(unstakeCoin))
//
//	handler := farming.NewHandler(suite.keeper)
//	_, err := handler(suite.ctx, msg)
//	suite.Require().NoError(err)
//
//	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	suite.Require().True(coinsEq(balancesBefore.Add(unstakeCoin), balancesAfter))
//}
//
//func (suite *ModuleTestSuite) TestMsgHarvest() {
//	for _, plan := range suite.samplePlans {
//		suite.keeper.SetPlan(suite.ctx, plan)
//	}
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom2, 10_000_000)))
//	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(types.Day))
//	farming.EndBlocker(suite.ctx, suite.keeper)
//
//	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-05T00:00:00Z"))
//	err := suite.keeper.AllocateRewards(suite.ctx)
//	suite.Require().NoError(err)
//
//	rewards := suite.Rewards(suite.addrs[0])
//
//	msg := types.NewMsgHarvest(suite.addrs[0], []string{denom2})
//
//	handler := farming.NewHandler(suite.keeper)
//	_, err = handler(suite.ctx, msg)
//	suite.Require().NoError(err)
//
//	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	suite.Require().True(coinsEq(balancesBefore.Add(rewards...), balancesAfter))
//	suite.Require().True(suite.app.BankKeeper.GetAllBalances(suite.ctx, types.RewardsReserveAcc).IsZero())
//	suite.Require().True(suite.Rewards(suite.addrs[0]).IsZero())
//}
//
//func (suite *ModuleTestSuite) TestMsgRemovePlan() {
//	handler := farming.NewHandler(suite.keeper)
//
//	// Create a private plan.
//	_, err := handler(suite.ctx, types.NewMsgCreateRatioPlan(
//		"plan1", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
//		types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
//		sdk.MustNewDecFromStr("0.1")))
//	suite.Require().NoError(err)
//
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))
//	farming.EndBlocker(suite.ctx, suite.keeper) // The plan above is not terminated yet.
//
//	// Plan is not terminated yet.
//	_, err = handler(suite.ctx, types.NewMsgRemovePlan(suite.addrs[4], 1))
//	suite.Require().EqualError(err, "plan 1 is not terminated yet: invalid request")
//
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2023-01-01T00:00:00Z"))
//	farming.EndBlocker(suite.ctx, suite.keeper) // The plan is terminated now.
//
//	// Wrong creator.
//	_, err = handler(suite.ctx, types.NewMsgRemovePlan(suite.addrs[0], 1))
//	suite.Require().EqualError(err, "only the plan creator can remove the plan: unauthorized")
//
//	// Happy case.
//	_, err = handler(suite.ctx, types.NewMsgRemovePlan(suite.addrs[4], 1))
//	suite.Require().NoError(err)
//}
