package liquidity

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/keeper"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// logger := k.Logger(ctx)

	params := k.GetParams(ctx)
	if ctx.BlockHeight()%int64(params.BatchSize) == 0 {
		// TODO: match orders

		// TODO: find deposit requests and handle them
		// TODO: find withdrawal requests and handle them
	}
}
