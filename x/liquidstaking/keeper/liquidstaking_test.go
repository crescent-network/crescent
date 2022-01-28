package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/k0kubun/pp"
)

// tests LiquidStaking, LiquidUnstaking
func (suite *KeeperTestSuite) TestLiquidStaking() {
	_, valOpers := suite.CreateValidators([]int64{1000000, 2000000, 3000000})
	suite.ctx = suite.ctx.WithBlockHeight(100).WithBlockTime(types.MustParseRFC3339("2022-03-01T00:00:00Z"))
	params := suite.keeper.GetParams(suite.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	stakingAmt := sdk.NewInt(50000)

	// fail, no active validator
	newShares, bTokenMintAmt, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().Error(err)
	suite.Require().Equal(newShares, sdk.ZeroDec())
	suite.Require().Equal(bTokenMintAmt, sdk.Int{})

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	activeVals := suite.keeper.GetActiveLiquidValidators(suite.ctx)
	_, crumb := types.DivideByWeight(activeVals, stakingAmt)
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

	liquidBondDenom := suite.keeper.LiquidBondDenom(suite.ctx)
	ubdAmt := sdk.NewCoin(liquidBondDenom, sdk.NewInt(10000))
	bTokenBalance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], liquidBondDenom)
	bTokenTotalSupply := suite.app.BankKeeper.GetSupply(suite.ctx, liquidBondDenom)
	suite.Require().Equal(bTokenBalance, sdk.NewCoin(liquidBondDenom, sdk.NewInt(50000)))
	suite.Require().Equal(bTokenBalance, bTokenTotalSupply)

	ubdTime, unbondingAmt, ubds, err := suite.keeper.LiquidUnstaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], ubdAmt)
	suite.Require().NoError(err)
	suite.Require().Len(ubds, 3)
	// truncated shares
	suite.Require().EqualValues(unbondingAmt, ubdAmt.Amount.QuoRaw(3).MulRaw(3).ToDec())
	//suite.Require().Equal(unbondingAmt, ubdAmt.Amount.ToDec())
	suite.Require().Equal(ubds[0].DelegatorAddress, suite.delAddrs[0].String())
	suite.Require().Equal(ubdTime, types.MustParseRFC3339("2022-03-22T00:00:00Z"))
	bTokenBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], liquidBondDenom)
	suite.Require().Equal(bTokenBalanceAfter, sdk.NewCoin(liquidBondDenom, sdk.NewInt(40000)))

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
	// TODO: add cases for different weight
}

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestLiquidStakingGov() {
	suite.SetupTest()
	params := types.DefaultParams()
	params.UnstakeFeeRate = sdk.ZeroDec()
	liquidBondDenom := suite.keeper.LiquidBondDenom(suite.ctx)

	// v1, v2, v3, v4
	vals, valOpers := suite.CreateValidators([]int64{10000000, 10000000, 10000000, 10000000, 10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[3].String(), Weight: sdk.NewInt(10)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	suite.ctx = suite.ctx.WithBlockHeight(100).WithBlockTime(types.MustParseRFC3339("2022-03-01T00:00:00Z"))
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

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
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delA, liquidBondDenom).Amount, delAbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, delBbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delB, liquidBondDenom).Amount, delBbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, delCbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delC, liquidBondDenom).Amount, delCbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, delDbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delD, liquidBondDenom).Amount, delDbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, delEbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delE, liquidBondDenom).Amount, delEbToken)

	newShares, bToken, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delF, sdk.NewCoin(sdk.DefaultBondDenom, delFbToken))
	suite.Require().NoError(err)
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delF, liquidBondDenom).Amount, delFbToken)

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
	cachedCtx, _ := suite.ctx.CacheContext()
	votes := suite.app.GovKeeper.GetVotes(cachedCtx, proposal.ProposalId)
	otherVotes := make(govtypes.OtherVotes)
	suite.keeper.TallyLiquidGov(cachedCtx, &votes, &otherVotes)
	squadtypes.PP(votes)
	squadtypes.PP(otherVotes)

	testOtherVotes := func(voter sdk.AccAddress, bTokenValue sdk.Int) {
		suite.Require().Len(otherVotes[voter.String()], liquidValidators.Len())
		for _, v := range liquidValidators {
			suite.Require().EqualValues(otherVotes[voter.String()][v.OperatorAddress], bTokenValue.ToDec().QuoInt64(int64(liquidValidators.Len())))
		}
	}

	suite.Require().Len(otherVotes, 5)
	testOtherVotes(delB, delBbToken)
	testOtherVotes(delC, delCbToken)
	testOtherVotes(delD, delDbToken)
	testOtherVotes(delE, delEbToken)
	testOtherVotes(delF, delFbToken)

	// TODO: add voter, btoken, farming of btoken, pool, farming of pool
}

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestLiquidStakingGov2() {
	suite.SetupTest()
	params := types.DefaultParams()
	params.UnstakeFeeRate = sdk.ZeroDec()
	suite.keeper.SetParams(suite.ctx, params)

	vals, valOpers := suite.CreateValidators([]int64{10000000})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(10)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	suite.ctx = suite.ctx.WithBlockHeight(100).WithBlockTime(types.MustParseRFC3339("2022-03-01T00:00:00Z"))
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

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
	suite.Require().EqualValues(newShares.TruncateInt(), bToken, suite.app.BankKeeper.GetBalance(suite.ctx, delB, params.LiquidBondDenom).Amount, sdk.NewInt(50000000))

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
