package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/app/params"
	"github.com/crescent-network/crescent/v2/x/farming/simulation"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func TestProposalContents(t *testing.T) {
	app, ctx := createTestApp(false)

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	require.Len(t, weightedProposalContent, 3)

	w0 := weightedProposalContent[0]
	w1 := weightedProposalContent[1]
	w2 := weightedProposalContent[2]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSimulateAddPublicPlanProposal, w0.AppParamsKey())
	require.Equal(t, params.DefaultWeightAddPublicPlanProposal, w0.DefaultWeight())

	// tests w1 interface:
	require.Equal(t, simulation.OpWeightSimulateUpdatePublicPlanProposal, w1.AppParamsKey())
	require.Equal(t, params.DefaultWeightUpdatePublicPlanProposal, w1.DefaultWeight())

	// tests w2 interface:
	require.Equal(t, simulation.OpWeightSimulateDeletePublicPlanProposal, w2.AppParamsKey())
	require.Equal(t, params.DefaultWeightDeletePublicPlanProposal, w2.DefaultWeight())

	content0 := w0.ContentSimulatorFn()(r, ctx, accounts)
	require.Equal(t, "eOcbWwNbeH", content0.GetTitle())
	require.Equal(t, "AjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjOygZsTxPjfweXhSUkMhPjMaxKlMIJMO", content0.GetDescription())
	require.Equal(t, "farming", content0.ProposalRoute())
	require.Equal(t, "PublicPlan", content0.ProposalType())

	// setup public fixed amount plan
	msgPlan := &types.MsgCreateFixedAmountPlan{
		Name:    "simulation",
		Creator: accounts[0].Address.String(),
		StakingCoinWeights: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10, 1)), // 100%
		),
		StartTime:   types.ParseTime("2021-08-01T00:00:00Z"),
		EndTime:     types.ParseTime("2021-08-31T00:00:00Z"),
		EpochAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200_000_000)),
	}

	_, err := app.FarmingKeeper.CreateFixedAmountPlan(
		ctx,
		msgPlan,
		accounts[0].Address,
		accounts[0].Address,
		types.PlanTypePublic,
	)
	require.NoError(t, err)

	content1 := w1.ContentSimulatorFn()(r, ctx, accounts)
	require.Equal(t, "OoMioXHRuF", content1.GetTitle())
	require.Equal(t, "REqrXZSGLqwTMcxHfWotDllNkIJPMbXzjDVjPOOjCFuIvTyhXKLyhUScOXvYthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZ", content1.GetDescription())
	require.Equal(t, "farming", content1.ProposalRoute())
	require.Equal(t, "PublicPlan", content1.ProposalType())

	content2 := w2.ContentSimulatorFn()(r, ctx, accounts)
	require.Equal(t, "wQMUgFFSKt", content2.GetTitle())
	require.Equal(t, "MwMANGoQwFnCqFrUGMCRZUGJKTZIGPyldsifauoMnJPLTcDHmilcmahlqOELaAUYDBuzsVywnDQfwRLGIWozYaOAilMBcObErwgT", content2.GetDescription())
	require.Equal(t, "farming", content2.ProposalRoute())
	require.Equal(t, "PublicPlan", content2.ProposalType())
}
