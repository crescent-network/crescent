package farming_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/farming"
	"github.com/crescent-network/crescent/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *ModuleTestSuite) TestMsgCreateFixedAmountPlan() {
	msg := types.NewMsgCreateFixedAmountPlan(
		"handlerTestPlan1",
		suite.addrs[0],
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
		),
		types.ParseTime("2021-08-02T00:00:00Z"),
		types.ParseTime("2021-08-10T00:00:00Z"),
		sdk.NewCoins(sdk.NewInt64Coin(denom3, 10_000_000)),
	)

	handler := farming.NewHandler(suite.keeper)
	_, err := handler(suite.ctx, msg)
	suite.Require().NoError(err)

	plan, found := suite.keeper.GetPlan(suite.ctx, 1)
	suite.Require().Equal(true, found)

	suite.Require().Equal(msg.Name, plan.GetName())
	suite.Require().Equal(msg.Creator, plan.GetTerminationAddress().String())
	suite.Require().Equal(msg.StakingCoinWeights, plan.GetStakingCoinWeights())
	suite.Require().Equal(types.PrivatePlanFarmingPoolAcc(msg.Name, 1), plan.GetFarmingPoolAddress())
	suite.Require().Equal(types.ParseTime("2021-08-02T00:00:00Z"), plan.GetStartTime())
	suite.Require().Equal(types.ParseTime("2021-08-10T00:00:00Z"), plan.GetEndTime())
	suite.Require().Equal(msg.EpochAmount, plan.(*types.FixedAmountPlan).EpochAmount)
}

func (suite *ModuleTestSuite) TestMsgCreateRatioPlan() {
	msg := types.NewMsgCreateRatioPlan(
		"handlerTestPlan2",
		suite.addrs[0],
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
		),
		types.ParseTime("2021-08-02T00:00:00Z"),
		types.ParseTime("2021-08-10T00:00:00Z"),
		sdk.NewDecWithPrec(4, 2), // 4%,
	)

	handler := farming.NewHandler(suite.keeper)
	_, err := handler(suite.ctx, msg)
	suite.Require().NoError(err)

	plan, found := suite.keeper.GetPlan(suite.ctx, 1)
	suite.Require().Equal(true, found)

	suite.Require().Equal(msg.Name, plan.GetName())
	suite.Require().Equal(msg.Creator, plan.GetTerminationAddress().String())
	suite.Require().Equal(msg.StakingCoinWeights, plan.GetStakingCoinWeights())
	suite.Require().Equal(types.PrivatePlanFarmingPoolAcc(msg.Name, 1), plan.GetFarmingPoolAddress())
	suite.Require().Equal(types.ParseTime("2021-08-02T00:00:00Z"), plan.GetStartTime())
	suite.Require().Equal(types.ParseTime("2021-08-10T00:00:00Z"), plan.GetEndTime())
	suite.Require().Equal(msg.EpochRatio, plan.(*types.RatioPlan).EpochRatio)
}

func (suite *ModuleTestSuite) TestMsgStake() {
	msg := types.NewMsgStake(
		suite.addrs[0],
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 10_000_000)),
	)

	handler := farming.NewHandler(suite.keeper)
	_, err := handler(suite.ctx, msg)
	suite.Require().NoError(err)

	_, found := suite.keeper.GetQueuedStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().Equal(true, found)

	queuedCoins := sdk.NewCoins()
	suite.keeper.IterateQueuedStakingsByFarmer(suite.ctx, suite.addrs[0],
		func(stakingCoinDenom string, queuedStaking types.QueuedStaking) (stop bool) {
			queuedCoins = queuedCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
			return false
		},
	)
	suite.Require().Equal(msg.StakingCoins, queuedCoins)
}

func (suite *ModuleTestSuite) TestMsgUnstake() {
	stakeCoin := sdk.NewInt64Coin(denom1, 10_000_000)
	suite.Stake(suite.addrs[0], sdk.NewCoins(stakeCoin))

	_, found := suite.keeper.GetQueuedStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().Equal(true, found)

	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])

	unstakeCoin := sdk.NewInt64Coin(denom1, 5_000_000)
	msg := types.NewMsgUnstake(suite.addrs[0], sdk.NewCoins(unstakeCoin))

	handler := farming.NewHandler(suite.keeper)
	_, err := handler(suite.ctx, msg)
	suite.Require().NoError(err)

	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(balancesBefore.Add(unstakeCoin), balancesAfter))
}

func (suite *ModuleTestSuite) TestMsgHarvest() {
	for _, plan := range suite.samplePlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom2, 10_000_000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-05T00:00:00Z"))
	err := suite.keeper.AllocateRewards(suite.ctx)
	suite.Require().NoError(err)

	rewards := suite.Rewards(suite.addrs[0])

	msg := types.NewMsgHarvest(suite.addrs[0], []string{denom2})

	handler := farming.NewHandler(suite.keeper)
	_, err = handler(suite.ctx, msg)
	suite.Require().NoError(err)

	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(balancesBefore.Add(rewards...), balancesAfter))
	suite.Require().True(suite.app.BankKeeper.GetAllBalances(suite.ctx, types.RewardsReserveAcc).IsZero())
	suite.Require().True(suite.Rewards(suite.addrs[0]).IsZero())
}
