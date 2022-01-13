package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// CancelSwapRequests cancels swap requests and deletes cancel swap requests.
func (k Keeper) CancelSwapRequests(ctx sdk.Context) {
	k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
		swapReq, found := k.GetSwapRequest(ctx, req.PairId, req.SwapRequestId)
		if !found {
			req.Succeeded = false
			req.ToBeDeleted = true
			k.SetCancelSwapRequest(ctx, req)
			return false // continue iteration
		}

		if swapReq.BatchId < req.BatchId {
			swapReq.Canceled = true
			k.SetSwapRequest(ctx, swapReq)
			req.Succeeded = true
			req.ToBeDeleted = true
			k.SetCancelSwapRequest(ctx, req)
		}

		return false
	})
}

func (k Keeper) DeleteRequestsToBeDeleted(ctx sdk.Context) {
	k.DeleteAllDepositRequestsToBeDeleted(ctx)
	k.DeleteAllWithdrawRequestsToBeDeleted(ctx)
	k.DeleteAllSwapRequestsToBeDeleted(ctx)
	k.DeleteAllCancelSwapRequestsToBeDeleted(ctx)
}
