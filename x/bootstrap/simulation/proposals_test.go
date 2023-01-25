package simulation_test

//import (
//	"math/rand"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/crescent-network/crescent/v4/app/params"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/simulation"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
//)
//
//func TestProposalContents(t *testing.T) {
//	app, ctx := createTestApp(false)
//
//	s := rand.NewSource(1)
//	r := rand.New(s)
//
//	accounts := getTestingAccounts(t, r, app, ctx, 10)
//
//	// initialize parameters
//	param := app.BootstrapKeeper.GetParams(ctx)
//	param.IncentivePairs = simulation.GenIncentivePairs(r)
//	param.IncentiveBudgetAddress = accounts[1].Address.String()
//	app.BootstrapKeeper.SetParams(ctx, param)
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[0].Address.String(),
//		PairId:   1,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[0].Address.String(),
//		PairId:   2,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[3].Address.String(),
//		PairId:   1,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[5].Address.String(),
//		PairId:   1,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[4].Address.String(),
//		PairId:   2,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[4].Address.String(),
//		PairId:   3,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[5].Address.String(),
//		PairId:   3,
//		Eligible: true,
//	})
//
//	app.BootstrapKeeper.SetBootstrap(ctx, types.Bootstrap{
//		Address:  accounts[6].Address.String(),
//		PairId:   3,
//		Eligible: true,
//	})
//
//	// execute ProposalContents function
//	weightedProposalContent := simulation.ProposalContents(app.BankKeeper, app.BootstrapKeeper)
//	require.Len(t, weightedProposalContent, 3)
//
//	w0 := weightedProposalContent[0]
//	w1 := weightedProposalContent[1]
//	w2 := weightedProposalContent[2]
//
//	// tests w0 interface:
//	require.Equal(t, simulation.OpWeightSimulateBootstrapProposal, w0.AppParamsKey())
//	require.Equal(t, params.DefaultWeightBootstrapProposal, w0.DefaultWeight())
//
//	// tests w1 interface:
//	require.Equal(t, simulation.OpWeightSimulateChangeIncentivePairsProposal, w1.AppParamsKey())
//	require.Equal(t, params.DefaultWeightChangeIncentivePairs, w1.DefaultWeight())
//
//	// tests w2 interface:
//	require.Equal(t, simulation.OpWeightSimulateChangeDepositAmountProposal, w2.AppParamsKey())
//	require.Equal(t, params.DefaultWeightChangeDepositAmount, w2.DefaultWeight())
//
//	content0 := w0.ContentSimulatorFn()(r, ctx, accounts)
//	require.Nil(t, content0)
//
//	content1 := w1.ContentSimulatorFn()(r, ctx, accounts)
//	require.Nil(t, content1)
//
//	content2 := w2.ContentSimulatorFn()(r, ctx, accounts)
//	require.Nil(t, content2)
//}
