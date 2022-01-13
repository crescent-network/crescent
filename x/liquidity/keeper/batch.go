package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (k Keeper) MarkDepositRequestToBeDeleted(ctx sdk.Context, req types.DepositRequest, succeeded bool) {
	req.Succeeded = succeeded
	req.ToBeDeleted = true
	k.SetDepositRequest(ctx, req)
}

func (k Keeper) MarkWithdrawRequestToBeDeleted(ctx sdk.Context, req types.WithdrawRequest, succeeded bool) {
	req.Succeeded = succeeded
	req.ToBeDeleted = true
	k.SetWithdrawRequest(ctx, req)
}

func (k Keeper) CancelSwapRequest(ctx sdk.Context, req types.SwapRequest) {
	req.Canceled = true
	k.SetSwapRequest(ctx, req)
}

func (k Keeper) MarkSwapRequestToBeDeleted(ctx sdk.Context, req types.SwapRequest) {
	req.ToBeDeleted = true
	k.SetSwapRequest(ctx, req)
}

func (k Keeper) MarkCancelSwapRequestToBeDeleted(ctx sdk.Context, req types.CancelSwapRequest, succeeded bool) {
	req.Succeeded = succeeded
	req.ToBeDeleted = true
	k.SetCancelSwapRequest(ctx, req)
}

// CancelSwapRequests cancels swap requests and deletes cancel swap requests.
func (k Keeper) CancelSwapRequests(ctx sdk.Context) {
	k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
		swapReq, found := k.GetSwapRequest(ctx, req.PairId, req.SwapRequestId)
		if !found {
			k.MarkCancelSwapRequestToBeDeleted(ctx, req, false)
			return false // continue iteration
		}

		if swapReq.BatchId < req.BatchId {
			k.CancelSwapRequest(ctx, swapReq)
			k.MarkCancelSwapRequestToBeDeleted(ctx, req, true)
		}

		return false
	})
}

// DeleteRequestsToBeDeleted deletes all requests that are marked as
// to be deleted.
func (k Keeper) DeleteRequestsToBeDeleted(ctx sdk.Context) {
	k.DeleteDepositRequestsToBeDeleted(ctx)
	k.DeleteWithdrawRequestsToBeDeleted(ctx)
	k.DeleteSwapRequestsToBeDeleted(ctx)
	k.DeleteCancelSwapRequestsToBeDeleted(ctx)
}

// ExecuteDepositRequest executes a deposit request.
func (k Keeper) ExecuteDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	// TODO: check if pool is disabled

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)

	poolInfo := types.NewPoolInfo(rx, ry, ps)
	ax, ay, pc := types.DepositToPool(poolInfo, req.XCoin.Amount, req.YCoin.Amount)

	if pc.IsZero() {
		k.MarkDepositRequestToBeDeleted(ctx, req, false)
		return nil
	}

	acceptedCoins := sdk.NewCoins(sdk.NewCoin(req.XCoin.Denom, ax), sdk.NewCoin(req.YCoin.Denom, ay))
	refundedCoins := sdk.NewCoins(
		sdk.NewCoin(req.XCoin.Denom, req.XCoin.Amount.Sub(ax)),
		sdk.NewCoin(req.YCoin.Denom, req.YCoin.Amount.Sub(ay)),
	)
	mintingCoins := sdk.NewCoins(sdk.NewCoin(pool.PoolCoinDenom, pc))

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintingCoins); err != nil {
		return err
	}

	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, pool.GetReserveAddress(), acceptedCoins)
	bulkOp.SendCoins(types.GlobalEscrowAddr, req.GetDepositor(), refundedCoins)
	bulkOp.SendCoins(k.accountKeeper.GetModuleAddress(types.ModuleName), req.GetDepositor(), mintingCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	k.MarkDepositRequestToBeDeleted(ctx, req, true)
	// TODO: emit an event?
	return nil
}

// ExecuteWithdrawRequest executes a withdraw request.
func (k Keeper) ExecuteWithdrawRequest(ctx sdk.Context, req types.WithdrawRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	// TODO: check if pool is disabled

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)

	poolInfo := types.NewPoolInfo(rx, ry, ps)
	params := k.GetParams(ctx)
	x, y := types.WithdrawFromPool(poolInfo, req.PoolCoin.Amount, params.WithdrawFeeRate)

	withdrawnCoins := sdk.NewCoins(sdk.NewCoin(pool.XCoinDenom, x), sdk.NewCoin(pool.YCoinDenom, y))
	burningCoins := sdk.NewCoins(req.PoolCoin)

	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, k.accountKeeper.GetModuleAddress(types.ModuleName), burningCoins)
	bulkOp.SendCoins(pool.GetReserveAddress(), req.GetWithdrawer(), withdrawnCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burningCoins); err != nil {
		return err
	}

	k.MarkWithdrawRequestToBeDeleted(ctx, req, true)
	// TODO: emit an event?
	return nil
}
