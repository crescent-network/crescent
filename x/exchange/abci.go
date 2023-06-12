package exchange

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	if err := k.CancelExpiredOrders(ctx); err != nil {
		panic(err)
	}
}

func MidBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Call RunBatchMatching for all markets.
	// This is because RunBatchMatching is not a heavy operation when the
	// market's order book is not crossed.
	// We could further optimize the process by settings transient flag
	// when we receive batch order msgs.
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		if err := k.RunBatchMatching(ctx, market); err != nil {
			panic(err)
		}
		return false
	})
}
