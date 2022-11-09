package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
)

//func testAddPlanRequest(name string, farmingPoolAcc, terminationAcc string, weightsStr, epochAmountStr, epochRatioStr string) types.AddPlanRequest {
//	weights, err := sdk.ParseDecCoins(weightsStr)
//	if err != nil {
//		panic(err)
//	}
//	var epochAmount sdk.Coins
//	var epochRatio sdk.Dec
//	if epochAmountStr != "" {
//		var err error
//		epochAmount, err = sdk.ParseCoinsNormalized(epochAmountStr)
//		if err != nil {
//			panic(err)
//		}
//	} else if epochRatioStr != "" {
//		epochRatio = sdk.MustNewDecFromStr(epochRatioStr)
//	}
//	return types.AddPlanRequest{
//		Name:               name,
//		FarmingPoolAddress: farmingPoolAcc,
//		TerminationAddress: terminationAcc,
//		StakingCoinWeights: weights,
//		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
//		EndTime:            types.ParseTime("9999-12-31T00:00:00Z"),
//		EpochAmount:        epochAmount,
//		EpochRatio:         epochRatio,
//	}
//}
//
//func (suite *KeeperTestSuite) TestAddPlanRequest() {
//	plans := suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Empty(plans)
//
//	addr := suite.addrs[4].String()
//
//	req := testAddPlanRequest("plan1", addr, addr, "1denom1", "1000000denom3", "")
//	proposal := types.NewPublicPlanProposal("title", "description", []types.AddPlanRequest{req}, nil, nil)
//	suite.handleProposal(proposal)
//
//	plans = suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Len(plans, 1)
//	plan := plans[0]
//	suite.Require().Equal(uint64(1), plan.GetId())
//	suite.Require().Equal("plan1", plan.GetName())
//
//	// Same plan name is allowed.
//	proposal = types.NewPublicPlanProposal("title", "description", []types.AddPlanRequest{req}, nil, nil)
//	suite.handleProposal(proposal)
//
//	plans = suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Len(plans, 2)
//	plan = plans[1]
//	suite.Require().Equal(uint64(2), plan.GetId())
//	suite.Require().Equal("plan1", plan.GetName())
//}
//
//func testModifyPlanRequest(
//	id uint64, name string, farmingPoolAcc, terminationAcc string,
//	weightsStr, startTimeStr, endTimeStr, epochAmountStr, epochRatioStr string,
//) types.ModifyPlanRequest {
//	weights, err := sdk.ParseDecCoins(weightsStr)
//	if err != nil {
//		panic(err)
//	}
//	var startTime, endTime *time.Time
//	if startTimeStr != "" {
//		t := types.ParseTime(startTimeStr)
//		startTime = &t
//	}
//	if endTimeStr != "" {
//		t := types.ParseTime(endTimeStr)
//		endTime = &t
//	}
//	var epochAmount sdk.Coins
//	var epochRatio sdk.Dec
//	if epochAmountStr != "" {
//		var err error
//		epochAmount, err = sdk.ParseCoinsNormalized(epochAmountStr)
//		if err != nil {
//			panic(err)
//		}
//	} else if epochRatioStr != "" {
//		epochRatio = sdk.MustNewDecFromStr(epochRatioStr)
//	}
//	return types.ModifyPlanRequest{
//		PlanId:             id,
//		Name:               name,
//		FarmingPoolAddress: farmingPoolAcc,
//		TerminationAddress: terminationAcc,
//		StakingCoinWeights: weights,
//		StartTime:          startTime,
//		EndTime:            endTime,
//		EpochAmount:        epochAmount,
//		EpochRatio:         epochRatio,
//	}
//}
//
//func (suite *KeeperTestSuite) TestModifyPlanRequest() {
//	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
//
//	addr := suite.addrs[5].String()
//
//	// Change name, addrs, epoch amount.
//	req := testModifyPlanRequest(1, "new name", addr, addr, "", "", "", "2000000denom3", "")
//	proposal := types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	plans := suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Len(plans, 1)
//	plan := plans[0]
//	suite.Require().Equal(uint64(1), plan.GetId())
//	suite.Require().Equal("new name", plan.GetName())
//	suite.Require().True(decCoinsEq(sdk.NewDecCoins(sdk.NewInt64DecCoin(denom1, 1)), plan.GetStakingCoinWeights()))
//	fixedAmountPlan, ok := plan.(*types.FixedAmountPlan)
//	suite.Require().True(ok)
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 2000000)), fixedAmountPlan.EpochAmount))
//
//	// Change staking coin weights, end time.
//	req = testModifyPlanRequest(1, "", "", "", "1denom2", "", "2021-12-31T00:00:00Z", "", "")
//	proposal = types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	plan, _ = suite.keeper.GetPlan(suite.ctx, 1)
//	// These fields below must not have been modified.
//	suite.Require().Equal("new name", plan.GetName())
//	suite.Require().Equal(addr, plan.GetFarmingPoolAddress().String())
//	suite.Require().Equal(addr, plan.GetTerminationAddress().String())
//	fixedAmountPlan, ok = plan.(*types.FixedAmountPlan)
//	suite.Require().True(ok)
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 2000000)), fixedAmountPlan.EpochAmount))
//	// These fields must have been modified.
//	suite.Require().True(decCoinsEq(sdk.NewDecCoins(sdk.NewInt64DecCoin(denom2, 1)), plan.GetStakingCoinWeights()))
//	suite.Require().Equal(types.ParseTime("2021-12-31T00:00:00Z"), plan.GetEndTime())
//
//	// Change plan type, from FixedAmountPlan to RatioPlan.
//	req = testModifyPlanRequest(1, "", "", "", "", "", "", "", "0.05")
//	proposal = types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	plan, _ = suite.keeper.GetPlan(suite.ctx, 1)
//	ratioPlan, ok := plan.(*types.RatioPlan)
//	suite.Require().True(ok)
//	suite.Require().True(decEq(sdk.MustNewDecFromStr("0.05"), ratioPlan.EpochRatio))
//
//	// Test for private plan cannot be modified.
//	err := plan.SetType(types.PlanTypePrivate)
//	suite.Require().NoError(err)
//	suite.Require().Equal(plan.GetType(), types.PlanTypePrivate)
//	suite.keeper.SetPlan(suite.ctx, plan)
//
//	req = testModifyPlanRequest(1, "", "", "", "", "", "", "", "0.1")
//	proposal = types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	err = proposal.ValidateBasic()
//	suite.Require().NoError(err)
//	err = suite.govHandler(suite.ctx, proposal)
//	suite.Require().ErrorIs(err, types.ErrInvalidPlanType, "plan 2 is not a public plan: invalid plan type")
//}
//
//func (suite *KeeperTestSuite) TestDeletePlanRequest() {
//	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
//
//	proposal := types.NewPublicPlanProposal("title", "description", nil, nil, []types.DeletePlanRequest{{PlanId: 1}})
//	suite.handleProposal(proposal)
//
//	plans := suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Empty(plans)
//
//	// Test for private plan cannot be deleted.
//	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
//	plans = suite.keeper.GetPlans(suite.ctx)
//	suite.Require().Equal(plans[0].GetId(), uint64(2))
//
//	err := plans[0].SetType(types.PlanTypePrivate)
//	suite.Require().NoError(err)
//	suite.Require().Equal(plans[0].GetType(), types.PlanTypePrivate)
//	suite.keeper.SetPlan(suite.ctx, plans[0])
//
//	proposal = types.NewPublicPlanProposal("title", "description", nil, nil, []types.DeletePlanRequest{{PlanId: 2}})
//	err = proposal.ValidateBasic()
//	suite.Require().NoError(err)
//	err = suite.govHandler(suite.ctx, proposal)
//	suite.Require().ErrorIs(err, types.ErrInvalidPlanType, "plan 2 is not a public plan: invalid plan type")
//}
//
//func (suite *KeeperTestSuite) TestWithdrawRewardsAfterPlanDeleted() {
//	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//
//	suite.advanceEpochDays()
//	suite.advanceEpochDays()
//
//	proposal := types.NewPublicPlanProposal("title", "description", nil, nil, []types.DeletePlanRequest{{PlanId: 1}})
//	suite.handleProposal(proposal)
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
//
//	// Additional epochs should not accumulate rewards anymore.
//	suite.advanceEpochDays()
//	suite.advanceEpochDays()
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
//
//	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	suite.Harvest(suite.addrs[0], []string{denom1})
//	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	diff := balancesAfter.Sub(balancesBefore)
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), diff))
//}
//
//func (suite *KeeperTestSuite) TestWithdrawRewardsAfterPlanTerminated() {
//	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-12-01T00:00:00Z"))
//	suite.advanceEpochDays()
//	suite.advanceEpochDays()
//
//	req := testModifyPlanRequest(1, "", "", "", "", "", "2021-11-01T00:00:00Z", "", "")
//	proposal := types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
//
//	// Additional epochs should not accumulate rewards anymore.
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-12-02T00:00:00Z"))
//	farming.EndBlocker(suite.ctx, suite.keeper) // The plan should be terminated here.
//	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-12-03T00:00:00Z"))
//	farming.EndBlocker(suite.ctx, suite.keeper)
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
//
//	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	suite.Harvest(suite.addrs[0], []string{denom1})
//	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
//	diff := balancesAfter.Sub(balancesBefore)
//
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), diff))
//}
//
//func (suite *KeeperTestSuite) TestAccumulatedRewardsAfterPlanModification() {
//	farmingPool := suite.AddTestAddrs(1, sdk.NewCoins(
//		sdk.NewInt64Coin(denom2, 10000000),
//		sdk.NewInt64Coin(denom3, 10000000),
//	))[0]
//
//	suite.CreateRatioPlan(farmingPool, map[string]string{denom1: "1"}, "0.1")
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//	suite.advanceEpochDays()
//	suite.advanceEpochDays() // The farmer has 1000000denom2,1000000denom3 as rewards.
//
//	req := testModifyPlanRequest(1, "", "", "", "", "", "", "1000000denom3", "")
//	proposal := types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	suite.advanceEpochDays() // This adds 1000000denom3 as rewards to the farmer.
//
//	suite.Require().True(coinsEq(
//		sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000000), sdk.NewInt64Coin(denom3, 2000000)),
//		suite.AllRewards(suite.addrs[0]),
//	))
//
//	req = testModifyPlanRequest(1, "", "", "", "", "", "", "", "0.5")
//	proposal = types.NewPublicPlanProposal("title", "description", nil, []types.ModifyPlanRequest{req}, nil)
//	suite.handleProposal(proposal)
//
//	suite.advanceEpochDays() // This adds 4500000denom2,4000000denom3 as rewards to the farmer.
//
//	suite.Require().True(coinsEq(
//		sdk.NewCoins(sdk.NewInt64Coin(denom2, 5500000), sdk.NewInt64Coin(denom3, 6000000)),
//		suite.AllRewards(suite.addrs[0]),
//	))
//}
//
//func (suite *KeeperTestSuite) TestValidateAddPublicPlanProposal() {
//	for _, tc := range []struct {
//		name        string
//		addReqs     []types.AddPlanRequest
//		expectedErr error
//	}{
//		{
//			"happy case",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//				sdk.Dec{},
//			)},
//			nil,
//		},
//		{
//			"request case #1",
//			nil,
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
//		},
//		{
//			"name case #1",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//				sdk.ZeroDec(),
//			)},
//			sdkerrors.Wrapf(types.ErrInvalidPlanName, "plan name cannot be longer than max length of %d", types.MaxNameLength),
//		},
//		{
//			"staking coin weights case #1",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//				sdk.ZeroDec(),
//			)},
//			sdkerrors.Wrap(types.ErrInvalidStakingCoinWeights, "staking coin weights must not be empty"),
//		},
//		{
//			"staking coin weights case #2",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.DecCoin{
//						Denom:  "pool1",
//						Amount: sdk.MustNewDecFromStr("0.1"),
//					},
//				),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//				sdk.ZeroDec(),
//			)},
//			sdkerrors.Wrap(types.ErrInvalidStakingCoinWeights, "total weight must be 1"),
//		},
//		{
//			"start time & end time case #1",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("2021-08-13T00:00:00Z"),
//				types.ParseTime("2021-08-06T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//				sdk.ZeroDec(),
//			)},
//			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
//				"end time %s must be greater than start time %s",
//				types.ParseTime("2021-08-06T00:00:00Z").Format(time.RFC3339), types.ParseTime("2021-08-13T00:00:00Z").Format(time.RFC3339)),
//		},
//		{
//			"epoch amount & epoch ratio case #1",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(sdk.NewInt64Coin(denom3, 1)),
//				sdk.NewDec(1),
//			)},
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "exactly one of epoch amount or epoch ratio must be provided"),
//		},
//		{
//			"epoch amount & epoch ratio case #2",
//			[]types.AddPlanRequest{types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("2021-08-01T00:00:00Z"),
//				types.ParseTime("2021-08-30T00:00:00Z"),
//				sdk.NewCoins(),
//				sdk.ZeroDec(),
//			)},
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "exactly one of epoch amount or epoch ratio must be provided"),
//		},
//	} {
//		suite.Run(tc.name, func() {
//			proposal := &types.PublicPlanProposal{
//				Title:              "testTitle",
//				Description:        "testDescription",
//				AddPlanRequests:    tc.addReqs,
//				ModifyPlanRequests: nil,
//				DeletePlanRequests: nil,
//			}
//
//			err := proposal.ValidateBasic()
//			if tc.expectedErr == nil {
//				suite.NoError(err)
//
//				suite.handleProposal(proposal)
//
//				_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
//				suite.Require().Equal(true, found)
//			} else {
//				suite.EqualError(err, tc.expectedErr.Error())
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestValidateModifyPublicPlanProposal() {
//	// create a ratio public plan
//	addRequests := []types.AddPlanRequest{
//		types.NewAddPlanRequest(
//			"testPlan",
//			suite.addrs[0].String(),
//			suite.addrs[0].String(),
//			sdk.NewDecCoins(
//				sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
//				sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
//			),
//			types.ParseTime("2021-08-01T00:00:00Z"),
//			types.ParseTime("2021-08-30T00:00:00Z"),
//			nil,
//			sdk.NewDecWithPrec(10, 2), // 10%
//		),
//	}
//
//	err := keeper.HandlePublicPlanProposal(
//		suite.ctx,
//		suite.keeper,
//		types.NewPublicPlanProposal("testTitle", "testDescription", addRequests, nil, nil),
//	)
//	suite.Require().NoError(err)
//
//	plan, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(true, found)
//
//	for _, tc := range []struct {
//		name        string
//		modifyReqs  []types.ModifyPlanRequest
//		expectedErr error
//	}{
//		{
//			"happy case #1 - decrease epoch ratio to 5%",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				sdk.NewDecWithPrec(5, 2),
//			)},
//			nil,
//		},
//		{
//			"request case #1",
//			nil,
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
//		},
//		{
//			"plan id case #1",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				uint64(0),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				plan.(*types.RatioPlan).EpochRatio,
//			)},
//			sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", uint64(0)),
//		},
//		{
//			"name case #1",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				plan.GetId(),
//				"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM"+
//					"OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM", // max length of name
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				plan.(*types.RatioPlan).EpochRatio,
//			)},
//			sdkerrors.Wrapf(types.ErrInvalidPlanName, "plan name cannot be longer than max length of %d", types.MaxNameLength),
//		},
//		{
//			"staking coin weights case #1",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				sdk.NewDecCoins(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				plan.(*types.RatioPlan).EpochRatio,
//			)},
//			sdkerrors.Wrap(types.ErrInvalidStakingCoinWeights, "staking coin weights must not be empty"),
//		},
//		{
//			"staking coin weights case #2",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				sdk.NewDecCoins(
//					sdk.DecCoin{
//						Denom:  "pool1",
//						Amount: sdk.MustNewDecFromStr("0.1"),
//					},
//				),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				plan.(*types.RatioPlan).EpochRatio,
//			)},
//			sdkerrors.Wrap(types.ErrInvalidStakingCoinWeights, "total weight must be 1"),
//		},
//		{
//			"start time & end time case #1",
//			[]types.ModifyPlanRequest{types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				types.ParseTime("2021-08-13T00:00:00Z"),
//				types.ParseTime("2021-08-06T00:00:00Z"),
//				nil,
//				plan.(*types.RatioPlan).EpochRatio,
//			)},
//			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
//				"end time %s must be greater than start time %s",
//				types.ParseTime("2021-08-06T00:00:00Z"), types.ParseTime("2021-08-13T00:00:00Z")),
//		},
//		{
//			"epoch amount & epoch ratio case #1",
//			[]types.ModifyPlanRequest{
//				types.NewModifyPlanRequest(
//					plan.GetId(),
//					plan.GetName(),
//					plan.GetFarmingPoolAddress().String(),
//					plan.GetTerminationAddress().String(),
//					plan.GetStakingCoinWeights(),
//					plan.GetStartTime(),
//					plan.GetEndTime(),
//					sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)),
//					plan.(*types.RatioPlan).EpochRatio,
//				)},
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "at most one of epoch amount or epoch ratio must be provided"),
//		},
//		{
//			"epoch amount & epoch ratio case #2",
//			[]types.ModifyPlanRequest{
//				types.NewModifyPlanRequest(
//					plan.GetId(),
//					plan.GetName(),
//					plan.GetFarmingPoolAddress().String(),
//					plan.GetTerminationAddress().String(),
//					plan.GetStakingCoinWeights(),
//					plan.GetStartTime(),
//					plan.GetEndTime(),
//					sdk.Coins{sdk.NewInt64Coin("stake", 0)},
//					sdk.ZeroDec(),
//				)},
//			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid epoch amount: coin 0stake amount is not positive"),
//		},
//	} {
//		suite.Run(tc.name, func() {
//			proposal := &types.PublicPlanProposal{
//				Title:              "testTitle",
//				Description:        "testDescription",
//				AddPlanRequests:    nil,
//				ModifyPlanRequests: tc.modifyReqs,
//				DeletePlanRequests: nil,
//			}
//
//			err := proposal.ValidateBasic()
//			if tc.expectedErr == nil {
//				suite.NoError(err)
//
//				suite.handleProposal(proposal)
//
//				_, found := suite.keeper.GetPlan(suite.ctx, tc.modifyReqs[0].GetPlanId())
//				suite.Require().Equal(true, found)
//			} else {
//				suite.EqualError(err, tc.expectedErr.Error())
//			}
//		})
//	}
//}

