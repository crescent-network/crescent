package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/k0kubun/pp"
)

// tests LiquidStaking, LiquidUnstaking
func (s *KeeperTestSuite) TestLiquidStaking() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(50000)

	// fail, no active validator
	newShares, bTokenMintAmt, err := s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().Error(err)
	s.Require().Equal(newShares, sdk.ZeroDec())
	s.Require().Equal(bTokenMintAmt, sdk.Int{})

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
	newShares, bTokenMintAmt, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
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
	ubdTime, unbondingAmt, ubds, err := s.keeper.LiquidUnstaking(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], ubdBToken)
	s.Require().NoError(err)
	s.Require().Len(ubds, 3)

	// Equal as NetAmount ( no rewards, balance )
	s.Require().EqualValues(unbondingAmt, ubdBToken.Amount)
	s.Require().Equal(ubds[0].DelegatorAddress, s.delAddrs[0].String())
	s.Require().Equal(ubdTime, squadtypes.MustParseRFC3339("2022-03-22T00:00:00Z"))
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
	s.Require().Equal(sdk.NewDec(13334), proxyAccDel1.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel2.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel3.Shares)

	res = s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	s.Require().Equal(sdk.NewDec(13334), res[0].DelShares)

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

	// stack rewards on net amount
	s.advanceHeight(1, false)
	rewards, _, _ := s.keeper.CheckTotalRewards(s.ctx, types.LiquidStakingProxyAcc)
	s.Require().NotEqualValues(rewards, sdk.ZeroDec())

	// last unstaking, unbond all
	btokenBalanceBefore := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], params.LiquidBondDenom).Amount
	s.liquidUnstaking(s.delAddrs[0], btokenBalanceBefore, true)

	// still active liquid validator after unbond all
	alv := s.keeper.GetActiveLiquidValidators(s.ctx, params.WhitelistedValMap())
	s.Require().True(len(alv) != 0)

	// no btoken supply after unbond all
	bTokenTotalSupply = s.app.BankKeeper.GetSupply(s.ctx, liquidBondDenom)
	s.Require().EqualValues(bTokenTotalSupply.Amount, sdk.ZeroInt())

	// TODO: make policy when last unstaking, unbond all, apply netAmount reward, balance policy, active
	proxyBalance := s.app.BankKeeper.GetBalance(s.ctx, types.LiquidStakingProxyAcc, s.app.StakingKeeper.BondDenom(s.ctx)).Amount
	rewards, totalDelShares, totalLiquidTokens := s.keeper.CheckTotalRewards(s.ctx, types.LiquidStakingProxyAcc)
	s.Require().Equal(rewards, sdk.ZeroDec())
	s.Require().Equal(totalDelShares, sdk.ZeroDec())
	s.Require().Equal(totalLiquidTokens, sdk.ZeroInt())
	s.Require().NotEqualValues(proxyBalance, sdk.ZeroInt())
	//s.Require().Equal(proxyBalance, rewards.TruncateInt())

}

