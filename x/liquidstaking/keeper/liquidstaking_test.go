package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/crescent-network/crescent/x/liquidstaking"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
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
	_, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().Error(err)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	newShares, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().NoError(err)
	suite.Require().Equal(newShares, stakingAmt.ToDec())

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
	suite.Require().Equal(stakingAmt.ToDec(), proxyAccDel1.Shares.Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	balanceBeforeUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	suite.Require().Equal(balanceBeforeUBD.Amount, sdk.NewInt(999950000))

	liquidBondDenom := suite.keeper.LiquidBondDenom(suite.ctx)
	ubdAmt := sdk.NewCoin(liquidBondDenom, sdk.NewInt(10000))
	bTokenBalance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], liquidBondDenom)
	bTokenTotalSupply := suite.app.BankKeeper.GetSupply(suite.ctx, liquidBondDenom)
	suite.Require().Equal(bTokenBalance, sdk.NewCoin(liquidBondDenom, sdk.NewInt(50000)))
	suite.Require().Equal(bTokenBalance, bTokenTotalSupply)

	ubdTime, ubds, err := suite.keeper.LiquidUnstaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], ubdAmt)
	suite.Require().NoError(err)
	suite.Require().Len(ubds, 3)
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
	suite.Require().Equal(stakingAmt.Sub(ubdAmt.Amount).ToDec(), proxyAccDel1.Shares.Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	suite.ctx = suite.ctx.WithBlockHeight(200).WithBlockTime(ubdTime.Add(1))
	updates := suite.app.StakingKeeper.BlockValidatorUpdates(suite.ctx) // EndBlock of staking keeper
	suite.Require().Empty(updates)
	balanceCompleteUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	suite.Require().Equal(balanceCompleteUBD.Amount, balanceBeforeUBD.Amount.Add(ubdAmt.Amount))

	proxyAccDel1, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)
	suite.Require().Equal(sdk.MustNewDecFromStr("13333.0"), proxyAccDel1.Shares)
	suite.Require().Equal(sdk.MustNewDecFromStr("13333.0"), proxyAccDel2.Shares)
	suite.Require().Equal(sdk.MustNewDecFromStr("13334.0"), proxyAccDel3.Shares)
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

	lValMap := suite.keeper.GetAllLiquidValidatorsMap(suite.ctx)
	fmt.Println(lValMap)

	//val1, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[0])
	//val2, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[1])
	//val3, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[2])
	val4, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[3])
	//val5, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[0])

	delA := suite.addrs[0]
	delB := suite.addrs[1]
	delC := suite.addrs[2]
	delD := suite.addrs[3]
	delE := suite.addrs[4]
	delF := suite.addrs[5]
	delG := suite.addrs[6]
	//delH := suite.addrs[3]
	//delTokens := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 10)

	_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delG, sdk.NewInt(60000000), stakingtypes.Unbonded, val4, true)
	suite.Require().NoError(err)

	// v5(H, 40) already
	//_, err = suite.app.StakingKeeper.Delegate(suite.ctx, suite.addrs[3], sdk.NewInt(40), stakingtypes.Unbonded, val2, true)
	//suite.Require().NoError(err)

	// 7 addr B, C, D, E, F, G, H
	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tp)
	suite.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[1], govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)
	//suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[2], govtypes.NewNonSplitVoteOption(govtypes.OptionNo))
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[3], govtypes.NewNonSplitVoteOption(govtypes.OptionNo))
	suite.Require().NoError(err)

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionNo))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delC, govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delD, govtypes.NewNonSplitVoteOption(govtypes.OptionNoWithVeto))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delE, govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delF, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain))
	suite.Require().NoError(err)
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delG, govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)

	suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		return false
	})

	cachedCtx, _ := suite.ctx.CacheContext()
	pass, burnDeposit, result := suite.app.GovKeeper.Tally(cachedCtx, proposal)
	pp.Print(pass, burnDeposit, result.String())
	suite.Require().Equal(sdk.NewInt(80000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(10000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delA, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delA, liquidBondDenom), "delA", delA.String())

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(80000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delB, liquidBondDenom), "delB", delB.String())

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(60000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delC, liquidBondDenom), "delC", delC.String())

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delD, liquidBondDenom), "delD", delD.String())

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(80000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delE, liquidBondDenom), "delE", delE.String())

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delF, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(120000000)))
	suite.Require().NoError(err)
	fmt.Println(suite.app.BankKeeper.GetBalance(suite.ctx, delF, liquidBondDenom), "delF", delF.String())

	totalPower := sdk.ZeroInt()
	totalShare := sdk.ZeroDec()
	suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		pp.Println(validator.GetOperator().String(), validator.GetDelegatorShares().String())
		totalPower = totalPower.Add(validator.GetTokens())
		totalShare = totalShare.Add(validator.GetDelegatorShares())
		return false
	})

	fmt.Println(totalPower, totalShare)
	cachedCtx, _ = suite.ctx.CacheContext()
	pass, burnDeposit, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(240000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(100000000), result.No)
	suite.Require().Equal(sdk.NewInt(20000000), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(120000000), result.Abstain)
	pp.Print(pass, burnDeposit, result.String())

	//_, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, suite.delAddrs[0], valOpers[0])
	//suite.Require().False(found)

	//balanceBeforeUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	//suite.Require().Equal(balanceBeforeUBD.Amount, sdk.NewInt(999950000))

	//balanceBeginUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	//suite.Require().Equal(balanceBeginUBD.Amount, balanceBeforeUBD.Amount)
	//
	//proxyAccDel, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	//suite.Require().Equal(proxyAccDel.Shares, stakingAmt.Sub(ubdAmt).ToDec())
	//suite.Require().True(found)
	//
	//suite.ctx = suite.ctx.WithBlockHeight(200).WithBlockTime(ubdTime.Add(1))
	//updates := suite.app.StakingKeeper.BlockValidatorUpdates(suite.ctx) // EndBlock of staking keeper
	//suite.Require().Empty(updates)
	//balanceCompleteUBD := suite.app.BankKeeper.GetBalance(suite.ctx, suite.delAddrs[0], sdk.DefaultBondDenom)
	//suite.Require().Equal(balanceCompleteUBD.Amount, balanceBeforeUBD.Amount.Add(ubdAmt))
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

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delA, govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	suite.Require().NoError(err)

	cachedCtx, _ := suite.ctx.CacheContext()
	_, _, result := suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(0), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionNo))
	suite.Require().NoError(err)
	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(10000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	_, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
	suite.Require().NoError(err)

	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(60000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(0), result.Abstain)

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain))
	suite.Require().NoError(err)

	cachedCtx, _ = suite.ctx.CacheContext()
	_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	suite.Require().Equal(sdk.NewInt(10000000), result.No)
	suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	suite.Require().Equal(sdk.NewInt(50000000), result.Abstain)
}

