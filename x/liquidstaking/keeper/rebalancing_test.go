package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

func (s *KeeperTestSuite) TestRebalancingCase1() {
	_, valOpers, pks := s.CreateValidators([]int64{1000000, 1000000, 1000000, 1000000, 1000000})
	s.ctx = s.ctx.WithBlockHeight(100).WithBlockTime(utils.ParseTime("2022-03-01T00:00:00Z"))
	params := s.keeper.GetParams(s.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	params.MinLiquidStakingAmount = sdk.NewInt(10000)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(49998)
	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	reds := s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	newShares, bTokenMintAmt, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().NoError(err)
	s.Require().Equal(newShares, stakingAmt.ToDec())
	s.Require().Equal(bTokenMintAmt, stakingAmt)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	proxyAccDel1, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)

	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(16666))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(16666))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(16666))
	totalLiquidTokens, _ := s.keeper.GetAllLiquidValidators(s.ctx).TotalLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	s.Require().EqualValues(stakingAmt, totalLiquidTokens)
	s.printRedelegationsLiquidTokens()

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 3)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)

	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(12501))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(12499))
	totalLiquidTokens, _ = s.keeper.GetAllLiquidValidators(s.ctx).TotalLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	s.Require().EqualValues(stakingAmt, totalLiquidTokens)
	s.printRedelegationsLiquidTokens()

	//reds := s.app.StakingKeeper.GetRedelegations(s.ctx, types.LiquidStakingProxyAcc, 20)
	s.Require().Len(reds, 3)

	utils.PP("before complete")
	utils.PP(s.keeper.GetAllLiquidValidatorStates(s.ctx))
	utils.PP(s.keeper.GetNetAmountState(s.ctx))

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	utils.PP("after complete")
	utils.PP(s.keeper.GetAllLiquidValidatorStates(s.ctx))
	utils.PP(s.keeper.GetNetAmountState(s.ctx))

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 4)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)
	proxyAccDel5, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().True(found)

	s.printRedelegationsLiquidTokens()
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(10002))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel5.Shares.TruncateInt(), sdk.NewInt(9999))
	totalLiquidTokens, _ = s.keeper.GetAllLiquidValidators(s.ctx).TotalLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	s.Require().EqualValues(stakingAmt, totalLiquidTokens)

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// remove whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
	}

	utils.PP(s.keeper.GetAllLiquidValidatorStates(s.ctx))
	s.keeper.SetParams(s.ctx, params)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 4)
	utils.PP(s.keeper.GetAllLiquidValidatorStates(s.ctx))

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)
	proxyAccDel5, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().False(found)

	s.printRedelegationsLiquidTokens()
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(12501))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(12499))
	totalLiquidTokens, _ = s.keeper.GetAllLiquidValidators(s.ctx).TotalLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	s.Require().EqualValues(stakingAmt, totalLiquidTokens)

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// remove whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
	}

	s.keeper.SetParams(s.ctx, params)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 3)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().False(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().False(found)
	proxyAccDel5, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().False(found)

	s.printRedelegationsLiquidTokens()
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(24999))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(24999))
	totalLiquidTokens, _ = s.keeper.GetAllLiquidValidators(s.ctx).TotalLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	s.Require().EqualValues(stakingAmt, totalLiquidTokens)

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// double sign, tombstone, slash, jail
	s.doubleSign(valOpers[1], sdk.ConsAddress(pks[1].Address()))

	// check inactive with zero weight after tombstoned
	lvState, found := s.keeper.GetLiquidValidatorState(s.ctx, proxyAccDel2.GetValidatorAddr())
	s.Require().True(found)
	s.Require().Equal(lvState.Status, types.ValidatorStatusInactive)
	s.Require().Equal(lvState.Weight, sdk.ZeroInt())
	s.Require().NotEqualValues(lvState.DelShares, sdk.ZeroDec())
	s.Require().NotEqualValues(lvState.LiquidTokens, sdk.ZeroInt())

	// rebalancing, remove tombstoned liquid validator
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 1)

	// all redelegated, no delShares
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().False(found)

	// liquid validator removed, invalid after tombstoned
	lvState, found = s.keeper.GetLiquidValidatorState(s.ctx, valOpers[1])
	s.Require().False(found)
	s.Require().Equal(lvState.OperatorAddress, valOpers[1].String())
	s.Require().Equal(lvState.Status, types.ValidatorStatusUnspecified)
	s.Require().EqualValues(lvState.DelShares, sdk.ZeroDec())
	s.Require().EqualValues(lvState.LiquidTokens, sdk.ZeroInt())

	// jail last liquid validator, undelegate all liquid tokens to proxy acc
	s.doubleSign(valOpers[0], sdk.ConsAddress(pks[0].Address()))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	// no delegation of proxy acc
	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().False(found)
	val1, found := s.app.StakingKeeper.GetValidator(s.ctx, valOpers[0])
	s.Require().True(found)
	s.Require().Equal(val1.Status, stakingtypes.Unbonding)

	// check unbonding delegation to proxy acc
	ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, types.LiquidStakingProxyAcc, val1.GetOperator())
	s.Require().True(found)

	// complete unbonding
	s.completeRedelegationUnbonding()

	// check validator Unbonded
	val1, found = s.app.StakingKeeper.GetValidator(s.ctx, valOpers[0])
	s.Require().True(found)
	s.Require().Equal(val1.Status, stakingtypes.Unbonded)

	// no rewards, delShares, liquid tokens
	nas := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nas.TotalRemainingRewards, sdk.ZeroDec())
	s.Require().EqualValues(nas.TotalDelShares, sdk.ZeroDec())
	s.Require().EqualValues(nas.TotalLiquidTokens, sdk.ZeroInt())

	// unbonded to balance, equal with netAmount
	s.Require().EqualValues(ubd.Entries[0].Balance, nas.ProxyAccBalance)
	s.Require().EqualValues(nas.NetAmount.TruncateInt(), nas.ProxyAccBalance)

	// mintRate over 1 due to slashing
	s.Require().True(nas.MintRate.GT(sdk.OneDec()))
	bTokenBalanceBefore := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], params.LiquidBondDenom).Amount
	nativeTokenBalanceBefore := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom).Amount
	s.Require().EqualValues(nas.BtokenTotalSupply, bTokenBalanceBefore)

	// withdraw directly unstaking when no totalLiquidTokens
	s.Require().NoError(s.liquidUnstaking(s.delAddrs[0], bTokenBalanceBefore, false))
	bTokenBalanceAfter := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], params.LiquidBondDenom).Amount
	nativeTokenBalanceAfter := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom).Amount
	s.Require().EqualValues(bTokenBalanceAfter, sdk.ZeroInt())
	s.Require().EqualValues(nativeTokenBalanceAfter.Sub(nativeTokenBalanceBefore), nas.NetAmount.TruncateInt())

	// zero net amount states
	s.RequireNetAmountStateZero()
}

