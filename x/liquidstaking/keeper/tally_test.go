package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	chain "github.com/crescent-network/crescent/v4/app"
	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/liquidstaking/types"
	lpfarmtypes "github.com/crescent-network/crescent/v4/x/lpfarm/types"
)

// test Liquid Staking gov power
func (s *KeeperTestSuite) TestGetVoterBalanceByDenom() {
	voter1, _ := sdk.AccAddressFromBech32("cosmos138w269yyeyj0unge54km8572lgf54l8e3yu8lg")
	voter2, _ := sdk.AccAddressFromBech32("cosmos1u0wfxlachgzqpwnkcwj2vzy025ehzv0qlhujnr")

	chain.InitAccountWithCoins(s.app, s.ctx, voter1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))))
	chain.InitAccountWithCoins(s.app, s.ctx, voter2, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))))
	chain.InitAccountWithCoins(s.app, s.ctx, voter2, sdk.NewCoins(sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(500000))))

	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, tp)
	s.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	s.app.GovKeeper.SetProposal(s.ctx, proposal)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, voter1, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, voter2, govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))

	votes := s.app.GovKeeper.GetVotes(s.ctx, proposal.ProposalId)
	voterBalanceByDenom := s.keeper.GetVoterBalanceByDenom(s.ctx, votes)

	s.Require().Len(voterBalanceByDenom, 2)
	s.Require().Len(voterBalanceByDenom[sdk.DefaultBondDenom], 2)
	s.Require().Len(voterBalanceByDenom[types.DefaultLiquidBondDenom], 1)

	s.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter1.String()], sdk.NewInt(1000000))
	s.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter2.String()], sdk.NewInt(1000000))
	s.Require().EqualValues(voterBalanceByDenom[types.DefaultLiquidBondDenom][voter2.String()], sdk.NewInt(500000))
}

