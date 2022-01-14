package liquidity

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/keeper"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.RefundAndDeleteDepositRequestsToBeDeleted(ctx)
	k.RefundAndDeleteWithdrawRequestsToBeDeleted(ctx)
	k.RefundAndDeleteSwapRequestsToBeDeleted(ctx)
	k.DeleteCancelSwapRequestsToBeDeleted(ctx)
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if ctx.BlockHeight()%int64(params.BatchSize) == 0 {
		// Handle CancelSwapRequests.
		k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
			k.ExecuteCancelSwapRequest(ctx, req)
			return false
		})
		// Run order book matching on all pairs.
		k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool) {
			if err := k.ExecuteMatching(ctx, pair); err != nil {
				panic(err)
			}
			return false
		})
		// Cancel expired SwapRequests.
		k.IterateAllSwapRequests(ctx, func(req types.SwapRequest) (stop bool) {
			if !ctx.BlockTime().Before(req.CanceledAt) { // CanceledAt <= BlockTime
				k.CancelSwapRequest(ctx, req)
			}
			return false
		})
		// Handle DepositRequests.
		k.IterateAllDepositRequests(ctx, func(req types.DepositRequest) (stop bool) {
			if err := k.ExecuteDepositRequest(ctx, req); err != nil {
				panic(err)
			}
			return false
		})
		// Handle WithdrawRequests.
		k.IterateAllWithdrawRequests(ctx, func(req types.WithdrawRequest) (stop bool) {
			if err := k.ExecuteWithdrawRequest(ctx, req); err != nil {
				panic(err)
			}
			return false
		})
	}
}