func (s *KeeperTestSuite) TestRebalancingConsecutiveCase() {
	_, valOpers, _ := s.CreateValidators([]int64{
		1000000, 1000000, 1000000, 1000000, 1000000,
		1000000, 1000000, 1000000, 1000000, 1000000,
		1000000, 1000000, 1000000, 1000000, 1000000})
	s.ctx = s.ctx.WithBlockHeight(100).WithBlockTime(utils.ParseTime("2022-03-01T00:00:00Z"))
	params := s.keeper.GetParams(s.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	params.MinLiquidStakingAmount = sdk.NewInt(10000)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(10000000000000)
	s.fundAddr(s.delAddrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt)))
	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	reds := s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	newShares, bTokenMintAmt, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().NoError(err)
	s.Require().Equal(newShares, stakingAmt.ToDec())
	s.Require().Equal(bTokenMintAmt, stakingAmt)
	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[8].String(), TargetWeight: sdk.NewInt(5)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 8)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()
	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[8].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[9].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 9)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()
	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	// complete redelegations
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24 * 20).Add(time.Hour))
	staking.EndBlocker(s.ctx, *s.app.StakingKeeper)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)
	// assert rebalanced
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()

	// remove active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[8].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[9].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 9)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()
	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[8].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[9].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[10].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[11].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[12].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 11)
	// fail rebalancing due to redelegation hopping
	s.Require().Equal(s.redelegationsErrorCount(reds), 11)
	s.printRedelegationsLiquidTokens()

	// complete redelegation and retry
	s.completeRedelegationUnbonding()
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.printRedelegationsLiquidTokens()
	s.Require().Len(reds, 11)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)

	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)

	// modify weight
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[5].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[6].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[7].String(), TargetWeight: sdk.NewInt(5)},
		{ValidatorAddress: valOpers[8].String(), TargetWeight: sdk.NewInt(5)},
		{ValidatorAddress: valOpers[9].String(), TargetWeight: sdk.NewInt(5)},
		{ValidatorAddress: valOpers[10].String(), TargetWeight: sdk.NewInt(5)},
		{ValidatorAddress: valOpers[11].String(), TargetWeight: sdk.NewInt(5)},
		{ValidatorAddress: valOpers[12].String(), TargetWeight: sdk.NewInt(5)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 6)
	// fail rebalancing partially due to redelegation hopping
	s.Require().Equal(s.redelegationsErrorCount(reds), 3)
	s.printRedelegationsLiquidTokens()

	// additional liquid staking when not rebalanced
	_, _, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
	s.Require().NoError(err)
	s.printRedelegationsLiquidTokens()

	// complete some redelegations
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24 * 20).Add(time.Hour))
	staking.EndBlocker(s.ctx, *s.app.StakingKeeper)
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 9)

	// failed redelegations with small amount (less than rebalancing trigger)
	s.Require().Equal(s.redelegationsErrorCount(reds), 6)
	s.printRedelegationsLiquidTokens()

	// assert rebalanced
	reds = s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.Require().Len(reds, 0)
	s.Require().Equal(s.redelegationsErrorCount(reds), 0)
	s.printRedelegationsLiquidTokens()
}

