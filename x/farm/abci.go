package farm

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/keeper"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// TODO: terminate ended plans
	if err := k.AllocateRewards(ctx); err != nil {
		panic(err)
	}
	k.SetLastBlockTime(ctx, ctx.BlockTime())
}
