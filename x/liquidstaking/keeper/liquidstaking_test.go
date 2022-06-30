package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
)

// tests LiquidStake, LiquidUnstake
func (s *KeeperTestSuite) TestLiquidStake() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	params.MinLiquidStakingAmount = sdk.NewInt(50000)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := params.MinLiquidStakingAmount

	// fail, no active validator
	cachedCtx, _ := s.ctx.CacheContext()
	newShares, bTokenMintAmt, err := s.keeper.LiquidStake(cachedCtx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().ErrorIs(err, types.ErrActiveLiquidValidatorsNotExists)
	s.Require().Equal(newShares, sdk.ZeroDec())
	s.Require().Equal(bTokenMintAmt, sdk.ZeroInt())

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	res := s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	s.Require().Equal(sdk.ZeroDec(), res[0].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[0].LiquidTokens)

	s.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	s.Require().Equal(sdk.ZeroDec(), res[1].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[1].LiquidTokens)

	s.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	s.Require().Equal(sdk.ZeroDec(), res[2].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[2].LiquidTokens)

	// liquid staking
	newShares, bTokenMintAmt, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().NoError(err)
	s.Require().Equal(newShares, stakingAmt.ToDec())
	s.Require().Equal(bTokenMintAmt, stakingAmt)

	_, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[0])
	s.Require().False(found)
	_, found = s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[1])
	s.Require().False(found)
	_, found = s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[2])
	s.Require().False(found)

	proxyAccDel1, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	s.Require().Equal(proxyAccDel1.Shares, sdk.NewDec(16668)) // 16666 + add crumb 2 to 1st active validator
	s.Require().Equal(proxyAccDel2.Shares, sdk.NewDec(16666))
	s.Require().Equal(proxyAccDel2.Shares, sdk.NewDec(16666))
	s.Require().Equal(stakingAmt.ToDec(), proxyAccDel1.Shares.Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	liquidBondDenom := s.keeper.LiquidBondDenom(s.ctx)
	balanceBeforeUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceBeforeUBD.Amount, sdk.NewInt(999950000))
	ubdBToken := sdk.NewCoin(liquidBondDenom, sdk.NewInt(10000))
	bTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], liquidBondDenom)
	bTokenTotalSupply := s.app.BankKeeper.GetSupply(s.ctx, liquidBondDenom)
	s.Require().Equal(bTokenBalance, sdk.NewCoin(liquidBondDenom, sdk.NewInt(50000)))
	s.Require().Equal(bTokenBalance, bTokenTotalSupply)

	// liquid unstaking
	ubdTime, unbondingAmt, ubds, unbondedAmt, err := s.keeper.LiquidUnstake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], ubdBToken)
	s.Require().NoError(err)
	s.Require().EqualValues(unbondedAmt, sdk.ZeroInt())
	s.Require().Len(ubds, 3)

	// crumb excepted on unbonding
	crumb := ubdBToken.Amount.Sub(ubdBToken.Amount.QuoRaw(3).MulRaw(3)) // 1
	s.Require().EqualValues(unbondingAmt, ubdBToken.Amount.Sub(crumb))  // 9999
	s.Require().Equal(ubds[0].DelegatorAddress, s.delAddrs[0].String())
	s.Require().Equal(ubdTime, utils.ParseTime("2022-03-22T00:00:00Z"))
	bTokenBalanceAfter := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], liquidBondDenom)
	s.Require().Equal(bTokenBalanceAfter, sdk.NewCoin(liquidBondDenom, sdk.NewInt(40000)))

	balanceBeginUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceBeginUBD.Amount, balanceBeforeUBD.Amount)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	s.Require().Equal(stakingAmt.Sub(unbondingAmt).ToDec(), proxyAccDel1.GetShares().Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	// complete unbonding
	s.ctx = s.ctx.WithBlockHeight(200).WithBlockTime(ubdTime.Add(1))
	updates := s.app.StakingKeeper.BlockValidatorUpdates(s.ctx) // EndBlock of staking keeper, mature UBD
	s.Require().Empty(updates)
	balanceCompleteUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceCompleteUBD.Amount, balanceBeforeUBD.Amount.Add(unbondingAmt))

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	// crumb added to first valid active liquid validator
	s.Require().Equal(sdk.NewDec(13335), proxyAccDel1.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel2.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel3.Shares)

	res = s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	s.Require().Equal(sdk.NewDec(13335), res[0].DelShares)

	s.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	s.Require().Equal(sdk.NewDec(13333), res[1].DelShares)

	s.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	s.Require().Equal(sdk.NewDec(13333), res[2].DelShares)

	// stack and withdraw liquid rewards and re-staking
	s.advanceHeight(10, true)
	rewards, _, _ := s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.Require().EqualValues(rewards, sdk.ZeroDec())

	// stack rewards on net amount
	s.advanceHeight(1, false)
	rewards, _, _ = s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.Require().NotEqualValues(rewards, sdk.ZeroDec())

	// failed requesting liquid unstaking bTokenTotalSupply when existing remaining rewards
	bTokenTotalSupply = s.app.BankKeeper.GetSupply(s.ctx, liquidBondDenom)
	btokenBalanceBefore := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], params.LiquidBondDenom).Amount
	s.Require().EqualValues(bTokenTotalSupply.Amount, btokenBalanceBefore)
	s.Require().ErrorIs(s.liquidUnstaking(s.delAddrs[0], btokenBalanceBefore, true), sdkerrors.ErrInvalidRequest)

	// all remaining rewards re-staked, request last unstaking, unbond all
	s.advanceHeight(1, true)
	rewards, _, _ = s.keeper.CheckDelegationStates(s.ctx, types.LiquidStakingProxyAcc)
	s.Require().EqualValues(rewards, sdk.ZeroDec())
	s.Require().NoError(s.liquidUnstaking(s.delAddrs[0], btokenBalanceBefore, true))

	// still active liquid validator after unbond all
	alv := s.keeper.GetActiveLiquidValidators(s.ctx, params.WhitelistedValsMap())
	s.Require().True(len(alv) != 0)

	// no btoken supply and netAmount after unbond all
	nas := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nas.BtokenTotalSupply, sdk.ZeroInt())
	s.Require().Equal(nas.TotalRemainingRewards, sdk.ZeroDec())
	s.Require().Equal(nas.TotalDelShares, sdk.ZeroDec())
	s.Require().Equal(nas.TotalLiquidTokens, sdk.ZeroInt())
	s.Require().Equal(nas.ProxyAccBalance, sdk.ZeroInt())
	s.Require().Equal(nas.NetAmount, sdk.ZeroDec())
}