func (s *KeeperTestSuite) TestWithdrawRewardsAndReStaking() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 1000000, 1000000})
	params := s.keeper.GetParams(s.ctx)

	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(100000000)
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], stakingAmt))

	// no rewards
	totalRewards, totalDelShares, totalLiquidTokens := s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.EqualValues(totalRewards, sdk.ZeroDec())
	s.EqualValues(totalDelShares, stakingAmt.ToDec(), totalLiquidTokens)

	// allocate rewards
	s.advanceHeight(100, false)
	totalRewards, totalDelShares, totalLiquidTokens = s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.NotEqualValues(totalRewards, sdk.ZeroDec())
	s.NotEqualValues(totalLiquidTokens, sdk.ZeroDec())

	// withdraw rewards and re-staking
	whitelistedValsMap := types.GetWhitelistedValsMap(params.WhitelistedValidators)
	s.keeper.WithdrawRewardsAndReStake(s.ctx, whitelistedValsMap)
	totalRewardsAfter, totalDelSharesAfter, totalLiquidTokensAfter := s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.EqualValues(totalRewardsAfter, sdk.ZeroDec())
	s.EqualValues(totalDelSharesAfter, totalRewards.TruncateDec().Add(totalDelShares), totalLiquidTokensAfter)
}

func (s *KeeperTestSuite) TestRemoveAllLiquidValidator() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 1000000, 1000000})
	params := s.keeper.GetParams(s.ctx)

	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(100000000)
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], stakingAmt))

	// allocate rewards
	s.advanceHeight(1, false)
	nasBefore := s.keeper.GetNetAmountState(s.ctx)
	s.Require().NotEqualValues(nasBefore.TotalRemainingRewards, sdk.ZeroDec())
	s.Require().NotEqualValues(nasBefore.TotalDelShares, sdk.ZeroDec())
	s.Require().NotEqualValues(nasBefore.NetAmount, sdk.ZeroDec())
	s.Require().NotEqualValues(nasBefore.TotalLiquidTokens, sdk.ZeroInt())
	s.Require().EqualValues(nasBefore.ProxyAccBalance, sdk.ZeroInt())

	// remove all whitelist
	params.WhitelistedValidators = []types.WhitelistedValidator{}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// no liquid validator
	lvs := s.keeper.GetAllLiquidValidators(s.ctx)
	s.Require().Len(lvs, 0)

	nasAfter := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nasAfter.TotalRemainingRewards, sdk.ZeroDec())
	s.Require().EqualValues(nasAfter.ProxyAccBalance, nasBefore.TotalRemainingRewards.TruncateInt())
	s.Require().EqualValues(nasAfter.TotalDelShares, sdk.ZeroDec())
	s.Require().EqualValues(nasAfter.TotalLiquidTokens, sdk.ZeroInt())
	s.Require().EqualValues(nasBefore.NetAmount.TruncateInt(), nasAfter.NetAmount.TruncateInt())

	s.completeRedelegationUnbonding()
	nasAfter2 := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nasAfter2.ProxyAccBalance, nasAfter.ProxyAccBalance.Add(nasBefore.TotalLiquidTokens))
	s.Require().EqualValues(nasAfter2.NetAmount.TruncateInt(), nasBefore.NetAmount.TruncateInt())
}
