package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	"github.com/crescent-network/crescent/v5/x/claim/simulation"
)

func TestSimulateMsgClaim(t *testing.T) {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s := rand.NewSource(0)
	r := rand.New(s)

	accs := getTestingAccounts(t, r, app, ctx, 1)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	op := simulation.SimulateMsgClaim(
		app.AccountKeeper, app.BankKeeper,
		app.LiquidityKeeper, app.LiquidStakingKeeper,
		app.GovKeeper, app.ClaimKeeper)
	opMsg, futureOps, err := op(r, app.BaseApp, ctx, accs, "")
	// disabled on antehandler
	require.Error(t, err)
	require.False(t, opMsg.OK)
	require.Len(t, futureOps, 0)
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *chain.App, ctx sdk.Context, n int) []simtypes.Account {
	accs := simtypes.RandomAccounts(r, n)

	params := app.LiquidStakingKeeper.GetParams(ctx)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewCoin(params.LiquidBondDenom, initAmt),
	)

	// add coins to the accounts
	for _, account := range accs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		require.NoError(t, chain.FundAccount(app.BankKeeper, ctx, account.Address, initCoins))
	}

	return accs
}