func (s *KeeperTestSuite) TestLiquidStakeFromVestingAccount() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	from := s.delAddrs[0]
	vestingAmt := s.app.BankKeeper.GetAllBalances(s.ctx, from)
	vestingStartTime := s.ctx.BlockTime().Add(1 * time.Hour)
	vestingEndTime := s.ctx.BlockTime().Add(2 * time.Hour)
	vestingMidTime := s.ctx.BlockTime().Add(90 * time.Minute)

	vestingAccAddr := "cosmos10n3ncmlsaqfuwsmfll8kq6hvt4x7c8cznmllss"
	vestingAcc, err := sdk.AccAddressFromBech32(vestingAccAddr)
	s.Require().NoError(err)

	// createContinuousVestingAccount
	cVestingAcc := s.createContinuousVestingAccount(from, vestingAcc, vestingAmt, vestingStartTime, vestingEndTime)
	spendableCoins := s.app.BankKeeper.SpendableCoins(s.ctx, cVestingAcc.GetAddress())
	s.Require().True(spendableCoins.IsZero())
	lockedCoins := s.app.BankKeeper.LockedCoins(s.ctx, cVestingAcc.GetAddress())
	s.Require().EqualValues(lockedCoins, vestingAmt)

	// failed liquid staking, no spendable coins on the vesting account ( not allowed locked coins )
	err = s.liquidStaking(vestingAcc, vestingAmt.AmountOf(sdk.DefaultBondDenom))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)

	// release some vesting coins
	s.ctx = s.ctx.WithBlockTime(vestingMidTime)
	spendableCoins = s.app.BankKeeper.SpendableCoins(s.ctx, cVestingAcc.GetAddress())
	s.Require().True(spendableCoins.IsAllPositive())
	lockedCoins = s.app.BankKeeper.LockedCoins(s.ctx, cVestingAcc.GetAddress())
	s.Require().True(lockedCoins.IsAllPositive())

	// success with released spendable coins
	err = s.liquidStaking(vestingAcc, spendableCoins.AmountOf(sdk.DefaultBondDenom))
	s.Require().NoError(err)
	nas := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nas.TotalLiquidTokens, spendableCoins.AmountOf(sdk.DefaultBondDenom))
}

