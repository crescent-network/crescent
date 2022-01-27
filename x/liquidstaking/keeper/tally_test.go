package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	simapp "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestGetVoterBalanceByDenom() {
	suite.SetupTest()

	voter1, _ := sdk.AccAddressFromBech32("cosmos138w269yyeyj0unge54km8572lgf54l8e3yu8lg")
	voter2, _ := sdk.AccAddressFromBech32("cosmos1u0wfxlachgzqpwnkcwj2vzy025ehzv0qlhujnr")
	//voter3, _ := sdk.AccAddressFromBech32("cosmos14sqkxzdjqwmyclur633wg85sjvvahscgfatvv7")
	//voter4, _ := sdk.AccAddressFromBech32("cosmos1pr7ux292w5ag3v29jzg3gfspw7hufp8l94xejs")

	simapp.InitAccountWithCoins(suite.app, suite.ctx, voter1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))))
	simapp.InitAccountWithCoins(suite.app, suite.ctx, voter2, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000))))
	simapp.InitAccountWithCoins(suite.app, suite.ctx, voter2, sdk.NewCoins(sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(500000))))

	tp := govtypes.NewTextProposal("Test", "description")
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tp)
	suite.Require().NoError(err)

	proposal.Status = govtypes.StatusVotingPeriod
	suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, voter1, govtypes.NewNonSplitVoteOption(govtypes.OptionYes)))
	suite.Require().NoError(suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, voter2, govtypes.NewNonSplitVoteOption(govtypes.OptionNo)))

	votes := suite.app.GovKeeper.GetVotes(suite.ctx, proposal.ProposalId)
	voterBalanceByDenom := suite.keeper.GetVoterBalanceByDenom(suite.ctx, &votes)

	suite.Require().Len(voterBalanceByDenom, 2)
	suite.Require().Len(voterBalanceByDenom[sdk.DefaultBondDenom], 2)
	suite.Require().Len(voterBalanceByDenom[types.DefaultLiquidBondDenom], 1)

	suite.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter1.String()], sdk.NewInt(1000000))
	suite.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter2.String()], sdk.NewInt(1000000))
	suite.Require().EqualValues(voterBalanceByDenom[types.DefaultLiquidBondDenom][voter2.String()], sdk.NewInt(500000))
}

// test Liquid Staking gov power tally
func (suite *KeeperTestSuite) TestTally() {
	//suite.SetupTest()
	//params := types.DefaultParams()
	//params.UnstakeFeeRate = sdk.ZeroDec()
	//suite.keeper.SetParams(suite.ctx, params)
	//
	//vals, valOpers := suite.CreateValidators([]int64{10000000})
	//params.WhitelistedValidators = []types.WhitelistedValidator{
	//	{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(10)},
	//}
	//suite.keeper.SetParams(suite.ctx, params)
	//suite.ctx = suite.ctx.WithBlockHeight(100).WithBlockTime(types.MustParseRFC3339("2022-03-01T00:00:00Z"))
	//liquidstaking.EndBlocker(suite.ctx, suite.keeper)
	//
	//val1, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, valOpers[0])
	//
	//delA := suite.addrs[0]
	//delB := suite.addrs[1]
	//
	//_, err := suite.app.StakingKeeper.Delegate(suite.ctx, delA, sdk.NewInt(50000000), stakingtypes.Unbonded, val1, true)
	//suite.Require().NoError(err)
	//
	//tp := govtypes.NewTextProposal("Test", "description")
	//proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, tp)
	//suite.Require().NoError(err)
	//
	//proposal.Status = govtypes.StatusVotingPeriod
	//suite.app.GovKeeper.SetProposal(suite.ctx, proposal)
	//
	//err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, suite.addrs[0], govtypes.NewNonSplitVoteOption(govtypes.OptionYes))
	//suite.Require().NoError(err)
	//
	////votes := suite.app.GovKeeper.GetVotes(suite.ctx, proposal.ProposalId)
	//
	//cachedCtx, _ := suite.ctx.CacheContext()
	//_, _, result := suite.app.GovKeeper.Tally(cachedCtx, proposal)
	//suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	//suite.Require().Equal(sdk.NewInt(0), result.No)
	//suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	//suite.Require().Equal(sdk.NewInt(0), result.Abstain)
	//
	//err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, vals[0], govtypes.NewNonSplitVoteOption(govtypes.OptionNo))
	//suite.Require().NoError(err)
	//cachedCtx, _ = suite.ctx.CacheContext()
	//_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	//suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	//suite.Require().Equal(sdk.NewInt(10000000), result.No)
	//suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	//suite.Require().Equal(sdk.NewInt(0), result.Abstain)
	//
	//_, _, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
	//suite.Require().NoError(err)
	//
	//cachedCtx, _ = suite.ctx.CacheContext()
	//_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	//suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	//suite.Require().Equal(sdk.NewInt(60000000), result.No)
	//suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	//suite.Require().Equal(sdk.NewInt(0), result.Abstain)
	//
	//err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.ProposalId, delB, govtypes.NewNonSplitVoteOption(govtypes.OptionAbstain))
	//suite.Require().NoError(err)
	//
	//cachedCtx, _ = suite.ctx.CacheContext()
	//_, _, result = suite.app.GovKeeper.Tally(cachedCtx, proposal)
	//suite.Require().Equal(sdk.NewInt(50000000), result.Yes)
	//suite.Require().Equal(sdk.NewInt(10000000), result.No)
	//suite.Require().Equal(sdk.NewInt(0), result.NoWithVeto)
	//suite.Require().Equal(sdk.NewInt(50000000), result.Abstain)
}
