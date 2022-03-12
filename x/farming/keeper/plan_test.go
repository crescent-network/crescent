package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/farming"
	"github.com/cosmosquad-labs/squad/x/farming/keeper"
	"github.com/cosmosquad-labs/squad/x/farming/types"
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

func (suite *KeeperTestSuite) TestPrivatePlanCreationFee() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))

	// Initially, there is no balances in the farming fee collector.
	params := suite.keeper.GetParams(suite.ctx)
	feeCollectorAcc, _ := sdk.AccAddressFromBech32(params.FarmingFeeCollector)
	suite.Require().True(coinsEq(sdk.Coins{}, suite.app.BankKeeper.GetAllBalances(suite.ctx, feeCollectorAcc)))

	// Create a new private plan.
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	msg := types.NewMsgCreateFixedAmountPlan(
		"plan1", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)))
	_, err := msgServer.CreateFixedAmountPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	// Now the farming fee collector has params.PrivatePlanCreationFee in its balance.
	suite.Require().True(coinsEq(params.PrivatePlanCreationFee, suite.app.BankKeeper.GetAllBalances(suite.ctx, feeCollectorAcc)))

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2023-01-02T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // This terminates the plan above.

	// The plan creator removes the plan, and gets plan creation fee refunded.
	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[4])
	err = suite.keeper.RemovePlan(suite.ctx, suite.addrs[4], 1)
	suite.Require().NoError(err)
	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[4])
	suite.Require().True(coinsEq(params.PrivatePlanCreationFee, balancesAfter.Sub(balancesBefore)))
}

func (suite *KeeperTestSuite) TestMaxNumPrivatePlans() {
	// Adjust the parameter.
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxNumPrivatePlans = 1
	suite.keeper.SetParams(suite.ctx, params)

	// Create a private plan.
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	msg := types.NewMsgCreateFixedAmountPlan(
		"plan1", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)))
	_, err := msgServer.CreateFixedAmountPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2023-01-01T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // The plan is terminated now.

	// Create a new private plan.
	// This is OK even with the MaxNumPrivatePlans param set to 1,
	// because there is no non-terminated private plans.
	msg = types.NewMsgCreateFixedAmountPlan(
		"plan2", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		types.ParseTime("2023-01-01T00:00:00Z"), types.ParseTime("2024-01-01T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)))
	_, err = msgServer.CreateFixedAmountPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	// Creating another private plan will be failed.
	msg = types.NewMsgCreateFixedAmountPlan(
		"plan3", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		types.ParseTime("2024-01-01T00:00:00Z"), types.ParseTime("2025-01-01T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)))
	_, err = msgServer.CreateFixedAmountPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().ErrorIs(err, types.ErrNumPrivatePlansLimit)
}

func (suite *KeeperTestSuite) TestCreateExpiredPlan() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))

	msg := types.NewMsgCreateFixedAmountPlan(
		"plan1", suite.addrs[4], sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		types.ParseTime("2021-01-01T00:00:00Z"), types.ParseTime("2022-01-01T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)))
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	_, err := msgServer.CreateFixedAmountPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().ErrorIs(err, types.ErrInvalidPlanEndTime)

	req := types.AddPlanRequest{
		Name:               "plan2",
		FarmingPoolAddress: suite.addrs[4].String(),
		TerminationAddress: suite.addrs[4].String(),
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)),
		StartTime:          types.ParseTime("2021-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("2022-01-01T00:00:00Z"),
		EpochRatio:         sdk.NewDecWithPrec(3, 1),
	}
	proposal := types.NewPublicPlanProposal("title", "description", []types.AddPlanRequest{req}, nil, nil)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrInvalidPlanEndTime)
}
