package farming

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	lastEpochTime, found := k.GetLastEpochTime(ctx)
	if !found {
		k.SetLastEpochTime(ctx, ctx.BlockTime())
	} else if ctx.BlockTime().Day()-lastEpochTime.Day() > 0 {
		if err := k.DistributeRewards(ctx); err != nil {
			panic(err)
		}
		k.ProcessQueuedCoins(ctx)

		k.SetLastEpochTime(ctx, ctx.BlockTime())
	}

	// TODO: implement plan termination logic
}
