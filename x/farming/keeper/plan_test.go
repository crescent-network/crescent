package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/types"
)

func (suite *KeeperTestSuite) TestGetSetNewPlan() {
	name := ""
	farmingPoolAddr := sdk.AccAddress("farmingPoolAddr")
	terminationAddr := sdk.AccAddress("terminationAddr")

	stakingCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)

	addrs := app.AddTestAddrs(suite.app, suite.ctx, 2, sdk.NewInt(2000000))
	farmerAddr := addrs[0]

	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)
	basePlan := types.NewBasePlan(1, name, 1, farmingPoolAddr.String(), terminationAddr.String(), coinWeights, startTime, endTime, false)
	fixedPlan := types.NewFixedAmountPlan(basePlan, sdk.NewCoins(sdk.NewCoin("testFarmCoinDenom", sdk.NewInt(1000000))))
	suite.keeper.SetPlan(suite.ctx, fixedPlan)

	planGet, found := suite.keeper.GetPlan(suite.ctx, 1)
	suite.Require().True(found)
	suite.Require().Equal(fixedPlan, planGet)

	plans := suite.keeper.GetAllPlans(suite.ctx)
	suite.Require().Len(plans, 1)
	suite.Require().Equal(fixedPlan, plans[0])

	_, err := suite.keeper.Stake(suite.ctx, farmerAddr, stakingCoins)
	suite.Require().NoError(err)

	stakings := suite.keeper.GetAllStakings(suite.ctx)
	fmt.Println(stakings)
	stakingByFarmer, found := suite.keeper.GetStakingByFarmer(suite.ctx, farmerAddr)
	stakingsByDenom := suite.keeper.GetStakingsByStakingCoinDenom(suite.ctx, sdk.DefaultBondDenom)

	suite.Require().True(found)
	suite.Require().Equal(stakings[0], stakingByFarmer)
	suite.Require().Equal(stakings, stakingsByDenom)
}

func (suite *KeeperTestSuite) TestGlobalPlanId() {
	globalPlanId := suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(0), globalPlanId)

	cacheCtx, _ := suite.ctx.CacheContext()
	nextPlanId := suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(1), nextPlanId)

	sampleFixedPlan := suite.sampleFixedAmtPlans[0].(*types.FixedAmountPlan)
	_, err := suite.keeper.CreateFixedAmountPlan(suite.ctx, &types.MsgCreateFixedAmountPlan{
		Name:               sampleFixedPlan.Name,
		FarmingPoolAddress: sampleFixedPlan.FarmingPoolAddress,
		StakingCoinWeights: sampleFixedPlan.GetStakingCoinWeights(),
		StartTime:          sampleFixedPlan.GetStartTime(),
		EndTime:            sampleFixedPlan.GetEndTime(),
		EpochAmount:        sampleFixedPlan.EpochAmount,
	}, types.PlanTypePublic)
	suite.Require().NoError(err)

	globalPlanId = suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(1), globalPlanId)

	plans := suite.keeper.GetAllPlans(suite.ctx)
	suite.Require().Len(plans, 1)
	suite.Require().Equal(uint64(len(plans)), globalPlanId)

	cacheCtx, _ = suite.ctx.CacheContext()
	nextPlanId = suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(2), nextPlanId)

	sampleRatioPlan := suite.sampleRatioPlans[0].(*types.RatioPlan)
	_, err = suite.keeper.CreateRatioPlan(suite.ctx, &types.MsgCreateRatioPlan{
		Name:               sampleRatioPlan.Name,
		FarmingPoolAddress: sampleRatioPlan.FarmingPoolAddress,
		StakingCoinWeights: sampleRatioPlan.GetStakingCoinWeights(),
		StartTime:          sampleRatioPlan.GetStartTime(),
		EndTime:            sampleRatioPlan.GetEndTime(),
		EpochRatio:         sampleRatioPlan.EpochRatio,
	}, types.PlanTypePrivate)
	suite.Require().NoError(err)

	globalPlanId = suite.keeper.GetGlobalPlanId(suite.ctx)
	suite.Require().Equal(uint64(2), globalPlanId)

	plans = suite.keeper.GetAllPlans(suite.ctx)
	suite.Require().Len(plans, 2)
	suite.Require().Equal(uint64(len(plans)), globalPlanId)

	cacheCtx, _ = suite.ctx.CacheContext()
	nextPlanId = suite.keeper.GetNextPlanIdWithUpdate(cacheCtx)
	suite.Require().Equal(uint64(3), nextPlanId)
}
