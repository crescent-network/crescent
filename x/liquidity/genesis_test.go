package liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crescentapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestDefaultGenesis(t *testing.T) {
	genesisState := *types.DefaultGenesis()

	app := crescentapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	liquidity.InitGenesis(ctx, app.LiquidityKeeper, genesisState)
	got := liquidity.ExportGenesis(ctx, app.LiquidityKeeper)
	require.NotNil(t, got)
}
