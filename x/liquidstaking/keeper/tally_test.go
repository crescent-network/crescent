package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// test Liquid Staking gov power
func (suite *KeeperTestSuite) TestLiquidStakingGov3() {
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

	_, _, err = suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, delB, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50000000)))
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