func (s *KeeperTestSuite) TestLiquidStakeEdgeCases() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)
	stakingAmt := sdk.NewInt(5000000)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// fail Invalid BondDenom case
	_, _, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin("bad", stakingAmt))
	s.Require().ErrorIs(err, types.ErrInvalidBondDenom)

	// liquid staking, unstaking with huge amount
	hugeAmt := sdk.NewInt(1_000_000_000_000_000_000)
	s.fundAddr(s.delAddrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, hugeAmt.MulRaw(2))))
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], hugeAmt))
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], hugeAmt))
	s.Require().NoError(s.liquidUnstaking(s.delAddrs[0], sdk.NewInt(10), true))
	s.Require().NoError(s.liquidUnstaking(s.delAddrs[0], hugeAmt, true))
	s.keeper.UpdateLiquidValidatorSet(s.ctx)
	s.completeRedelegationUnbonding()
	states := s.keeper.GetNetAmountState(s.ctx)
	states.TotalLiquidTokens.Equal(hugeAmt)
}

func (s *KeeperTestSuite) TestLiquidUnstakeEdgeCases() {
	mintParams := s.app.MintKeeper.GetParams(s.ctx)
	mintParams.InflationSchedules = []minttypes.InflationSchedule{}
	s.app.MintKeeper.SetParams(s.ctx, mintParams)

	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)
	stakingAmt := sdk.NewInt(5000000)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// success liquid staking
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], stakingAmt))

	// fail when liquid unstaking with too small amount
	_, _, _, _, err := s.liquidUnstakingWithResult(s.delAddrs[0], sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(2)))
	s.Require().ErrorIs(err, types.ErrTooSmallLiquidUnstakingAmount)

	// fail when liquid unstaking with zero amount
	_, _, _, _, err = s.liquidUnstakingWithResult(s.delAddrs[0], sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(0)))
	s.Require().ErrorIs(err, types.ErrTooSmallLiquidUnstakingAmount)

	// fail when invalid liquid bond denom
	_, _, _, _, err = s.liquidUnstakingWithResult(s.delAddrs[0], sdk.NewCoin("stake", sdk.NewInt(10000)))
	s.Require().ErrorIs(err, types.ErrInvalidLiquidBondDenom)

	// verify that there is no problem performing liquid unstaking as much as the MaxEntries
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	for i := uint32(0); i < stakingParams.MaxEntries; i++ {
		s.Require().NoError(s.liquidUnstaking(s.delAddrs[0], sdk.NewInt(1000), false))
	}
	// fail in an attempt beyond MaxEntries
	s.Require().ErrorIs(s.liquidUnstaking(s.delAddrs[0], sdk.NewInt(1000), false), stakingtypes.ErrMaxUnbondingDelegationEntries)
	dels := s.app.StakingKeeper.GetUnbondingDelegations(s.ctx, s.delAddrs[0], 100)
	for _, ubd := range dels {
		s.Require().EqualValues(stakingParams.MaxEntries, len(ubd.Entries))
	}

	// set empty whitelisted, active liquid validator
	params.WhitelistedValidators = []types.WhitelistedValidator{}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// error case where there is a quantity that are unbonding balance or remaining rewards that is not re-stake or withdrawn in netAmount.
	_, _, _, _, err = s.liquidUnstakingWithResult(s.delAddrs[0], sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(1000)))
	s.Require().ErrorIs(err, types.ErrInsufficientProxyAccBalance)

	// success after complete unbonding
	s.completeRedelegationUnbonding()
	ubdTime, unbondingAmt, ubds, unbondedAmt, err := s.liquidUnstakingWithResult(s.delAddrs[0], sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(1000)))
	s.Require().NoError(err)
	s.Require().EqualValues(unbondedAmt, sdk.NewInt(1000))
	s.Require().EqualValues(unbondingAmt, sdk.ZeroInt())
	s.Require().EqualValues(ubdTime, time.Time{})
	s.Require().Len(ubds, 0)
}
