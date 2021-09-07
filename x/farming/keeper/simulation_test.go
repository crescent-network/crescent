package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/types"
)

type Action interface {
	Do(*KeeperTestSuite)
}

type StakeAction struct {
	farmerAcc sdk.AccAddress
	amount    sdk.Coins
}

func (sa StakeAction) Do(suite *KeeperTestSuite) {
	fmt.Printf("Stake(%s, %s)\n", sa.farmerAcc, sa.amount)
	err := suite.keeper.Stake(suite.ctx, sa.farmerAcc, sa.amount)
	suite.Require().NoError(err)
}

type UnstakeAction struct {
	farmerAcc sdk.AccAddress
	amount    sdk.Coins
}

func (ua UnstakeAction) Do(suite *KeeperTestSuite) {
	fmt.Printf("Unstake(%s, %s)\n", ua.farmerAcc, ua.amount)
	err := suite.keeper.Unstake(suite.ctx, ua.farmerAcc, ua.amount)
	suite.Require().NoError(err)
}

type HarvestAction struct {
	farmerAcc         sdk.AccAddress
	stakingCoinDenoms []string
}

func (ha HarvestAction) Do(suite *KeeperTestSuite) {
	fmt.Printf("Harvest(%s, %s)\n", ha.farmerAcc, ha.stakingCoinDenoms)
	err := suite.keeper.Harvest(suite.ctx, ha.farmerAcc, ha.stakingCoinDenoms)
	suite.Require().NoError(err)
}

type AdvanceEpochAction struct{}

func (AdvanceEpochAction) Do(suite *KeeperTestSuite) {
	fmt.Println("AdvanceEpoch()")
	err := suite.keeper.AdvanceEpoch(suite.ctx)
	suite.Require().NoError(err)
}

type BalanceAssertion struct {
	acc    sdk.AccAddress
	denom  string
	amount sdk.Int
}

func (ba BalanceAssertion) Do(suite *KeeperTestSuite) {
	balance := suite.app.BankKeeper.GetBalance(suite.ctx, ba.acc, ba.denom)
	fmt.Printf("BalanceAssertion(%s, %s, %s)\n", ba.acc, ba.denom, ba.amount)
	suite.Require().True(intEq(ba.amount, balance.Amount))
}

type RewardsAssertion struct {
	acc              sdk.AccAddress
	stakingCoinDenom string
	rewards          sdk.Coins
}

func (ra RewardsAssertion) Do(suite *KeeperTestSuite) {
	current := suite.keeper.GetCurrentRewards(suite.ctx, ra.stakingCoinDenom)
	rewards := suite.keeper.CalculateRewards(suite.ctx, ra.acc, ra.stakingCoinDenom, current.Epoch-1)
	fmt.Printf("RewardsAssertion(%s, %s, %s)\n", ra.acc, ra.stakingCoinDenom, ra.rewards)
	suite.Require().True(coinsEq(ra.rewards, rewards))
}

type TotalRewardsAssertion struct {
	acc     sdk.AccAddress
	rewards sdk.Coins
}

func (tra TotalRewardsAssertion) Do(suite *KeeperTestSuite) {
	fmt.Printf("TotalRewardsAssertion(%s, %s)\n", tra.acc, tra.rewards)
	cacheCtx, _ := suite.ctx.CacheContext()
	rewards, err := suite.keeper.WithdrawAllRewards(cacheCtx, tra.acc)
	suite.Require().NoError(err)
	suite.Require().True(coinsEq(tra.rewards, rewards))
}

func (suite *KeeperTestSuite) TestSimulation() {
	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-09-01T00:00:00Z"))

	for _, plan := range []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				mustParseRFC3339("0001-01-01T00:00:00Z"),
				mustParseRFC3339("9999-12-31T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)),
		),
	} {
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	addrs := simapp.AddTestAddrs(suite.app, suite.ctx, 2, sdk.ZeroInt())
	for _, addr := range addrs {
		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(
			sdk.NewInt64Coin(denom1, 1_000_000_000_000),
			sdk.NewInt64Coin(denom2, 1_000_000_000_000)))
		suite.Require().NoError(err)
	}

	for i, action := range []Action{
		BalanceAssertion{addrs[0], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins()},
		BalanceAssertion{addrs[1], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins()},

		StakeAction{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000))},
		StakeAction{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000), sdk.NewInt64Coin(denom2, 500000))},
		AdvanceEpochAction{},
		BalanceAssertion{addrs[0], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins()},
		BalanceAssertion{addrs[1], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins()},

		AdvanceEpochAction{},
		BalanceAssertion{addrs[0], denom3, sdk.ZeroInt()},
		RewardsAssertion{addrs[0], denom1, sdk.NewCoins(sdk.NewInt64Coin(denom3, 200000))},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 200000))},
		BalanceAssertion{addrs[1], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 800000))},

		StakeAction{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000))},
		AdvanceEpochAction{},
		BalanceAssertion{addrs[0], denom3, sdk.NewInt(400000)},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 0))},
		BalanceAssertion{addrs[1], denom3, sdk.ZeroInt()},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 1600000))},

		UnstakeAction{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 250000), sdk.NewInt64Coin(denom2, 250000))},
		BalanceAssertion{addrs[1], denom3, sdk.NewInt(1600000)},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins()},
		AdvanceEpochAction{},
		BalanceAssertion{addrs[0], denom3, sdk.NewInt(400000)},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 257142))}, // 300000 * (6/7)
		BalanceAssertion{addrs[1], denom3, sdk.NewInt(1600000)},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 742857))}, // 300000 * (1/7) + 700000

		HarvestAction{addrs[0], []string{denom1}},
		HarvestAction{addrs[1], []string{denom1, denom2}},
		BalanceAssertion{addrs[0], denom3, sdk.NewInt(657142)},
		BalanceAssertion{addrs[1], denom3, sdk.NewInt(2342857)},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins()},
		AdvanceEpochAction{},
		TotalRewardsAssertion{addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom3, 257142))},
		TotalRewardsAssertion{addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 742857))},
	} {
		suite.Run(fmt.Sprintf("%d", i), func() {
			action.Do(suite)
		})
	}
}