//func (suite *KeeperTestSuite) TestValidateDeletePublicPlanProposal() {
//	// create a ratio public plan
//	addRequests := []types.AddPlanRequest{types.NewAddPlanRequest(
//		"testPlan",
//		suite.addrs[0].String(),
//		suite.addrs[0].String(),
//		sdk.NewDecCoins(
//			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
//			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
//		),
//		types.ParseTime("2021-08-01T00:00:00Z"),
//		types.ParseTime("2021-08-30T00:00:00Z"),
//		nil,
//		sdk.NewDecWithPrec(10, 2), // 10%
//	)}
//
//	suite.handleProposal(types.NewPublicPlanProposal("testTitle", "testDescription", addRequests, nil, nil))
//
//	// should exist
//	_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(true, found)
//
//	// delete the proposal
//	deleteRequests := []types.DeletePlanRequest{types.NewDeletePlanRequest(uint64(1))}
//
//	suite.handleProposal(types.NewPublicPlanProposal("testTitle", "testDescription", nil, nil, deleteRequests))
//
//	// shouldn't exist
//	_, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(false, found)
//}
//
//func (suite *KeeperTestSuite) TestUpdatePlanType() {
//	// create a ratio public plan
//	suite.handleProposal(
//		types.NewPublicPlanProposal("testTitle", "testDescription", []types.AddPlanRequest{
//			types.NewAddPlanRequest(
//				"testPlan",
//				suite.addrs[0].String(),
//				suite.addrs[0].String(),
//				sdk.NewDecCoins(
//					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//				),
//				types.ParseTime("0001-01-01T00:00:00Z"),
//				types.ParseTime("9999-12-31T00:00:00Z"),
//				sdk.NewCoins(),
//				sdk.NewDecWithPrec(10, 2),
//			),
//		}, nil, nil),
//	)
//
//	plan, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(true, found)
//	suite.Require().Equal(plan.(*types.RatioPlan).EpochRatio, sdk.NewDecWithPrec(10, 2))
//
//	// update the ratio plan type to fixed amount plan type
//	suite.handleProposal(
//		types.NewPublicPlanProposal("testTitle", "testDescription", nil, []types.ModifyPlanRequest{
//			types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)),
//				sdk.ZeroDec(),
//			),
//		}, nil),
//	)
//
//	plan, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(true, found)
//	suite.Require().Equal(plan.(*types.FixedAmountPlan).EpochAmount, sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)))
//
//	// update back to ratio plan with different epoch ratio
//	suite.handleProposal(
//		types.NewPublicPlanProposal("testTitle", "testDescription", nil, []types.ModifyPlanRequest{
//			types.NewModifyPlanRequest(
//				plan.GetId(),
//				plan.GetName(),
//				plan.GetFarmingPoolAddress().String(),
//				plan.GetTerminationAddress().String(),
//				plan.GetStakingCoinWeights(),
//				plan.GetStartTime(),
//				plan.GetEndTime(),
//				nil,
//				sdk.NewDecWithPrec(7, 2), // 7%
//			),
//		}, nil),
//	)
//
//	plan, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
//	suite.Require().Equal(true, found)
//	suite.Require().Equal(plan.(*types.RatioPlan).EpochRatio, sdk.NewDecWithPrec(7, 2))
//}
//
//func (suite *KeeperTestSuite) TestDeletePublicPlan() {
//	for _, tc := range []struct {
//		name             string
//		farmingPoolAddr  sdk.AccAddress
//		terminationAddr  sdk.AccAddress
//		expectedBalances sdk.Coins
//	}{
//		{
//			"farming pool address and termination address are equal",
//			suite.addrs[0],
//			suite.addrs[0],
//			initialBalances,
//		},
//		{
//			"farming pool address and termination address are not equal",
//			suite.addrs[1],
//			suite.addrs[2],
//			sdk.Coins{},
//		},
//	} {
//		suite.Run(tc.name, func() {
//			cacheCtx, _ := suite.ctx.CacheContext()
//
//			// create a public plan
//			err := keeper.HandlePublicPlanProposal(
//				cacheCtx,
//				suite.keeper,
//				types.NewPublicPlanProposal("testTitle", "testDescription", []types.AddPlanRequest{
//					types.NewAddPlanRequest(
//						"testPlan",
//						tc.farmingPoolAddr.String(),
//						tc.terminationAddr.String(),
//						sdk.NewDecCoins(
//							sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
//							sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
//						),
//						types.ParseTime("0001-01-01T00:00:00Z"),
//						types.ParseTime("9999-12-31T00:00:00Z"),
//						sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
//						sdk.ZeroDec(),
//					),
//				}, nil, nil),
//			)
//			suite.Require().NoError(err)
//
//			plans := suite.keeper.GetPlans(cacheCtx)
//
//			// delete the plan
//			err = keeper.HandlePublicPlanProposal(
//				cacheCtx,
//				suite.keeper,
//				types.NewPublicPlanProposal("testTitle", "testDescription", nil, nil, []types.DeletePlanRequest{
//					types.NewDeletePlanRequest(plans[0].GetId()),
//				}),
//			)
//			suite.Require().NoError(err)
//
//			// the plan should be successfully removed and coins meet the expected balances
//			_, found := suite.keeper.GetPlan(cacheCtx, plans[0].GetId())
//			suite.Require().Equal(false, found)
//			suite.Require().Equal(tc.expectedBalances, suite.app.BankKeeper.GetAllBalances(cacheCtx, tc.farmingPoolAddr))
//
//			isPlanTerminatedEventType := false
//			for _, e := range cacheCtx.EventManager().ABCIEvents() {
//				if e.Type == types.EventTypePlanTerminated {
//					suite.Require().Equal(fmt.Sprint(plans[0].GetId()), string(e.Attributes[0].Value))
//					isPlanTerminatedEventType = true
//					break
//				}
//			}
//			suite.Require().True(isPlanTerminatedEventType, "plan_terminated events should be emitted")
//		})
//	}
//}
