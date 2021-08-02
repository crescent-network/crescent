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
	basePlan := types.NewBasePlan(1, name, 1, farmingPoolAddr.String(), terminationAddr.String(), coinWeights, startTime, endTime)
	fixedPlan := types.NewFixedAmountPlan(basePlan, sdk.NewCoins(sdk.NewCoin("testFarmCoinDenom", sdk.NewInt(1000000))))
	suite.keeper.SetPlan(suite.ctx, fixedPlan)

	planGet, found := suite.keeper.GetPlan(suite.ctx, 1)
	suite.Require().True(found)
	suite.Require().Equal(fixedPlan, planGet)

	plans := suite.keeper.GetAllPlans(suite.ctx)
	suite.Require().Len(plans, 1)
	suite.Require().Equal(fixedPlan, plans[0])

	// TODO: tmp test codes for testing functionality, need to separated
	_, err := suite.keeper.Stake(suite.ctx, farmerAddr, stakingCoins)
	suite.Require().NoError(err)

	stakings := suite.keeper.GetAllStakings(suite.ctx)
	fmt.Println(stakings)
	stakingByFarmer, found := suite.keeper.GetStakingByFarmer(suite.ctx, farmerAddr)
	stakingsByDenom := suite.keeper.GetStakingsByStakingCoinDenom(suite.ctx, sdk.DefaultBondDenom)

	suite.Require().True(found)
	suite.Require().Equal(stakings[0], stakingByFarmer)
	suite.Require().Equal(stakings, stakingsByDenom)

	suite.keeper.SetReward(suite.ctx, sdk.DefaultBondDenom, farmerAddr, stakingCoins)

	//rewards := suite.keeper.GetAllRewards(ctx)
	//rewardsByFarmer := suite.keeper.GetRewardsByFarmer(ctx, farmerAddr)
	//rewardsByDenom := suite.keeper.GetRewardsByStakingCoinDenom(ctx, sdk.DefaultBondDenom)
	//
	//suite.Require().Equal(rewards, rewardsByFarmer)
	//suite.Require().Equal(rewards, rewardsByDenom)
}
