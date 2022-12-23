package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v4/app"
	"github.com/crescent-network/crescent/v4/app/params"
	"github.com/crescent-network/crescent/v4/x/marketmaker/simulation"
	"github.com/crescent-network/crescent/v4/x/marketmaker/types"
	minttypes "github.com/crescent-network/crescent/v4/x/mint/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(false)

	ctx.WithChainID("test-chain")

	cdc := types.ModuleCdc
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper, app.BankKeeper, app.MarketMakerKeeper)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 1)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{params.DefaultWeightMsgApplyMarketMaker, types.ModuleName, types.TypeMsgApplyMarketMaker},
		{params.DefaultWeightMsgClaimIncentives, types.ModuleName, types.TypeMsgClaimIncentives},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgApplyMarketMaker tests the normal scenario of a valid message of type TypeMsgApplyMarketMaker.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgApplyMarketMaker(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated incentive pairs
	params := app.MarketMakerKeeper.GetParams(ctx)
	params.IncentivePairs = simulation.GenIncentivePairs(r)
	app.MarketMakerKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgApplyMarketMaker(app.AccountKeeper, app.BankKeeper, app.MarketMakerKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgApplyMarketMaker
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgApplyMarketMaker, msg.Type())
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Address)
	require.Equal(t, []uint64{2, 3}, msg.PairIds)
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgClaimIncentives tests the normal scenario of a valid message of type TypeMsgClaimIncentives.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgClaimIncentives(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 2)

	// setup randomly generated incentive pairs
	params := app.MarketMakerKeeper.GetParams(ctx)
	params.IncentivePairs = simulation.GenIncentivePairs(r)
	params.IncentiveBudgetAddress = accounts[1].Address.String()
	app.MarketMakerKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	app.MarketMakerKeeper.SetMarketMaker(ctx, types.MarketMaker{
		Address:  accounts[0].Address.String(),
		PairId:   1,
		Eligible: true,
	})

	err := app.MarketMakerKeeper.DistributeMarketMakerIncentives(ctx, []types.IncentiveDistribution{
		{
			Address: accounts[0].Address.String(),
			PairId:  1,
			Amount:  sdk.NewCoins(sdk.NewCoin("reward", sdk.NewInt(50000000))),
		},
	})
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgClaimIncentives(app.AccountKeeper, app.BankKeeper, app.MarketMakerKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgClaimIncentives
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgClaimIncentives, msg.Type())
	require.Equal(t, accounts[0].Address.String(), msg.Address)
	require.Len(t, futureOperations, 0)

	balances := app.BankKeeper.GetBalance(ctx, accounts[0].Address, "reward")
	require.Equal(t, sdk.NewInt64Coin("reward", 100000000), balances)
	balances = app.BankKeeper.GetBalance(ctx, accounts[1].Address, "reward")
	require.Equal(t, sdk.NewInt64Coin("reward", 0), balances)
}

func createTestApp(isCheckTx bool) (*chain.App, sdk.Context) {
	app := chain.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *chain.App, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 100_000_000)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewCoin("reward", sdk.NewInt(50000000)),
	)

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		err := simapp.FundAccount(app.BankKeeper, ctx, account.Address, initCoins)
		require.NoError(t, err)
	}

	return accounts
}