// test Liquid Staking gov power
func (s *KeeperTestSuite) TestLiquidStakingGov() {
	params := types.DefaultParams()
	liquidBondDenom := s.keeper.LiquidBondDenom(s.ctx)

	// v1, v2, v3, v4
	vals, valOpers, _ := s.CreateValidators([]int64{10000000, 10000000, 10000000, 10000000, 10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	liquidValidators := s.keeper.GetAllLiquidValidators(s.ctx)

	val4, _ := s.app.StakingKeeper.GetValidator(s.ctx, valOpers[3])

	delA := s.addrs[0]
	delB := s.addrs[1]
	delC := s.addrs[2]
	delD := s.addrs[3]
	delE := s.addrs[4]
	delF := s.addrs[5]
	delG := s.addrs[6]

	_, err := s.app.StakingKeeper.Delegate(s.ctx, delG, sdk.NewInt(60000000), stakingtypes.Unbonded, val4, true)
	s.Require().NoError(err)

	// 7 addr B, C, D, E, F, G, H
	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, tp)
	s.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	s.app.GovKeeper.SetProposal(s.ctx, proposal)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, vals[1], govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, vals[3], govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delC, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delD, govtypes.NewNonSplitVoteOption(govtypes.OptionNoWithVeto)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delE, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delF, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delG, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))

	s.app.StakingKeeper.IterateBondedValidatorsByPower(s.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		return false
	})

	assertTallyResult := func(yes, no, vito, abstain int64) {
		cachedCtx, _ := s.ctx.CacheContext()
		_, _, result := s.app.GovKeeper.Tally(cachedCtx, proposal)
		s.Require().Equal(sdk.NewInt(yes), result.Yes)
		s.Require().Equal(sdk.NewInt(no), result.No)
		s.Require().Equal(sdk.NewInt(vito), result.NoWithVeto)
		s.Require().Equal(sdk.NewInt(abstain), result.Abstain)
	}

	assertTallyResult(80000000, 10000000, 0, 0)

	delAbToken := sdk.NewInt(40000000)
	delBbToken := sdk.NewInt(80000000)
	delCbToken := sdk.NewInt(60000000)
	delDbToken := sdk.NewInt(20000000)
	delEbToken := sdk.NewInt(80000000)
	delFbToken := sdk.NewInt(120000000)
	newShares, bToken, err := s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delA, sdk.NewCoin(sdk.DefaultBondDenom, delAbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delA, liquidBondDenom).Amount, delAbToken)

	newShares, bToken, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, delBbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delB, liquidBondDenom).Amount, delBbToken)

	newShares, bToken, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, delCbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delC, liquidBondDenom).Amount, delCbToken)

	newShares, bToken, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, delDbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delD, liquidBondDenom).Amount, delDbToken)

	newShares, bToken, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, delEbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delE, liquidBondDenom).Amount, delEbToken)

	newShares, bToken, err = s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delF, sdk.NewCoin(sdk.DefaultBondDenom, delFbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delF, liquidBondDenom).Amount, delFbToken)

	totalPower := sdk.ZeroInt()
	totalShare := sdk.ZeroDec()
	s.app.StakingKeeper.IterateBondedValidatorsByPower(s.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		totalPower = totalPower.Add(validator.GetTokens())
		totalShare = totalShare.Add(validator.GetDelegatorShares())
		return false
	})

	assertTallyResult(240000000, 100000000, 20000000, 120000000)

	// Test TallyLiquidGov
	otherVotes := make(govtypes.OtherVotes)
	testOtherVotes := func(voter sdk.AccAddress, bTokenValue sdk.Int) {
		s.Require().Len(otherVotes[voter.String()], liquidValidators.Len())
		for _, v := range liquidValidators {
			s.Require().EqualValues(otherVotes[voter.String()][v.OperatorAddress], bTokenValue.ToDec().QuoInt64(int64(liquidValidators.Len())))
		}
	}
	tallyLiquidGov := func() {
		cachedCtx, _ := s.ctx.CacheContext()
		otherVotes = make(govtypes.OtherVotes)
		votes := s.app.GovKeeper.GetVotes(cachedCtx, proposal.ProposalId)
		s.keeper.TallyLiquidGov(cachedCtx, &votes, &otherVotes)
		squadtypes.PP(otherVotes)

		s.Require().Len(otherVotes, 5)
		testOtherVotes(delB, delBbToken)
		testOtherVotes(delC, delCbToken)
		testOtherVotes(delD, delDbToken)
		testOtherVotes(delE, delEbToken)
		testOtherVotes(delF, delFbToken)
	}

	tallyLiquidGov()

	// Test balance of PoolTokens including bToken
	pair1 := s.createPair(delB, params.LiquidBondDenom, sdk.DefaultBondDenom, false)
	pool1 := s.createPool(delB, pair1.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	tallyLiquidGov()
	pair2 := s.createPair(delC, sdk.DefaultBondDenom, params.LiquidBondDenom, false)
	pool2 := s.createPool(delC, pair2.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	balance := s.app.BankKeeper.GetBalance(s.ctx, delC, pool2.PoolCoinDenom)
	fmt.Println(balance)
	tallyLiquidGov()

	// Test Farming Queued Staking of bToken
	s.CreateFixedAmountPlan(s.addrs[0], map[string]string{params.LiquidBondDenom: "0.4", pool1.PoolCoinDenom: "0.3", pool2.PoolCoinDenom: "0.3"}, map[string]int64{"testdenom": 1})
	s.Stake(delD, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(10000000))))
	queuedStaking, found := s.app.FarmingKeeper.GetQueuedStaking(s.ctx, params.LiquidBondDenom, delD)
	s.True(found)
	s.Equal(queuedStaking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Staking Position Staking of bToken
	s.AdvanceEpoch()
	staking, found := s.app.FarmingKeeper.GetStaking(s.ctx, params.LiquidBondDenom, delD)
	s.True(found)
	s.Equal(staking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Queued Staking of PoolTokens including bToken
	s.Stake(delC, sdk.NewCoins(sdk.NewCoin(pool2.PoolCoinDenom, sdk.NewInt(10000000))))
	queuedStaking, found = s.app.FarmingKeeper.GetQueuedStaking(s.ctx, pool2.PoolCoinDenom, delC)
	s.True(found)
	s.Equal(queuedStaking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Staking Position of PoolTokens including bToken
	s.AdvanceEpoch()
	staking, found = s.app.FarmingKeeper.GetStaking(s.ctx, pool2.PoolCoinDenom, delC)
	s.True(found)
	s.Equal(staking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()
}

// test Liquid Staking gov power
func (s *KeeperTestSuite) TestLiquidStakingGov2() {
	params := types.DefaultParams()

	vals, valOpers, _ := s.CreateValidators([]int64{10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	val1, _ := s.app.StakingKeeper.GetValidator(s.ctx, valOpers[0])

	delA := s.addrs[0]
	delB := s.addrs[1]

	_, err := s.app.StakingKeeper.Delegate(s.ctx, delA, sdk.NewInt(50000000), stakingtypes.Unbonded, val1, true)
	s.Require().NoError(err)

	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, tp)
	s.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	s.app.GovKeeper.SetProposal(s.ctx, proposal)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delA, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))

	cachedCtx, _ := s.ctx.CacheContext()
	_, _, result := s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes)
	s.Require().Equal(sdk.NewInt(0), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))
	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes)
	s.Require().Equal(sdk.NewInt(10000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	newShares, bToken, err := s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delB, params.LiquidBondDenom).Amount, sdk.NewInt(50000000))

	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes)
	s.Require().Equal(sdk.NewInt(60000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain)))

	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes)
	s.Require().Equal(sdk.NewInt(10000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(50000000), result.Abstain)
}
