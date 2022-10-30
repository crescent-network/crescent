package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v3/app"
	"github.com/crescent-network/crescent/v3/app/params"
	"github.com/crescent-network/crescent/v3/x/liquidstaking/simulation"
	"github.com/crescent-network/crescent/v3/x/liquidstaking/types"
	minttypes "github.com/crescent-network/crescent/v3/x/mint/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(false)

	ctx.WithChainID("test-chain")

	cdc := types.ModuleCdc
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper, app.BankKeeper, app.LiquidStakingKeeper)

	s := rand.NewSource(2)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 10)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{params.DefaultWeightMsgLiquidStake, types.ModuleName, types.TypeMsgLiquidStake},
		{params.DefaultWeightMsgLiquidUnstake, types.ModuleName, types.TypeMsgLiquidUnstake},
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

func createTestApp(isCheckTx bool) (*chain.App, sdk.Context) {
	app := chain.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.LiquidStakingKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *chain.App, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 100_000_000)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewCoin("bstake", sdk.NewInt(50000000)),
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
