package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetNextPoolIdWithUpdate increments pool id by one and set it.
func (k Keeper) GetNextPoolIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastPoolId(ctx) + 1
	k.SetLastPoolId(ctx, id)
	return id
}

// GetNextDepositRequestIdWithUpdate increments the pool's last deposit request
// id and returns it.
func (k Keeper) GetNextDepositRequestIdWithUpdate(ctx sdk.Context, pool types.Pool) uint64 {
	id := pool.LastDepositRequestId + 1
	pool.LastDepositRequestId = id
	k.SetPool(ctx, pool)
	return id
}

// GetNextWithdrawRequestIdWithUpdate increments the pool's last withdraw
// request id and returns it.
func (k Keeper) GetNextWithdrawRequestIdWithUpdate(ctx sdk.Context, pool types.Pool) uint64 {
	id := pool.LastWithdrawRequestId + 1
	pool.LastWithdrawRequestId = id
	k.SetPool(ctx, pool)
	return id
}

// GetPoolBalance returns x coin and y coin balance of the pool.
func (k Keeper) GetPoolBalance(ctx sdk.Context, pool types.Pool) (rx sdk.Int, ry sdk.Int) {
	reserveAddr := pool.GetReserveAddress()
	rx = k.bankKeeper.GetBalance(ctx, reserveAddr, pool.XCoinDenom).Amount
	ry = k.bankKeeper.GetBalance(ctx, reserveAddr, pool.YCoinDenom).Amount
	return
}

// GetPoolCoinSupply returns total pool coin supply of the pool.
func (k Keeper) GetPoolCoinSupply(ctx sdk.Context, pool types.Pool) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, pool.PoolCoinDenom).Amount
}

func (k Keeper) MarkPoolAsDisabled(ctx sdk.Context, pool types.Pool) {
	pool.Disabled = true
	k.SetPool(ctx, pool)
}

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

// CreatePool handles types.MsgCreatePool and creates a pool.
func (k Keeper) CreatePool(ctx sdk.Context, msg *types.MsgCreatePool) (types.Pool, error) {
	params := k.GetParams(ctx)
	if msg.XCoin.Amount.LT(params.MinInitialDepositAmount) || msg.YCoin.Amount.LT(params.MinInitialDepositAmount) {
		return types.Pool{}, types.ErrInsufficientDepositAmount // TODO: more detail error?
	}

	pair, found := k.GetPairByDenoms(ctx, msg.XCoin.Denom, msg.YCoin.Denom)
	if !found {
		// If there is no such pair, create one and store it to the variable.
		pair = k.CreatePair(ctx, msg.XCoin.Denom, msg.YCoin.Denom)
	}

	// If there is a pair with given denoms, check if there is a pool with
	// the pair.
	// Current version disallows to create multiple pools with same pair,
	// but later this can be changed(in v2).
	found = false
	k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool) {
		if !pool.Disabled {
			found = true
			return true
		}
		return false
	})
	if found {
		return types.Pool{}, types.ErrPoolAlreadyExists
	}

	// Create and save the new pool object.
	poolId := k.GetNextPoolIdWithUpdate(ctx)
	pool := types.NewPool(poolId, pair.Id, msg.XCoin.Denom, msg.YCoin.Denom)
	k.SetPool(ctx, pool)

	// Send deposit coins to the pool's reserve account.
	creator := msg.GetCreator()
	depositCoins := sdk.NewCoins(msg.XCoin, msg.YCoin)
	if err := k.bankKeeper.SendCoins(ctx, creator, pool.GetReserveAddress(), depositCoins); err != nil {
		return types.Pool{}, err
	}
	// Send the pool creation fee to the fee collector.
	feeCollectorAddr, _ := sdk.AccAddressFromBech32(params.FeeCollectorAddress)
	if err := k.bankKeeper.SendCoins(ctx, creator, feeCollectorAddr, params.PoolCreationFee); err != nil {
		return types.Pool{}, sdkerrors.Wrap(err, "insufficient pool creation fee")
	}
	// Mint and send pool coin to the creator.
	poolCoin := sdk.NewCoin(pool.PoolCoinDenom, params.InitialPoolCoinSupply)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(poolCoin)); err != nil {
		return types.Pool{}, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, sdk.NewCoins(poolCoin)); err != nil {
		return types.Pool{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePool,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyXCoin, msg.XCoin.String()),
			sdk.NewAttribute(types.AttributeKeyYCoin, msg.YCoin.String()),
			sdk.NewAttribute(types.AttributeKeyMintedPoolCoin, poolCoin.String()),
		),
	})

	return pool, nil
}

// DepositBatch handles types.MsgDepositBatch and stores the request.
func (k Keeper) DepositBatch(ctx sdk.Context, msg *types.MsgDepositBatch) (types.DepositRequest, error) {
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return types.DepositRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool with id %d not found", msg.PoolId)
	}

	if pool.Disabled {
		return types.DepositRequest{}, types.ErrDisabledPool
	}

	if msg.XCoin.Denom != pool.XCoinDenom || msg.YCoin.Denom != pool.YCoinDenom {
		return types.DepositRequest{}, types.ErrWrongPair
	}

	depositCoins := sdk.NewCoins(msg.XCoin, msg.YCoin)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetDepositor(), types.GlobalEscrowAddr, depositCoins); err != nil {
		return types.DepositRequest{}, err
	}

	requestId := k.GetNextDepositRequestIdWithUpdate(ctx, pool)
	req := types.NewDepositRequest(msg, pool, requestId, ctx.BlockHeight())
	k.SetDepositRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDepositBatch,
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.Depositor),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyXCoin, msg.XCoin.String()),
			sdk.NewAttribute(types.AttributeKeyYCoin, msg.YCoin.String()),
		),
	})

	return req, nil
}

