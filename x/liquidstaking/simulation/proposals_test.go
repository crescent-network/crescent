package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/crescent-network/crescent/v3/app/params"
	"github.com/crescent-network/crescent/v3/x/liquidstaking/simulation"
	"github.com/crescent-network/crescent/v3/x/liquidstaking/types"
)

func TestProposalContents(t *testing.T) {
	app, ctx := createTestApp(false)

	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 10)

	// setup accounts[0] as validator0 and accounts[1] as validator1
	val0 := getTestingValidator0(t, app, ctx, accounts)
	val1 := getTestingValidator1(t, app, ctx, accounts)

	param := app.LiquidStakingKeeper.GetParams(ctx)
	param.WhitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: val0.OperatorAddress,
			TargetWeight:     sdk.OneInt(),
		},
		{
			ValidatorAddress: val1.OperatorAddress,
			TargetWeight:     sdk.OneInt(),
		},
	}
	app.LiquidStakingKeeper.SetParams(ctx, param)

	// begin a new block
	blockTime := time.Now().UTC()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})
	app.EndBlock(abci.RequestEndBlock{Height: app.LastBlockHeight() + 1})

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GovKeeper, app.LiquidStakingKeeper)
	require.Len(t, weightedProposalContent, 5)

	w0 := weightedProposalContent[0]
	w1 := weightedProposalContent[1]
	w2 := weightedProposalContent[2]
	w3 := weightedProposalContent[3]
	w4 := weightedProposalContent[4]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSimulateAddWhitelistValidatorsProposal, w0.AppParamsKey())
	require.Equal(t, params.DefaultWeightAddWhitelistValidatorsProposal, w0.DefaultWeight())

	// tests w1 interface:
	require.Equal(t, simulation.OpWeightSimulateUpdateWhitelistValidatorsProposal, w1.AppParamsKey())
	require.Equal(t, params.DefaultWeightUpdateWhitelistValidatorsProposal, w1.DefaultWeight())

	// tests w2 interface:
	require.Equal(t, simulation.OpWeightSimulateDeleteWhitelistValidatorsProposal, w2.AppParamsKey())
	require.Equal(t, params.DefaultWeightDeleteWhitelistValidatorsProposal, w2.DefaultWeight())

	// tests w3 interface:
	require.Equal(t, simulation.OpWeightCompleteRedelegationUnbonding, w3.AppParamsKey())
	require.Equal(t, params.DefaultWeightCompleteRedelegationUnbonding, w3.DefaultWeight())

	// tests w4 interface:
	require.Equal(t, simulation.OpWeightTallyWithLiquidStaking, w4.AppParamsKey())
	require.Equal(t, params.DefaultWeightTallyWithLiquidStaking, w4.DefaultWeight())

	content0 := w0.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content0)

	content1 := w1.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content1)

	content2 := w2.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content2)

	content3 := w3.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content3)

	content4 := w4.ContentSimulatorFn()(r, ctx, accounts)
	require.Nil(t, content4)
}