// test Liquid Staking gov voting power
func (s *KeeperTestSuite) TestSetLiquidStakingVotingPowers() {
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

	delA := s.addrs[0] // no vote
	delB := s.addrs[1] // vote, Balance of bToken case
	delC := s.addrs[2] // vote, Balance of PoolCoins including bToken, Farming position of PoolCoins that include bToken
	delD := s.addrs[3] // vote, Farming position of bToken case
	delE := s.addrs[4] // vote, Balance of bToken case
	delF := s.addrs[5] // vote, Balance of bToken case
	delG := s.addrs[6] // vote, Balance of bToken case

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

	s.assertTallyResult(80000000, 10000000, 0, 0, proposal)

	delAbToken := sdk.NewInt(40000000)
	delBbToken := sdk.NewInt(80000000)
	delCbToken := sdk.NewInt(60000000)
	delDbToken := sdk.NewInt(20000000)
	delEbToken := sdk.NewInt(80000000)
	delFbToken := sdk.NewInt(120000000)
	newShares, bToken, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delA, sdk.NewCoin(sdk.DefaultBondDenom, delAbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delA, liquidBondDenom).Amount, delAbToken)

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, delBbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delB, liquidBondDenom).Amount, delBbToken)

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, delCbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delC, liquidBondDenom).Amount, delCbToken)

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, delDbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delD, liquidBondDenom).Amount, delDbToken)

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, delEbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delE, liquidBondDenom).Amount, delEbToken)

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delF, sdk.NewCoin(sdk.DefaultBondDenom, delFbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delF, liquidBondDenom).Amount, delFbToken)

	totalPower := sdk.ZeroInt()
	totalShare := sdk.ZeroDec()
	s.app.StakingKeeper.IterateBondedValidatorsByPower(s.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		totalPower = totalPower.Add(validator.GetTokens())
		totalShare = totalShare.Add(validator.GetDelegatorShares())
		return false
	})

	s.assertTallyResult(240000000, 100000000, 20000000, 120000000, proposal)

	// Test SetLiquidStakingVotingPowers
	votingPowers := make(govtypes.AdditionalVotingPowers)
	testVotingPowers := func(voter sdk.AccAddress, bTokenValue sdk.Int) {
		s.Require().Len(votingPowers[voter.String()], liquidValidators.Len())
		totalVotingPower := sdk.ZeroDec()
		for _, v := range liquidValidators {
			votingPower := votingPowers[voter.String()][v.OperatorAddress]
			totalVotingPower = totalVotingPower.Add(votingPower)
			// equal when all liquid validator has same currentWeight
			s.Require().EqualValues(votingPower, bTokenValue.ToDec().QuoInt64(int64(liquidValidators.Len())))
		}
		s.Require().EqualValues(totalVotingPower.TruncateInt(), s.keeper.CalcLiquidStakingVotingPower(s.ctx, voter))
	}
	setLiquidStakingVotingPowers := func() {
		cachedCtx, _ := s.ctx.CacheContext()
		votingPowers = govtypes.AdditionalVotingPowers{}
		votes := s.app.GovKeeper.GetVotes(cachedCtx, proposal.ProposalId)
		s.keeper.SetLiquidStakingVotingPowers(cachedCtx, votes, &votingPowers)
		utils.PP(votingPowers)

		s.Require().Len(votingPowers, 5)
		testVotingPowers(delB, delBbToken)
		testVotingPowers(delC, delCbToken)
		testVotingPowers(delD, delDbToken)
		testVotingPowers(delE, delEbToken)
		testVotingPowers(delF, delFbToken)
		s.assertTallyResult(240000000, 100000000, 20000000, 120000000, proposal)
	}
	setLiquidStakingVotingPowers()

	// Test balance of PoolTokens including bToken
	pair1 := s.createPair(delB, params.LiquidBondDenom, sdk.DefaultBondDenom, false)
	pool1 := s.createPool(delB, pair1.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	setLiquidStakingVotingPowers()
	pair2 := s.createPair(delC, sdk.DefaultBondDenom, params.LiquidBondDenom, false)
	pool2 := s.createPool(delC, pair2.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	balance := s.app.BankKeeper.GetBalance(s.ctx, delC, pool2.PoolCoinDenom)
	fmt.Println(balance)
	setLiquidStakingVotingPowers()

	// Test Farming Queued Staking of bToken
	s.createPublicPlan(
		s.addrs[0], []lpfarmtypes.RewardAllocation{
			lpfarmtypes.NewDenomRewardAllocation(params.LiquidBondDenom, utils.ParseCoins("400000stake")),
			lpfarmtypes.NewDenomRewardAllocation(pool1.PoolCoinDenom, utils.ParseCoins("300000stake")),
			lpfarmtypes.NewDenomRewardAllocation(pool2.PoolCoinDenom, utils.ParseCoins("300000stake")),
		})
	s.farm(delD, sdk.NewInt64Coin(params.LiquidBondDenom, 10000000))
	setLiquidStakingVotingPowers()

	// Test Farming Staking Position Staking of bToken
	s.advanceEpochDays()
	setLiquidStakingVotingPowers()

	// Test Farming Queued Staking of bToken
	s.farm(delD, sdk.NewInt64Coin(params.LiquidBondDenom, 10000000))
	setLiquidStakingVotingPowers()

	// Test Farming Staking Position Staking of bToken
	s.advanceEpochDays()
	setLiquidStakingVotingPowers()

	// Test Farming Queued Staking of PoolTokens including bToken
	s.farm(delC, sdk.NewInt64Coin(pool2.PoolCoinDenom, 500000000000))
	setLiquidStakingVotingPowers()

	// Test Farming Staking Position of PoolTokens including bToken
	s.advanceEpochDays()
	setLiquidStakingVotingPowers()

	// Test Farming Queued Staking of PoolTokens including bToken
	s.farm(delC, sdk.NewInt64Coin(pool2.PoolCoinDenom, 500000000000))
	setLiquidStakingVotingPowers()

	// Test Farming Staking Position of PoolTokens including bToken
	s.advanceEpochDays()
	setLiquidStakingVotingPowers()
}

// test Liquid Staking gov voting power for the address
func (s *KeeperTestSuite) TestGetVotingPower() {
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

	val4, _ := s.app.StakingKeeper.GetValidator(s.ctx, valOpers[3])

	delA := s.addrs[0] // zero power case
	delB := s.addrs[1] // Balance of bToken case
	delC := s.addrs[2] // Balance of PoolCoins including bToken, Farming position of PoolCoins that include bToken
	delD := s.addrs[3] // Farming position of bToken case
	delE := s.addrs[6] // normal staking case, balance of bToken case

	s.assertVotingPower(delA, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt())
	s.assertVotingPower(delE, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt())
	s.assertVotingPower(vals[3], sdk.ZeroInt(), sdk.ZeroInt(), sdk.NewInt(10000000))

	_, err := s.app.StakingKeeper.Delegate(s.ctx, delE, sdk.NewInt(60000000), stakingtypes.Unbonded, val4, true)
	s.Require().NoError(err)

	s.assertVotingPower(delE, sdk.NewInt(60000000), sdk.ZeroInt(), sdk.ZeroInt())
	s.assertVotingPower(vals[3], sdk.ZeroInt(), sdk.ZeroInt(), sdk.NewInt(70000000))

	delBbToken := sdk.NewInt(80000000)
	delCbToken := sdk.NewInt(60000000)
	delDbToken := sdk.NewInt(20000000)
	delEbToken := sdk.NewInt(80000000)

	newShares, bToken, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, delBbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delB, liquidBondDenom).Amount, delBbToken)
	s.assertVotingPower(delB, sdk.ZeroInt(), delBbToken, sdk.ZeroInt())

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delC, sdk.NewCoin(sdk.DefaultBondDenom, delCbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delC, liquidBondDenom).Amount, delCbToken)
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delD, sdk.NewCoin(sdk.DefaultBondDenom, delDbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delD, liquidBondDenom).Amount, delDbToken)
	s.assertVotingPower(delD, sdk.ZeroInt(), delDbToken, sdk.ZeroInt())

	newShares, bToken, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delE, sdk.NewCoin(sdk.DefaultBondDenom, delEbToken))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delE, liquidBondDenom).Amount, delEbToken)
	s.assertVotingPower(delE, sdk.NewInt(60000000), delEbToken, sdk.ZeroInt())

	s.assertVotingPower(vals[3], sdk.ZeroInt(), sdk.ZeroInt(), sdk.NewInt(130000000)) // self bonding 10000000 + normal staking 60000000 + liquid staking 240000000/4

	// Test balance of PoolTokens including bToken
	pair1 := s.createPair(delB, params.LiquidBondDenom, sdk.DefaultBondDenom, false)
	pool1 := s.createPool(delB, pair1.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(40000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	s.assertVotingPower(delB, sdk.ZeroInt(), delBbToken, sdk.ZeroInt())

	pair2 := s.createPair(delC, sdk.DefaultBondDenom, params.LiquidBondDenom, false)
	pool2 := s.createPool(delC, pair2.Id, sdk.NewCoins(sdk.NewCoin(params.LiquidBondDenom, sdk.NewInt(60000000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(44000000))), false)
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Queued Staking of bToken
	s.createPublicPlan(
		s.addrs[0],
		[]lpfarmtypes.RewardAllocation{
			lpfarmtypes.NewDenomRewardAllocation(params.LiquidBondDenom, utils.ParseCoins("400000stake")),
			lpfarmtypes.NewDenomRewardAllocation(pool1.PoolCoinDenom, utils.ParseCoins("300000stake")),
			lpfarmtypes.NewDenomRewardAllocation(pool2.PoolCoinDenom, utils.ParseCoins("300000stake")),
		})
	s.farm(delD, sdk.NewInt64Coin(params.LiquidBondDenom, 10000000))
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Staking Position Staking of bToken without balance
	s.advanceEpochDays()
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Queued Staking of bToken
	s.farm(delD, sdk.NewInt64Coin(params.LiquidBondDenom, 10000000))
	s.assertVotingPower(delD, sdk.ZeroInt(), delDbToken, sdk.ZeroInt())

	// Test Farming Staking Position Staking of bToken without balance
	s.advanceEpochDays()
	s.assertVotingPower(delD, sdk.ZeroInt(), delDbToken, sdk.ZeroInt())

	// Test Farming Queued Staking of PoolTokens including bToken
	s.farm(delC, sdk.NewInt64Coin(pool2.PoolCoinDenom, 500000000000))
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Staking Position of PoolTokens including bToken
	s.advanceEpochDays()
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Queued Staking of PoolTokens including bToken without balance
	s.farm(delC, sdk.NewInt64Coin(pool2.PoolCoinDenom, 500000000000))
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	// Test Farming Staking Position of PoolTokens including bToken
	s.advanceEpochDays()
	s.assertVotingPower(delC, sdk.ZeroInt(), delCbToken, sdk.ZeroInt())

	s.assertVotingPower(delA, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt())
	s.assertVotingPower(delE, sdk.NewInt(60000000), delEbToken, sdk.ZeroInt())

	// send bToken of E to vals[3]
	err = s.app.BankKeeper.SendCoins(s.ctx, delE, vals[3], sdk.NewCoins(sdk.NewCoin(liquidBondDenom, delEbToken)))
	s.Require().NoError(err)
	s.assertVotingPower(delE, sdk.NewInt(60000000), sdk.ZeroInt(), sdk.ZeroInt())
	s.assertVotingPower(vals[3], sdk.ZeroInt(), delEbToken, sdk.NewInt(130000000)) // self bonding 10000000 + normal staking 60000000 + liquid staking 240000000/4
}

// test Liquid Staking gov power
func (s *KeeperTestSuite) TestTallyLiquidStakingGov2() {
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

	newShares, bToken, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
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

// TestVotingPower tests voting power of staking, liquid staking
func (s *KeeperTestSuite) TestVotingPower() {
	params := types.DefaultParams()

	selfDelegationAmount := sdk.NewInt(10000000)
	normalStakingAmount := sdk.NewInt(50000000)
	liquidStakingAmount := sdk.NewInt(50000000)

	vals, valOpers, pks := s.CreateValidators([]int64{selfDelegationAmount.Int64(), selfDelegationAmount.Int64()})
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(5)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	val1, _ := s.app.StakingKeeper.GetValidator(s.ctx, valOpers[0])

	delA := s.addrs[0]

	_, err := s.app.StakingKeeper.Delegate(s.ctx, delA, normalStakingAmount, stakingtypes.Unbonded, val1, true)
	s.Require().NoError(err)

	// normal staking voting power
	svp := s.keeper.CalcStakingVotingPower(s.ctx, delA)
	s.Require().EqualValues(svp, normalStakingAmount)

	// no liquid staking voting power
	lsvp := s.keeper.CalcLiquidStakingVotingPower(s.ctx, delA)
	s.Require().EqualValues(lsvp, sdk.ZeroInt())

	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, tp)
	s.Require().NoError(err)
	proposal.Status = govtypes.StatusVotingPeriod
	s.app.GovKeeper.SetProposal(s.ctx, proposal)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, delA, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))

	cachedCtx, _ := s.ctx.CacheContext()
	_, _, result := s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes) // normalStakingAmount
	s.Require().Equal(sdk.NewInt(0), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	s.Require().NoError(s.app.GovKeeper.AddVote(s.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))
	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(50000000), result.Yes) // normalStakingAmount
	s.Require().Equal(sdk.NewInt(10000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	newShares, bToken, err := s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, delA, sdk.NewCoin(sdk.DefaultBondDenom, liquidStakingAmount))
	s.Require().NoError(err)
	s.Require().EqualValues(newShares.TruncateInt(), bToken, s.app.BankKeeper.GetBalance(s.ctx, delA, params.LiquidBondDenom).Amount, liquidStakingAmount)

	// normal staking voting power
	votingPower := s.keeper.GetVotingPower(s.ctx, delA)
	s.Require().EqualValues(votingPower.StakingVotingPower, normalStakingAmount)
	// liquid staking voting power
	s.Require().EqualValues(votingPower.LiquidStakingVotingPower, liquidStakingAmount)

	// voting power of valoper
	votingPowerValOper1 := s.keeper.GetVotingPower(s.ctx, vals[0])
	votingPowerValOper2 := s.keeper.GetVotingPower(s.ctx, vals[1])
	liquidValStates := s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().EqualValues(votingPowerValOper1.ValidatorVotingPower, selfDelegationAmount.Add(normalStakingAmount).Add(liquidValStates[0].LiquidTokens))
	s.Require().EqualValues(votingPowerValOper2.ValidatorVotingPower, selfDelegationAmount.Add(liquidValStates[1].LiquidTokens))

	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(votingPower.StakingVotingPower.Add(votingPower.LiquidStakingVotingPower), result.Yes)
	s.Require().Equal(sdk.NewInt(100000000), result.Yes) // normalStakingAmount + liquidStakingAmount
	s.Require().Equal(sdk.NewInt(10000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	// double sign second liquid validator
	s.doubleSign(valOpers[1], sdk.ConsAddress(pks[1].Address()))

	// reduced liquid staking voting power because of unbonded liquid validator by double sign
	votingPower = s.keeper.GetVotingPower(s.ctx, delA)
	s.Require().EqualValues(votingPower.StakingVotingPower, normalStakingAmount)
	s.Require().EqualValues(votingPower.LiquidStakingVotingPower, sdk.NewInt(33333334))

	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(votingPower.StakingVotingPower.Add(votingPower.LiquidStakingVotingPower), result.Yes)
	s.Require().Equal(sdk.NewInt(83333334), result.Yes) // staking voting power + reduced liquid staking voting power
	s.Require().Equal(sdk.NewInt(10000000), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)

	// rebalancing for non-active liquid validator by double sign, voting power don't need to wait unbonding period when rebalancing, redelegation
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// recovered liquid staking voting power because of rebalancing the liquid validator except slashing amount
	votingPower = s.keeper.GetVotingPower(s.ctx, delA)
	s.Require().EqualValues(votingPower.StakingVotingPower, normalStakingAmount)
	s.Require().EqualValues(votingPower.LiquidStakingVotingPower, sdk.NewInt(49187500))

	// double sign first liquid validator
	s.doubleSign(valOpers[0], sdk.ConsAddress(pks[0].Address()))

	// normal, liquid staking voting power is zero because of unbonded all validators by double signing
	votingPower = s.keeper.GetVotingPower(s.ctx, delA)
	s.Require().EqualValues(votingPower.StakingVotingPower, sdk.ZeroInt())
	s.Require().EqualValues(votingPower.LiquidStakingVotingPower, sdk.ZeroInt())
	utils.PP(s.keeper.GetNetAmountState(s.ctx))
	utils.PP(s.keeper.GetAllLiquidValidatorStates(s.ctx))

	// rebalancing not occurred because no active liquid validators, unbonding started all liquid tokens, no liquid staking voting power
	s.keeper.UpdateLiquidValidatorSet(s.ctx)
	votingPower = s.keeper.GetVotingPower(s.ctx, delA)
	s.Require().EqualValues(votingPower.StakingVotingPower, sdk.ZeroInt())
	s.Require().EqualValues(votingPower.LiquidStakingVotingPower, sdk.ZeroInt())

	cachedCtx, _ = s.ctx.CacheContext()
	_, _, result = s.app.GovKeeper.Tally(cachedCtx, proposal)
	s.Require().Equal(sdk.NewInt(0), result.Yes)
	s.Require().Equal(sdk.NewInt(0), result.No)
	s.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	s.Require().Equal(sdk.NewInt(0), result.Abstain)
}
