package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func (suite *KeeperTestSuite) TestScenario1() {
	assertBalance := func(addr sdk.AccAddress, denom string, amt int64) {
		suite.T().Helper()
		suite.Require().True(intEq(sdk.NewInt(amt), suite.app.BankKeeper.GetBalance(suite.ctx, addr, denom).Amount))
	}
	assertAllRewards := func(addr sdk.AccAddress, rewards string) {
		suite.T().Helper()
		suite.Require().True(coinsEq(utils.ParseCoins(rewards), suite.AllRewards(addr)))
	}
	assertAllUnharvestedRewards := func(addr sdk.AccAddress, rewards string) {
		suite.T().Helper()
		suite.Require().True(coinsEq(utils.ParseCoins(rewards), suite.allUnharvestedRewards(addr)))
	}

	for _, plan := range []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1, "", types.PlanTypePrivate, suite.addrs[0].String(), suite.addrs[0].String(),
				parseDecCoins("0.3denom1,0.7denom2"), sampleStartTime, sampleEndTime),
			utils.ParseCoins("1000000denom3")),
	} {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	addrs := chain.AddTestAddrs(suite.app, suite.ctx, 2, sdk.ZeroInt())
	for _, addr := range addrs {
		suite.Require().NoError(
			chain.FundAccount(
				suite.app.BankKeeper, suite.ctx, addr,
				utils.ParseCoins("1000000000000denom1, 1000000000000denom2")))
	}

	assertBalance(addrs[0], denom3, 0)
	assertAllRewards(addrs[0], "")
	assertBalance(addrs[1], denom3, 0)
	assertAllRewards(addrs[1], "")

	suite.executeBlock(utils.ParseTime("2022-04-01T23:00:00Z"), func() {
		suite.Stake(addrs[0], utils.ParseCoins("1000000denom1"))
		suite.Stake(addrs[1], utils.ParseCoins("500000denom1,500000denom2"))
	})
	suite.executeBlock(utils.ParseTime("2022-04-02T00:00:00Z"), nil) // next epoch
	suite.executeBlock(utils.ParseTime("2022-04-02T23:00:00Z"), nil) // queued -> staked

	assertBalance(addrs[0], denom3, 0)
	assertAllRewards(addrs[0], "")
	assertBalance(addrs[1], denom3, 0)
	assertAllRewards(addrs[1], "")

	suite.executeBlock(utils.ParseTime("2022-04-03T00:00:00Z"), nil) // rewards distribution
	assertBalance(addrs[0], denom3, 0)
	assertAllRewards(addrs[0], "200000denom3") // 300000 * 2/3
	assertBalance(addrs[1], denom3, 0)
	assertAllRewards(addrs[1], "800000denom3") // 300000 * 1/3 + 700000

	suite.executeBlock(utils.ParseTime("2022-04-03T01:00:00Z"), func() {
		suite.Stake(addrs[0], utils.ParseCoins("500000denom1"))
	})
	suite.executeBlock(utils.ParseTime("2022-04-04T00:00:00Z"), nil)
	assertAllRewards(addrs[0], "400000denom3")
	assertAllUnharvestedRewards(addrs[0], "")
	suite.executeBlock(utils.ParseTime("2022-04-04T01:00:00Z"), nil)
	assertBalance(addrs[0], denom3, 0)
	assertAllRewards(addrs[0], "")
	assertAllUnharvestedRewards(addrs[0], "400000denom3")
	assertBalance(addrs[1], denom3, 0)
	assertAllRewards(addrs[1], "1600000denom3")

	suite.executeBlock(utils.ParseTime("2022-04-04T01:30:00Z"), func() {
		// User can unstake multiple times before the end of the epoch
		suite.Unstake(addrs[1], utils.ParseCoins("200000denom1,200000denom2"))
		suite.Unstake(addrs[1], utils.ParseCoins("50000denom1,50000denom2"))
		// addrs[1] now has 250000denom1,250000denom2 staked
	})

	assertBalance(addrs[1], denom3, 0)
	assertAllRewards(addrs[1], "")
	assertAllUnharvestedRewards(addrs[1], "1600000denom3")

	suite.executeBlock(utils.ParseTime("2022-04-05T00:00:00Z"), nil)
	assertAllRewards(addrs[0], "257142denom3") // 300000 * (6/7)
	assertAllRewards(addrs[1], "742857denom3") // 300000 * (1/7) + 700000

	suite.executeBlock(utils.ParseTime("2022-04-05T12:00:00Z"), func() {
		suite.Harvest(addrs[0], []string{denom1})
		suite.Harvest(addrs[1], []string{denom1, denom2})
		suite.Harvest(addrs[1], []string{denom1, denom2})
		suite.Harvest(addrs[1], []string{denom1, denom2})
	})

	assertBalance(addrs[0], denom3, 657142)
	assertBalance(addrs[1], denom3, 2342857)
	assertAllRewards(addrs[0], "")
	assertAllRewards(addrs[1], "")

	suite.executeBlock(utils.ParseTime("2022-04-06T00:00:00Z"), nil)

	assertAllRewards(addrs[0], "257142denom3")
	assertAllRewards(addrs[1], "742857denom3")
}
