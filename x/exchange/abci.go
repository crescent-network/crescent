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
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	for _, market := range markets {
		cacheCtx, writeCache := ctx.CacheContext()
		func() {
			defer func() {
				if r := recover(); r != nil {
					k.Logger(ctx).Error("panic in batch matching", "value", r)
				}
			}()
			if err := k.RunBatchMatching(cacheCtx, market); err != nil {
				k.Logger(ctx).Error("failed to run batch matching", "error", err)
			} else {
				writeCache()
			}
		}()
	}
}
