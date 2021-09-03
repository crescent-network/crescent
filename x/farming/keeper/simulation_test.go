package keeper_test

import (
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

func NewStakeAction(farmerAcc sdk.AccAddress, amount sdk.Coins) Action {
	return StakeAction{farmerAcc, amount}
}

func (sa StakeAction) Do(suite *KeeperTestSuite) {
	err := suite.keeper.Stake(suite.ctx, sa.farmerAcc, sa.amount)
	suite.Require().NoError(err)
}

type UnstakeAction struct {
	farmerAcc sdk.AccAddress
	amount    sdk.Coins
}

func NewUnstakeAction(farmerAcc sdk.AccAddress, amount sdk.Coins) Action {
	return UnstakeAction{farmerAcc, amount}
}

func (ua UnstakeAction) Do(suite *KeeperTestSuite) {
	err := suite.keeper.Unstake(suite.ctx, ua.farmerAcc, ua.amount)
	suite.Require().NoError(err)
}

type HarvestAction struct {
	farmerAcc         sdk.AccAddress
	stakingCoinDenoms []string
}

func NewHarvestAction(farmerAcc sdk.AccAddress, stakingCoinDenoms []string) Action {
	return HarvestAction{farmerAcc, stakingCoinDenoms}
}

func (ha HarvestAction) Do(suite *KeeperTestSuite) {
	err := suite.keeper.Harvest(suite.ctx, ha.farmerAcc, ha.stakingCoinDenoms)
	suite.Require().NoError(err)
}

type Assertion interface {
	Assert(*KeeperTestSuite)
}

type BalancesAssertion struct {
	acc      sdk.AccAddress
	balances sdk.Coins
}

func NewBalancesAssertion(acc sdk.AccAddress, balances sdk.Coins) Assertion {
	return BalancesAssertion{acc, balances}
}

func (ba BalancesAssertion) Assert(suite *KeeperTestSuite) {
	balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, ba.acc)
	suite.Require().True(coinsEq(ba.balances, balances))
}

type BalanceAssertion struct {
	acc    sdk.AccAddress
	denom  string
	amount sdk.Int
}

func NewBalanceAssertion(acc sdk.AccAddress, denom string, amount sdk.Int) Assertion {
	return BalanceAssertion{acc, denom, amount}
}

func (ba BalanceAssertion) Assert(suite *KeeperTestSuite) {
	balance := suite.app.BankKeeper.GetBalance(suite.ctx, ba.acc, ba.denom)
	suite.Require().True(intEq(ba.amount, balance.Amount))
}

type RewardsAssertion struct {
	acc              sdk.AccAddress
	stakingCoinDenom string
	rewards          sdk.Coins
}

func NewRewardsAssertion(acc sdk.AccAddress, stakingCoinDenom string, rewards sdk.Coins) Assertion {
	return RewardsAssertion{acc, stakingCoinDenom, rewards}
}

func (ra RewardsAssertion) Assert(suite *KeeperTestSuite) {
	current := suite.keeper.GetCurrentRewards(suite.ctx, ra.stakingCoinDenom)
	rewards := suite.keeper.CalculateRewards(suite.ctx, ra.acc, ra.stakingCoinDenom, current.Epoch)
	suite.Require().True(coinsEq(ra.rewards, rewards))
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

	for _, entry := range []struct {
		actions      []Action
		advanceEpoch bool
		assertions   []Assertion
	}{
		{
			[]Action{
				NewStakeAction(addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000))),
			},
			false,
			[]Assertion{
				NewBalanceAssertion(addrs[0], denom3, sdk.ZeroInt()),
				NewRewardsAssertion(addrs[0], denom1, sdk.NewCoins()),
				NewBalanceAssertion(addrs[1], denom3, sdk.ZeroInt()),
				NewRewardsAssertion(addrs[1], denom1, sdk.NewCoins()),
			},
		},
		{
			[]Action{
				NewStakeAction(addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000), sdk.NewInt64Coin(denom2, 500000))),
			},
			true,
			[]Assertion{
				NewBalanceAssertion(addrs[0], denom3, sdk.ZeroInt()),
				NewBalanceAssertion(addrs[1], denom3, sdk.ZeroInt()),
			},
		},
	} {
		for _, action := range entry.actions {
			action.Do(suite)
		}
		if entry.advanceEpoch {
			err := suite.keeper.AdvanceEpoch(suite.ctx)
			suite.Require().NoError(err)
		}
		for _, assertion := range entry.assertions {
			assertion.Assert(suite)
		}
	}
}
