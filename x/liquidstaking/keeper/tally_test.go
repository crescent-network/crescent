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
	simapp.InitAccountWithCoins(suite.app, suite.ctx, voter2, sdk.NewCoins(sdk.NewCoin(types.DefaultBondedBondDenom, sdk.NewInt(500000))))

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
	suite.Require().Len(voterBalanceByDenom[types.DefaultBondedBondDenom], 1)

	suite.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter1.String()], sdk.NewInt(1000000))
	suite.Require().EqualValues(voterBalanceByDenom[sdk.DefaultBondDenom][voter2.String()], sdk.NewInt(1000000))
	suite.Require().EqualValues(voterBalanceByDenom[types.DefaultBondedBondDenom][voter2.String()], sdk.NewInt(500000))
}
