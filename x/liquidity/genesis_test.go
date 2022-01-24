package liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidity"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func TestGenesis(t *testing.T) {
	genesisState := *types.DefaultGenesis()

	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	liquidity.InitGenesis(ctx, app.LiquidityKeeper, genesisState)
	got := liquidity.ExportGenesis(ctx, app.LiquidityKeeper)
	require.NotNil(t, got)
}
