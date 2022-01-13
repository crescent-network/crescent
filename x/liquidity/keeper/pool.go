package keeper

import (
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
	rx = k.bankKeeper.GetBalance(ctx, reserveAddr, pool.YCoinDenom).Amount
	return
}

// GetPoolCoinSupply returns total pool coin supply of the pool.
func (k Keeper) GetPoolCoinSupply(ctx sdk.Context, pool types.Pool) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, pool.PoolCoinDenom).Amount
}

// CreatePool handles types.MsgCreatePool and creates a pool.
func (k Keeper) CreatePool(ctx sdk.Context, msg *types.MsgCreatePool) error {
	params := k.GetParams(ctx)
	if msg.XCoin.Amount.LT(params.MinInitialDepositAmount) || msg.YCoin.Amount.LT(params.MinInitialDepositAmount) {
		return types.ErrInsufficientDepositAmount // TODO: more detail error?
	}

	pair, found := k.GetPairByDenoms(ctx, msg.XCoin.Denom, msg.YCoin.Denom)
	if found {
		// If there is a pair with given denoms, check if there is a pool with
		// the pair.
		// Current version disallows to create multiple pools with same pair,
		// but later this can be changed(in v2).
		found := false
		k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool) {
			// TODO: check if pool isn't disabled
			// if !pool.Disabled {
			//     found = true
			//     return true
			// }
			// return false
			found = true
			return true
		})
		if found {
			return types.ErrPoolAlreadyExists
		}
	} else {
		// If there is no such pair, create one and store it to the variable.
		pair = k.CreatePair(ctx, msg.XCoin.Denom, msg.YCoin.Denom)
	}

	// Create and save the new pool object.
	poolId := k.GetNextPoolIdWithUpdate(ctx)
	pool := types.NewPool(poolId, pair.Id, msg.XCoin.Denom, msg.YCoin.Denom)
	k.SetPool(ctx, pool)

	// Send deposit coins to the pool's reserve account.
	creator := msg.GetCreator()
	depositCoins := sdk.NewCoins(msg.XCoin, msg.YCoin)
	if err := k.bankKeeper.SendCoins(ctx, creator, pool.GetReserveAddress(), depositCoins); err != nil {
		return err
	}
	// Send the pool creation fee to the fee collector.
	feeCollectorAddr, _ := sdk.AccAddressFromBech32(params.FeeCollectorAddress)
	if err := k.bankKeeper.SendCoins(ctx, creator, feeCollectorAddr, params.PoolCreationFee); err != nil {
		return sdkerrors.Wrap(err, "insufficient pool creation fee")
	}
	// Mint and send pool coin to the creator.
	poolCoin := sdk.NewCoin(pool.PoolCoinDenom, params.InitialPoolCoinSupply)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(poolCoin)); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, sdk.NewCoins(poolCoin)); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePool,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyXCoin, msg.XCoin.String()),
			sdk.NewAttribute(types.AttributeKeyYCoin, msg.YCoin.String()),
			sdk.NewAttribute(types.AttributeKeyMintedPoolCoin, poolCoin.String()),
		),
	})

	return nil
}

// DepositBatch handles types.MsgDepositBatch and stores the request.
func (k Keeper) DepositBatch(ctx sdk.Context, msg *types.MsgDepositBatch) error {
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool with id %d not found", msg.PoolId)
	}

	if msg.XCoin.Denom != pool.XCoinDenom || msg.YCoin.Denom != pool.YCoinDenom {
		return types.ErrWrongPair
	}

	depositCoins := sdk.NewCoins(msg.XCoin, msg.YCoin)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetDepositor(), types.GlobalEscrowAddr, depositCoins); err != nil {
		return err
	}

	requestId := k.GetNextDepositRequestIdWithUpdate(ctx, pool)
	req := types.DepositRequest{
		Id:        requestId,
		PoolId:    pool.Id,
		MsgHeight: ctx.BlockHeight(),
		Depositor: msg.Depositor,
		XCoin:     msg.XCoin,
		YCoin:     msg.YCoin,
	}
	k.SetDepositRequest(ctx, pool.Id, req)

	// TODO: need to emit an event?

	return nil
}

// WithdrawBatch handles types.MsgWithdrawBatch and stores the request.
func (k Keeper) WithdrawBatch(ctx sdk.Context, msg *types.MsgWithdrawBatch) error {
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool with id %d not found", msg.PoolId)
	}

	if msg.PoolCoin.Denom != pool.PoolCoinDenom {
		return types.ErrWrongPoolCoinDenom
	}

	if err := k.bankKeeper.SendCoins(ctx, msg.GetWithdrawer(), types.GlobalEscrowAddr, sdk.NewCoins(msg.PoolCoin)); err != nil {
		return err
	}

	requestId := k.GetNextWithdrawRequestIdWithUpdate(ctx, pool)
	req := types.WithdrawRequest{
		Id:         requestId,
		PoolId:     pool.Id,
		MsgHeight:  ctx.BlockHeight(),
		Withdrawer: msg.Withdrawer,
		PoolCoin:   msg.PoolCoin,
	}
	k.SetWithdrawRequest(ctx, pool.Id, req)

	// TODO: need to emit an event?

	return nil
}
