package liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{}

	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	liquidity.InitGenesis(ctx, app.LiquidityKeeper, genesisState)
	got := liquidity.ExportGenesis(ctx, app.LiquidityKeeper)
	require.NotNil(t, got)
}
