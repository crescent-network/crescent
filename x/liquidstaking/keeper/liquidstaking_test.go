package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/k0kubun/pp"
)

// tests LiquidStaking, LiquidUnstaking
func (suite *KeeperTestSuite) TestLiquidStaking() {
	_, valOpers := suite.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := suite.keeper.GetParams(suite.ctx)
	suite.keeper.EndBlocker(suite.ctx)

	stakingAmt := sdk.NewInt(50000)

	// fail, no active validator
	newShares, bTokenMintAmt, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().Error(err)
	suite.Require().Equal(newShares, sdk.ZeroDec())
	suite.Require().Equal(bTokenMintAmt, sdk.Int{})

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	suite.keeper.EndBlocker(suite.ctx)

	res := suite.keeper.GetLiquidValidatorStates(suite.ctx)
	suite.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	suite.Require().Equal(sdk.ZeroInt(), res[0].DelShares)

	suite.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	suite.Require().Equal(sdk.ZeroInt(), res[1].DelShares)

	suite.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	suite.Require().Equal(sdk.ZeroInt(), res[2].DelShares)

	valMap := suite.keeper.GetValidatorsMap(suite.ctx)
	activeVals := suite.keeper.GetActiveLiquidValidators(suite.ctx, valMap, params.WhitelistedValMap())
	_, crumb := types.DivideByWeight(activeVals, stakingAmt, params.WhitelistedValMap())
	newShares, bTokenMintAmt, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().NoError(err)
	suite.Require().Equal(newShares.Add(crumb.ToDec()), stakingAmt.ToDec())
	suite.Require().Equal(bTokenMintAmt, stakingAmt)

	_, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, suite.delAddrs[0], valOpers[0])
	suite.Require().False(found)
	_, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, suite.delAddrs[0], valOpers[1])
	suite.Require().False(found)
	_, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, suite.delAddrs[0], valOpers[2])
	suite.Require().False(found)

	proxyAccDel1, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)
	suite.Require().Equal(proxyAccDel1.Shares, stakingAmt.ToDec().QuoInt64(3).TruncateDec())
	suite.Require().Equal(stakingAmt.QuoRaw(3).MulRaw(3).ToDec(), proxyAccDel1.Shares.Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	balanceBeforeUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	suite.Require().Equal(balanceBeforeUBD.Amount, sdk.NewInt(999950000))

	bondedBondDenom := suite.keeper.BondedBondDenom(suite.ctx)
	ubdAmt := sdk.NewCoin(bondedBondDenom, sdk.NewInt(10000))
	bTokenBalance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], bondedBondDenom)
	bTokenTotalSupply := suite.app.BankKeeper.GetSupply(suite.ctx, bondedBondDenom)
	suite.Require().Equal(bTokenBalance, sdk.NewCoin(bondedBondDenom, sdk.NewInt(50000)))
	suite.Require().Equal(bTokenBalance, bTokenTotalSupply)

	ubdTime, unbondingAmt, ubds, err := suite.keeper.LiquidUnstaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], ubdAmt)
	suite.Require().NoError(err)
	suite.Require().Len(ubds, 3)
	// truncated shares
	suite.Require().EqualValues(unbondingAmt, ubdAmt.Amount.QuoRaw(3).MulRaw(3).ToDec())
	//suite.Require().Equal(unbondingAmt, ubdAmt.Amount.ToDec())
	suite.Require().Equal(ubds[0].DelegatorAddress, suite.delAddrs[0].String())
	suite.Require().Equal(ubdTime, squadtypes.MustParseRFC3339("2022-03-22T00:00:00Z"))
	bTokenBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], bondedBondDenom)
	suite.Require().Equal(bTokenBalanceAfter, sdk.NewCoin(bondedBondDenom, sdk.NewInt(40000)))

	balanceBeginUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	suite.Require().Equal(balanceBeginUBD.Amount, balanceBeforeUBD.Amount)

	proxyAccDel1, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)
	suite.Require().Equal(stakingAmt.Sub(unbondingAmt.TruncateInt()).Sub(crumb).ToDec(), proxyAccDel1.GetShares().Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	suite.ctx = suite.ctx.WithBlockHeight(200).WithBlockTime(ubdTime.Add(1))
	updates := suite.app.StakingKeeper.BlockValidatorUpdates(suite.ctx) // EndBlock of staking keeper
	suite.Require().Empty(updates)
	balanceCompleteUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	suite.Require().Equal(balanceCompleteUBD.Amount, balanceBeforeUBD.Amount.Add(unbondingAmt.TruncateInt()))

	proxyAccDel1, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)
	suite.Require().Equal(sdk.MustNewDecFromStr("13333.0"), proxyAccDel1.Shares)
	suite.Require().Equal(sdk.MustNewDecFromStr("13333.0"), proxyAccDel2.Shares)
	suite.Require().Equal(sdk.MustNewDecFromStr("13333.0"), proxyAccDel3.Shares)

	res = suite.keeper.GetLiquidValidatorStates(suite.ctx)
	suite.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	suite.Require().Equal(sdk.NewInt(13333), res[0].DelShares)

	suite.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	suite.Require().Equal(sdk.NewInt(13333), res[1].DelShares)

	suite.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	suite.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	suite.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	suite.Require().Equal(sdk.NewInt(13333), res[2].DelShares)

	// test withdraw liquid reward and re-staking
	suite.advanceHeight(100, true)
	// invariant check
	crisis.EndBlocker(suite.ctx, suite.app.CrisisKeeper)
	// TODO: add cases for different weight
}

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestLiquidStakingGov() {
	params := types.DefaultParams()
	bondedBondDenom := suite.keeper.BondedBondDenom(suite.ctx)

	// v1, v2, v3, v4
	vals, valOpers := suite.CreateValidators([]int64{10000000, 10000000, 10000000, 10000000, 10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(10)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	suite.keeper.EndBlocker(suite.ctx)

	liquidValidators := suite.keeper.GetAllLiquidValidators(suite.ctx)
	lValMap := liquidValidators.Map()
	fmt.Println(lValMap)

	val4, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[3])

	delA := suite.addrs[0]
	delB := suite.addrs[1]
	delC := suite.addrs[2]
	delD := suite.addrs[3]
	delE := suite.addrs[4]
	delF := suite.addrs[5]
	delG := suite.addrs[6]

	_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delG, sdk.NewInt(60000000), stakingtypes.Unbonded, val4, true)
	suite.Require().NoError(err)

	// 7 addr B, C, D, E, F, G, H
	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tp)
	suite.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[1], govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[3], govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delC, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delD, govtypes.NewNonSplitVoteOption(govtypes.OptionNoWithVeto)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delE, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delF, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delG, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))

	suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		return false
	})

	assertTallyResult := func(yes, no, vito, abstain int64) {
		cachedCtx, _ := suite.ctx.CacheContext()
		_, _, result := suite.app.GovKeeper.Tally(cachedCtx, proposal)
		suite.Require().Equal(sdk.NewInt(yes), result.Yes)
		suite.Require().Equal(sdk.NewInt(no), result.No)
		suite.Require().Equal(sdk.NewInt(vito), result.NoWithVeto)
		suite.Require().Equal(sdk.NewInt(abstain), result.Abstain)
	}

	assertTallyResult(80000000, 10000000, 0, 0)

	delAbToken := sdk.NewInt(40000000)
	delBbToken := sdk.NewInt(80000000)
	delCbToken := sdk.NewInt(60000000)
	delDbToken := sdk.NewInt(20000000)
	delEbToken := sdk.NewInt(80000000)
	delFbToken := sdk.NewInt(120000000)
	newShares, bToken, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delA, sdk.NewCoin(sdk.DefaultBondDenom, delAbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delA, bondedBondDenom).Amount, delAbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, delBbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delB, bondedBondDenom).Amount, delBbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, delCbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delC, bondedBondDenom).Amount, delCbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, delDbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delD, bondedBondDenom).Amount, delDbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, delEbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delE, bondedBondDenom).Amount, delEbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delF, sdk.NewCoin(sdk.DefaultBondDenom, delFbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delF, bondedBondDenom).Amount, delFbToken)

	totalPower := sdk.ZeroInt()
	totalShare := sdk.ZeroDec()
	suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		totalPower = totalPower.Add(validator.GetTokens())
		totalShare = totalShare.Add(validator.GetDelegatorShares())
		return false
	})

	assertTallyResult(240000000, 100000000, 20000000, 120000000)

	// Test TallyLiquidGov
	otherVotes := make(govtypes.OtherVotes)
	testOtherVotes := func(voter sdk.AccAddress, bTokenValue sdk.Int) {
		suite.Require().Len(otherVotes[voter.String()], liquidValidators.Len())
		for _, v := range liquidValidators {
			suite.Require().EqualValues(otherVotes[voter.String()][v.OperatorAddress], bTokenValue.ToDec().QuoInt64(int64(liquidValidators.Len())))
		}
	}
	tallyLiquidGov := func() {
		cachedCtx, _ := suite.ctx.CacheContext()
		otherVotes = make(govtypes.OtherVotes)
		votes := suite.app.GovKeeper.GetVotes(cachedCtx, proposal.ProposalId)
		suite.keeper.TallyLiquidGov(cachedCtx, &votes, &otherVotes)
		squadtypes.PP(otherVotes)

		suite.Require().Len(otherVotes, 5)
		testOtherVotes(delB, delBbToken)
		testOtherVotes(delC, delCbToken)
		testOtherVotes(delD, delDbToken)
		testOtherVotes(delE, delEbToken)
		testOtherVotes(delF, delFbToken)
	}

	tallyLiquidGov()

	// Test balance of PoolTokens including bToken
	pair1 := suite.createPair(delB, params.BondedBondDenom, sdk.DefaultBondDenom, false)
	pool1 := suite.createPool(delB, pair1.Id, sdk.NewCoins(sdk.NewCoin(params.BondedBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	tallyLiquidGov()
	pair2 := suite.createPair(delC, sdk.DefaultBondDenom, params.BondedBondDenom, false)
	pool2 := suite.createPool(delC, pair2.Id, sdk.NewCoins(sdk.NewCoin(params.BondedBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	balance := suite.app.BankKeeper.GetBalance(suite.ctx, delC, pool2.PoolCoinDenom)
	fmt.Println(balance)
	tallyLiquidGov()

	// Test Farming Queued Staking of bToken
	suite.CreateFixedAmountPlan(suite.addrs[0], map[string]string{params.BondedBondDenom: "0.4", pool1.PoolCoinDenom: "0.3", pool2.PoolCoinDenom: "0.3"}, map[string]int64{"testdenom": 1})
	suite.Stake(delD, sdk.NewCoins(sdk.NewCoin(params.BondedBondDenom, sdk.NewInt(10000000))))
	queuedStaking, found := suite.app.FarmingKeeper.GetQueuedStaking(suite.ctx, params.BondedBondDenom, delD)
	suite.True(found)
	suite.Equal(queuedStaking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Staking Position Staking of bToken
	suite.AdvanceEpoch()
	staking, found := suite.app.FarmingKeeper.GetStaking(suite.ctx, params.BondedBondDenom, delD)
	suite.True(found)
	suite.Equal(staking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Queued Staking of PoolTokens including bToken
	suite.Stake(delC, sdk.NewCoins(sdk.NewCoin(pool2.PoolCoinDenom, sdk.NewInt(10000000))))
	queuedStaking, found = suite.app.FarmingKeeper.GetQueuedStaking(suite.ctx, pool2.PoolCoinDenom, delC)
	suite.True(found)
	suite.Equal(queuedStaking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()

	// Test Farming Staking Position of PoolTokens including bToken
	suite.AdvanceEpoch()
	staking, found = suite.app.FarmingKeeper.GetStaking(suite.ctx, pool2.PoolCoinDenom, delC)
	suite.True(found)
	suite.Equal(staking.Amount, sdk.NewInt(10000000))
	tallyLiquidGov()
}

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestLiquidStakingGov2() {
	params := types.DefaultParams()

	vals, valOpers := suite.CreateValidators([]int64{10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	suite.keeper.EndBlocker(suite.ctx)

	val1, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[0])

	delA := suite.addrs[0]
	delB := suite.addrs[1]

	_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delA, sdk.NewInt(50000000), stakingtypes.Unbonded, val1, true)
	suite.Require().NoError(err)

	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tp)
	suite.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delA, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))

	cachedCtx, _ := suite.ctx.CacheContext()
	_, _, result := suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(0), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))
	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(10000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	newShares, bToken, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delB, params.BondedBondDenom).Amount, sdk.NewInt(50000000))

	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(60000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain)))

	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(10000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(50000000), result.Abstain)
}
