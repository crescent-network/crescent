package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"
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

	req := types.NewAddPlanRequest(
		"plan2", suite.addrs[4].String(), suite.addrs[4].String(), parseDecCoins("1denom1"),
		types.ParseTime("2021-01-01T00:00:00Z"), types.ParseTime("2022-01-01T00:00:00Z"),
		nil, parseDec("0.3"))
	proposal := types.NewPublicPlanProposal("title", "description", []types.AddPlanRequest{req}, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrInvalidPlanEndTime)
}

func (suite *KeeperTestSuite) TestPrivatePlanNumMaxDenoms() {
	numDenoms := types.PrivatePlanMaxNumDenoms + 1 // Invalid number of denoms

	weights := make(sdk.DecCoins, numDenoms)
	totalWeight := sdk.ZeroDec()
	for i := range weights {
		var weight sdk.Dec
		if i < numDenoms {
			weight = sdk.OneDec().QuoTruncate(sdk.NewDec(int64(numDenoms)))
		} else {
			weight = sdk.OneDec().Sub(totalWeight)
		}
		weights[i] = sdk.NewDecCoinFromDec(fmt.Sprintf("stake%d", i), weight)
		totalWeight = totalWeight.Add(weight)
	}
	suite.addDenomsFromDecCoins(weights)
	_, err := suite.createPrivateFixedAmountPlan(
		suite.addrs[0], weights,
		sampleStartTime, sampleEndTime, parseCoins("1000000denom3"))
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)

	epochAmt := make(sdk.Coins, numDenoms)
	for i := range epochAmt {
		epochAmt[i] = sdk.NewInt64Coin(fmt.Sprintf("reward%d", i), 1000000)
	}
	suite.addDenomsFromCoins(epochAmt)
	_, err = suite.createPrivateFixedAmountPlan(
		suite.addrs[0], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, epochAmt)
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)

	_, err = suite.createPrivateRatioPlan(
		suite.addrs[0], weights,
		sampleStartTime, sampleEndTime, parseDec("0.1"))
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)
}

func (suite *KeeperTestSuite) TestPublicPlanMaxNumDenoms() {
	numDenoms := types.PublicPlanMaxNumDenoms + 1 // Invalid number of denoms

	weights := make(sdk.DecCoins, numDenoms)
	totalWeight := sdk.ZeroDec()
	for i := range weights {
		var weight sdk.Dec
		if i < numDenoms {
			weight = sdk.OneDec().QuoTruncate(sdk.NewDec(int64(numDenoms)))
		} else {
			weight = sdk.OneDec().Sub(totalWeight)
		}
		weights[i] = sdk.NewDecCoinFromDec(fmt.Sprintf("stake%d", i), weight)
		totalWeight = totalWeight.Add(weight)
	}
	suite.addDenomsFromDecCoins(weights)
	_, err := suite.createPublicFixedAmountPlan(
		suite.addrs[0], suite.addrs[0], weights,
		sampleStartTime, sampleEndTime, parseCoins("1000000denom3"))
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)

	epochAmt := make(sdk.Coins, numDenoms)
	for i := range epochAmt {
		epochAmt[i] = sdk.NewInt64Coin(fmt.Sprintf("reward%d", i), 1000000)
	}
	suite.addDenomsFromCoins(epochAmt)
	_, err = suite.createPublicFixedAmountPlan(
		suite.addrs[0], suite.addrs[0], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, epochAmt)
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)

	_, err = suite.createPublicRatioPlan(
		suite.addrs[0], suite.addrs[0], weights,
		sampleStartTime, sampleEndTime, parseDec("0.1"))
	suite.Require().ErrorIs(err, types.ErrNumMaxDenomsLimit)
}

func (suite *KeeperTestSuite) TestCreatePlanSupply() {
	weight := parseDecCoins("1nosupply1")
	_, err := suite.createPublicFixedAmountPlan(
		suite.addrs[0], suite.addrs[0], weight,
		sampleStartTime, sampleEndTime, parseCoins("1000000denom3"))
	suite.Require().ErrorIs(err, types.ErrInvalidStakingCoinWeights)

	_, err = suite.createPublicRatioPlan(
		suite.addrs[0], suite.addrs[0], weight,
		sampleStartTime, sampleEndTime, parseDec("0.1"))
	suite.Require().ErrorIs(err, types.ErrInvalidStakingCoinWeights)

	epochAmt := sdk.NewCoins(
		sdk.NewInt64Coin("nosupply", 10000),
		sdk.NewInt64Coin("stake", 10000))

	suite.addDenomsFromDecCoins(parseDecCoins("1denom1"))
	_, err = suite.createPublicFixedAmountPlan(
		suite.addrs[0], suite.addrs[0], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, epochAmt)
	suite.Require().ErrorIs(err, types.ErrInvalidEpochAmount)
}

func (suite *KeeperTestSuite) TestRatioPlanDefaultDisabled() {
	keeper.EnableRatioPlan = false
	defer func() {
		keeper.EnableRatioPlan = true // Rollback the change
	}()

	// Creating a ratio plan through the msg server will fail.
	msg := types.NewMsgCreateRatioPlan(
		"plan1", suite.addrs[0], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, parseDec("0.01"))
	_, err := suite.msgServer.CreateRatioPlan(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().ErrorIs(err, types.ErrRatioPlanDisabled)

	// Adding a ratio plan through public plan proposal will fail.
	addReq := types.NewAddPlanRequest(
		"plan1", suite.addrs[1].String(), suite.addrs[1].String(),
		parseDecCoins("1denom1"), sampleStartTime, sampleEndTime, nil, parseDec("0.01"))
	proposal := types.NewPublicPlanProposal(
		"title", "description", []types.AddPlanRequest{addReq}, nil, nil)
	suite.Require().NoError(proposal.ValidateBasic())
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrRatioPlanDisabled)

	// Modifying a plan type to ratio plan through proposal will fail, too.
	plan, err := suite.createPublicFixedAmountPlan(
		suite.addrs[2], suite.addrs[2], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, parseCoins("1000000stake"))
	suite.Require().NoError(err)
	modifyReq := types.NewModifyPlanRequest(
		plan.GetId(), "", "", "", nil, plan.GetStartTime(), plan.GetEndTime(), nil, parseDec("0.01"))
	proposal = types.NewPublicPlanProposal(
		"title", "description", nil, []types.ModifyPlanRequest{modifyReq}, nil)
	suite.Require().NoError(proposal.ValidateBasic())
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrRatioPlanDisabled)
}