//func (suite *KeeperTestSuite) TestCollectBiquidStakings() {
//	for _, tc := range []struct {
//		name           string
//		liquidStakings       []types.BiquidStaking
//		epochBlocks    uint32
//		accAsserts     []sdk.AccAddress
//		balanceAsserts []sdk.Coins
//		expectErr      bool
//	}{
//		{
//			"basic liquidStakings case",
//			suite.liquidStakings[:4],
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//			},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"only expired liquidstaking case",
//			[]types.BiquidStaking{suite.liquidStakings[3]},
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[2],
//			},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"source has small balances case",
//			suite.liquidStakings[4:6],
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
//				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
//				mustParseCoinsNormalized("1denom1,1denom3"),
//			},
//			false,
//		},
//		{
//			"none liquidStakings case",
//			nil,
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"disabled liquidstaking epoch",
//			nil,
//			0,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"disabled liquidstaking epoch with liquidStakings",
//			suite.liquidStakings[:4],
//			0,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//	} {
//		suite.Run(tc.name, func() {
//			suite.SetupTest()
//			params := suite.keeper.GetParams(suite.ctx)
//			params.BiquidStakings = tc.liquidStakings
//			params.EpochBlocks = tc.epochBlocks
//			suite.keeper.SetParams(suite.ctx, params)
//
//			err := suite.keeper.CollectBiquidStakings(suite.ctx)
//			if tc.expectErr {
//				suite.Error(err)
//			} else {
//				suite.NoError(err)
//
//				for i, acc := range tc.accAsserts {
//					suite.True(suite.app.BankKeeper.GetAllBalances(suite.ctx, acc).IsEqual(tc.balanceAsserts[i]))
//				}
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestBiquidStakingChangeSituation() {
//	encCfg := app.MakeTestEncodingConfig()
//	params := suite.keeper.GetParams(suite.ctx)
//	suite.keeper.SetParams(suite.ctx, params)
//	height := 1
//	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-01T00:00:00Z"))
//	suite.ctx = suite.ctx.WithBlockHeight(int64(height))
//
//	// cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az
//	// inflation occurs by 1000000000de
//	nom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
//	liquidStakingSource := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "InflationPool")
//
//	for _, tc := range []struct {
//		name                    string
//		proposal                *proposal.ParameterChangeProposal
//		liquidStakingCount            int
//		collectibleBiquidStakingCount int
//		govTime                 time.Time
//		nextBlockTime           time.Time
//		expErr                  error
//		accAsserts              []sdk.AccAddress
//		balanceAsserts          []sdk.Coins
//	}{
//		{
//			"add liquidstaking 1",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					}
//				]`,
//			}),
//			1,
//			0,
//			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//				{},
//				{},
//			},
//		},
//		{
//			"add liquidstaking 2",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			2,
//			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3 with invalid total rate case 1",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2, // left last liquidStakings of 2nd tc
//			1, // left last liquidStakings of 2nd tc
//			types.MustParseRFC3339("2021-09-29T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-30T00:00:00Z"),
//			types.ErrInvalidTotalBiquidStakingRate,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3 with invalid total rate case 2",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2, // left last liquidStakings of 2nd tc
//			1, // left last liquidStakings of 2nd tc
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.ErrInvalidTotalBiquidStakingRate,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("750000000denom1,750000000denom2,750000000denom3,750000000stake"),
//				mustParseCoinsNormalized("2250000000denom1,2250000000denom2,2250000000denom3,2250000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			2,
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("3125000000denom1,3125000000denom2,3125000000denom3,3125000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//		{
//			"add liquidstaking 4 without date range overlap",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-4",
//					"rate": "1.000000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2031-09-30T00:00:01Z",
//					"end_time": "2031-12-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			1,
//			types.MustParseRFC3339("2021-09-29T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-30T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//		{
//			"remove all liquidStakings",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value:    `[]`,
//			}),
//			0,
//			0,
//			types.MustParseRFC3339("2021-10-25T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-26T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
//				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//	} {
//		suite.Run(tc.name, func() {
//			proposalJson := paramscutils.ParamChangeProposalJSON{}
//			bz, err := tc.proposal.Marshal()
//			suite.Require().NoError(err)
//			err = encCfg.Amino.Unmarshal(bz, &proposalJson)
//			suite.Require().NoError(err)
//			proposal := paramproposal.NewParameterChangeProposal(
//				proposalJson.Title, proposalJson.Description, proposalJson.Changes.ToParamChanges(),
//			)
//			suite.Require().NoError(err)
//
//			// endblock gov paramchange ->(new block)-> beginblock liquidstaking -> mempool -> endblock gov paramchange ->(new block)-> ...
//			suite.ctx = suite.ctx.WithBlockTime(tc.govTime)
//			err = suite.govHandler(suite.ctx, proposal)
//			if tc.expErr != nil {
//				suite.Require().Error(err)
//			} else {
//				suite.Require().NoError(err)
//			}
//
//			// (new block)
//			height += 1
//			suite.ctx = suite.ctx.WithBlockHeight(int64(height))
//			suite.ctx = suite.ctx.WithBlockTime(tc.nextBlockTime)
//
//			params := suite.keeper.GetParams(suite.ctx)
//			suite.Require().Len(params.BiquidStakings, tc.liquidStakingCount)
//			for _, liquidStaking := range params.BiquidStakings {
//				err := liquidStaking.Validate()
//				suite.Require().NoError(err)
//			}
//
//			liquidStakings := types.CollectibleBiquidStakings(params.BiquidStakings, suite.ctx.BlockTime())
//			suite.Require().Len(liquidStakings, tc.collectibleBiquidStakingCount)
//
//			// BeginBlocker - inflation or mint on liquidStakingSource
//			// inflation occurs by 1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
//			err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, liquidStakingSource, initialBalances)
//			suite.Require().NoError(err)
//
//			// BeginBlocker - Collect liquidStakings
//			err = suite.keeper.CollectBiquidStakings(suite.ctx)
//			suite.Require().NoError(err)
//
//			// Assert liquidstaking collections
//			for i, acc := range tc.accAsserts {
//				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc)
//				suite.Require().Equal(tc.balanceAsserts[i], balances)
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestGetSetTotalCollectedCoins() {
//	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().Nil(collectedCoins)
//
//	suite.keeper.SetTotalCollectedCoins(suite.ctx, "liquidStaking1", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))
//
//	suite.keeper.AddTotalCollectedCoins(suite.ctx, "liquidStaking1", sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)), collectedCoins))
//
//	suite.keeper.AddTotalCollectedCoins(suite.ctx, "liquidStaking2", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking2")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))
//}
//
//func (suite *KeeperTestSuite) TestTotalCollectedCoins() {
//	liquidStaking := types.BiquidStaking{
//		Name:               "liquidStaking1",
//		Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//		SourceAddress:      suite.sourceAddrs[0].String(),
//		DestinationAddress: suite.destinationAddrs[0].String(),
//		StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//		EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//	}
//
//	params := suite.keeper.GetParams(suite.ctx)
//	params.BiquidStakings = []types.BiquidStaking{liquidStaking}
//	suite.keeper.SetParams(suite.ctx, params)
//
//	balance := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.sourceAddrs[0])
//	expectedCoins, _ := sdk.NewDecCoinsFromCoins(balance...).MulDec(sdk.NewDecWithPrec(5, 2)).TruncateDecimal()
//
//	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().Equal(sdk.Coins(nil), collectedCoins)
//
//	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-31T00:00:00Z"))
//	err := suite.keeper.CollectBiquidStakings(suite.ctx)
//	suite.Require().NoError(err)
//
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(expectedCoins, collectedCoins))
//}
