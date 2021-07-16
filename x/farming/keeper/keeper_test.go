package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/types"
)

// createTestApp returns a farming app with custom FarmingKeeper.
func createTestApp(isCheckTx bool) (*app.FarmingApp, sdk.Context) {
	app := app.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.FarmingKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}
