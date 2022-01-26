package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func (k Keeper) ExecuteRequests(ctx sdk.Context) {
	k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
		if req.Status == types.RequestStatusNotExecuted {
			if err := k.ExecuteCancelSwapRequest(ctx, req); err != nil {
				panic(err)
			}
		}
		return false
	})
	k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool) {
		if err := k.ExecuteMatching(ctx, pair); err != nil {
			panic(err)
		}
		return false
	})
	k.IterateAllSwapRequests(ctx, func(req types.SwapRequest) (stop bool) {
		if !req.Status.IsCanceledOrExpired() && !ctx.BlockTime().Before(req.ExpireAt) { // ExpireAt <= BlockTime
			if err := k.RefundSwapRequestAndSetStatus(ctx, req, types.SwapRequestStatusExpired); err != nil {
				panic(err)
			}
		} else if req.Status == types.SwapRequestStatusCompleted {
			if err := k.RefundSwapRequest(ctx, req); err != nil {
				panic(err)
			}
		}
		return false
	})
	k.IterateAllDepositRequests(ctx, func(req types.DepositRequest) (stop bool) {
		if err := k.ExecuteDepositRequest(ctx, req); err != nil {
			panic(err)
		}
		return false
	})
	k.IterateAllWithdrawRequests(ctx, func(req types.WithdrawRequest) (stop bool) {
		if err := k.ExecuteWithdrawRequest(ctx, req); err != nil {
			panic(err)
		}
		return false
	})
}

func (k Keeper) DeleteOutdatedRequests(ctx sdk.Context) {
	k.IterateAllDepositRequests(ctx, func(req types.DepositRequest) (stop bool) {
		if req.Status.ShouldBeDeleted() {
			k.DeleteDepositRequest(ctx, req.PoolId, req.Id)
		}
		return false
	})
	k.IterateAllWithdrawRequests(ctx, func(req types.WithdrawRequest) (stop bool) {
		if req.Status.ShouldBeDeleted() {
			k.DeleteWithdrawRequest(ctx, req.PoolId, req.Id)
		}
		return false
	})
	k.IterateAllSwapRequests(ctx, func(req types.SwapRequest) (stop bool) {
		if req.Status.ShouldBeDeleted() {
			k.DeleteSwapRequest(ctx, req.PairId, req.Id)
		}
		return false
	})
	k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
		if req.Status.ShouldBeDeleted() {
			k.DeleteCancelSwapRequest(ctx, req.PairId, req.Id)
		}
		return false
	})
}