// WithdrawBatch handles types.MsgWithdrawBatch and stores the request.
func (k Keeper) WithdrawBatch(ctx sdk.Context, msg *types.MsgWithdrawBatch) (types.WithdrawRequest, error) {
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return types.WithdrawRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool with id %d not found", msg.PoolId)
	}

	if pool.Disabled {
		return types.WithdrawRequest{}, types.ErrDisabledPool
	}

	if msg.PoolCoin.Denom != pool.PoolCoinDenom {
		return types.WithdrawRequest{}, types.ErrWrongPoolCoinDenom
	}

	if err := k.bankKeeper.SendCoins(ctx, msg.GetWithdrawer(), types.GlobalEscrowAddr, sdk.NewCoins(msg.PoolCoin)); err != nil {
		return types.WithdrawRequest{}, err
	}

	requestId := k.GetNextWithdrawRequestIdWithUpdate(ctx, pool)
	req := types.NewWithdrawRequest(msg, pool, requestId, ctx.BlockHeight())
	k.SetWithdrawRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawBatch,
			sdk.NewAttribute(types.AttributeKeyWithdrawer, msg.Withdrawer),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyPoolCoin, msg.PoolCoin.String()),
		),
	})

	return req, nil
}

// ExecuteDepositRequest executes a deposit request.
func (k Keeper) ExecuteDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	pool, found := k.GetPool(ctx, req.PoolId)
	if !found || pool.Disabled {
		k.MarkDepositRequestToBeDeleted(ctx, req, false)
		return nil
	}

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)
	poolInfo := types.NewPoolInfo(rx, ry, ps)
	if types.IsDepletedPool(poolInfo) {
		k.MarkPoolAsDisabled(ctx, pool)
		k.MarkDepositRequestToBeDeleted(ctx, req, false)
		return nil
	}

	ax, ay, pc := types.DepositToPool(poolInfo, req.XCoin.Amount, req.YCoin.Amount)

	if pc.IsZero() {
		k.MarkDepositRequestToBeDeleted(ctx, req, false)
		return nil
	}

	mintingCoins := sdk.NewCoins(sdk.NewCoin(pool.PoolCoinDenom, pc))

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintingCoins); err != nil {
		return err
	}

	acceptedXCoin, acceptedYCoin := sdk.NewCoin(req.XCoin.Denom, ax), sdk.NewCoin(req.YCoin.Denom, ay)
	acceptedCoins := sdk.NewCoins(acceptedXCoin, acceptedYCoin)
	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, pool.GetReserveAddress(), acceptedCoins)
	bulkOp.SendCoins(k.accountKeeper.GetModuleAddress(types.ModuleName), req.GetDepositor(), mintingCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	req.AcceptedXCoin = acceptedXCoin
	req.AcceptedYCoin = acceptedYCoin
	req.Succeeded = true
	req.ToBeDeleted = true
	k.SetDepositRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) RefundDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	refundingCoins := sdk.NewCoins(req.XCoin.Sub(req.AcceptedXCoin), req.YCoin.Sub(req.AcceptedYCoin))
	if !refundingCoins.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, types.GlobalEscrowAddr, req.GetDepositor(), refundingCoins); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) RefundAndDeleteDepositRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateDepositRequestsToBeDeleted(ctx, func(req types.DepositRequest) (stop bool) {
		if err := k.RefundDepositRequest(ctx, req); err != nil {
			panic(err)
		}
		k.DeleteDepositRequest(ctx, req.PoolId, req.Id)
		return false
	})
}

// ExecuteWithdrawRequest executes a withdraw request.
func (k Keeper) ExecuteWithdrawRequest(ctx sdk.Context, req types.WithdrawRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	if pool.Disabled {
		k.MarkWithdrawRequestToBeDeleted(ctx, req, false)
		return nil
	}

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)
	poolInfo := types.NewPoolInfo(rx, ry, ps)
	if types.IsDepletedPool(poolInfo) {
		k.MarkPoolAsDisabled(ctx, pool)
		k.MarkWithdrawRequestToBeDeleted(ctx, req, false)
		return nil
	}

	params := k.GetParams(ctx)
	x, y := types.WithdrawFromPool(poolInfo, req.PoolCoin.Amount, params.WithdrawFeeRate)

	withdrawnXCoin, withdrawnYCoin := sdk.NewCoin(pool.XCoinDenom, x), sdk.NewCoin(pool.YCoinDenom, y)
	withdrawnCoins := sdk.NewCoins(withdrawnXCoin, withdrawnYCoin)
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

	// If the pool coin supply becomes 0, disable the pool.
	if req.PoolCoin.Amount.Equal(ps) {
		k.MarkPoolAsDisabled(ctx, pool)
	}

	req.WithdrawnXCoin = withdrawnXCoin
	req.WithdrawnYCoin = withdrawnYCoin
	req.Succeeded = true
	req.ToBeDeleted = true
	k.SetWithdrawRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) RefundAndDeleteWithdrawRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateWithdrawRequestsToBeDeleted(ctx, func(req types.WithdrawRequest) (stop bool) {
		// TODO: need a refund? maybe not
		k.DeleteWithdrawRequest(ctx, req.PoolId, req.Id)
		return false
	})
}
