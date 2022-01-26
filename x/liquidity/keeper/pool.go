package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
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
func (k Keeper) GetPoolBalance(ctx sdk.Context, pool types.Pool, pair types.Pair) (rx sdk.Int, ry sdk.Int) {
	reserveAddr := pool.GetReserveAddress()
	rx = k.bankKeeper.GetBalance(ctx, reserveAddr, pair.QuoteCoinDenom).Amount
	ry = k.bankKeeper.GetBalance(ctx, reserveAddr, pair.BaseCoinDenom).Amount
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

// CreatePool handles types.MsgCreatePool and creates a pool.
func (k Keeper) CreatePool(ctx sdk.Context, msg *types.MsgCreatePool) (types.Pool, error) {
	params := k.GetParams(ctx)

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found {
		return types.Pool{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", msg.PairId)
	}

	for _, coin := range msg.DepositCoins {
		if coin.Denom != pair.BaseCoinDenom && coin.Denom != pair.QuoteCoinDenom {
			return types.Pool{}, sdkerrors.Wrapf(types.ErrInvalidCoinDenom, "coin denom %s is not in the pair", coin.Denom)
		}
		if coin.Amount.LT(params.MinInitialDepositAmount) {
			return types.Pool{}, types.ErrInsufficientDepositAmount // TODO: more detail error?
		}
	}

	// Check to see if there is a pool with the pair.
	// Creating multiple pools with the same pair is disallowed, but it will be allowed in v2.
	duplicate := false
	k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool) {
		if !pool.Disabled {
			duplicate = true
			return true
		}
		return false
	})
	if duplicate {
		return types.Pool{}, types.ErrPoolAlreadyExists
	}

	// Create and save the new pool object.
	poolId := k.GetNextPoolIdWithUpdate(ctx)
	pool := types.NewPool(poolId, pair.Id)
	k.SetPool(ctx, pool)

	// Send deposit coins to the pool's reserve account.
	creator := msg.GetCreator()
	if err := k.bankKeeper.SendCoins(ctx, creator, pool.GetReserveAddress(), msg.DepositCoins); err != nil {
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
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositCoins, msg.DepositCoins.String()),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyReserveAddress, pool.ReserveAddress),
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

	pair, _ := k.GetPair(ctx, pool.PairId)

	for _, coin := range msg.DepositCoins {
		if coin.Denom != pair.BaseCoinDenom && coin.Denom != pair.QuoteCoinDenom {
			return types.DepositRequest{}, sdkerrors.Wrapf(types.ErrInvalidCoinDenom, "coin denom %s is not in the pair", coin.Denom)
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, msg.GetDepositor(), types.GlobalEscrowAddr, msg.DepositCoins); err != nil {
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
			sdk.NewAttribute(types.AttributeKeyDepositCoins, msg.DepositCoins.String()),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
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
	req := types.NewWithdrawRequest(msg, requestId, ctx.BlockHeight())
	k.SetWithdrawRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawBatch,
			sdk.NewAttribute(types.AttributeKeyWithdrawer, msg.Withdrawer),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyPoolCoin, msg.PoolCoin.String()),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
		),
	})

	return req, nil
}

// ExecuteDepositRequest executes a deposit request.
func (k Keeper) ExecuteDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	if pool.Disabled {
		if err := k.RefundDepositRequestAndSetStatus(ctx, req, types.RequestStatusFailed); err != nil {
			return fmt.Errorf("refund deposit request: %w", err)
		}
		return nil
	}

	pair, _ := k.GetPair(ctx, pool.PairId)

	rx, ry := k.GetPoolBalance(ctx, pool, pair)
	ps := k.GetPoolCoinSupply(ctx, pool)
	poolInfo := types.NewPoolInfo(rx, ry, ps)
	if types.IsDepletedPool(poolInfo) {
		k.MarkPoolAsDisabled(ctx, pool)
		if err := k.RefundDepositRequestAndSetStatus(ctx, req, types.RequestStatusFailed); err != nil {
			return fmt.Errorf("refund deposit request: %w", err)
		}
		return nil
	}

	ax, ay, pc := types.DepositToPool(poolInfo, req.DepositCoins.AmountOf(pair.QuoteCoinDenom), req.DepositCoins.AmountOf(pair.BaseCoinDenom))

	if pc.IsZero() {
		if err := k.RefundDepositRequestAndSetStatus(ctx, req, types.RequestStatusFailed); err != nil {
			return fmt.Errorf("refund deposit request: %w", err)
		}
		return nil
	}

	mintedPoolCoin := sdk.NewCoin(pool.PoolCoinDenom, pc)
	mintingCoins := sdk.NewCoins(mintedPoolCoin)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintingCoins); err != nil {
		return err
	}

	acceptedCoins := sdk.NewCoins(sdk.NewCoin(pair.QuoteCoinDenom, ax), sdk.NewCoin(pair.BaseCoinDenom, ay))
	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, pool.GetReserveAddress(), acceptedCoins)
	bulkOp.SendCoins(k.accountKeeper.GetModuleAddress(types.ModuleName), req.GetDepositor(), mintingCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	req.AcceptedCoins = acceptedCoins
	req.MintedPoolCoin = mintedPoolCoin
	req.Status = types.RequestStatusSucceeded
	k.SetDepositRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) RefundDepositRequestAndSetStatus(ctx sdk.Context, req types.DepositRequest, status types.RequestStatus) error {
	refundingCoins, hasNeg := req.DepositCoins.SafeSub(req.AcceptedCoins)
	if hasNeg {
		return fmt.Errorf("refunding coins amount is negative")
	}
	if !refundingCoins.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, types.GlobalEscrowAddr, req.GetDepositor(), refundingCoins); err != nil {
			return err
		}
	}
	req.Status = status
	k.SetDepositRequest(ctx, req)
	return nil
}

// ExecuteWithdrawRequest executes a withdraw request.
func (k Keeper) ExecuteWithdrawRequest(ctx sdk.Context, req types.WithdrawRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	if pool.Disabled {
		k.SetWithdrawRequestStatus(ctx, req, types.RequestStatusFailed)
		return nil
	}

	pair, _ := k.GetPair(ctx, pool.PairId)

	rx, ry := k.GetPoolBalance(ctx, pool, pair)
	ps := k.GetPoolCoinSupply(ctx, pool)
	poolInfo := types.NewPoolInfo(rx, ry, ps)
	if types.IsDepletedPool(poolInfo) {
		k.MarkPoolAsDisabled(ctx, pool)
		k.SetWithdrawRequestStatus(ctx, req, types.RequestStatusFailed)
		return nil
	}

	params := k.GetParams(ctx)
	x, y := types.WithdrawFromPool(poolInfo, req.PoolCoin.Amount, params.WithdrawFeeRate)

	withdrawnCoins := sdk.NewCoins(sdk.NewCoin(pair.QuoteCoinDenom, x), sdk.NewCoin(pair.BaseCoinDenom, y))
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

	req.WithdrawnCoins = withdrawnCoins
	req.Status = types.RequestStatusSucceeded
	k.SetWithdrawRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) SetWithdrawRequestStatus(ctx sdk.Context, req types.WithdrawRequest, status types.RequestStatus) {
	req.Status = status
	k.SetWithdrawRequest(ctx, req)
}
