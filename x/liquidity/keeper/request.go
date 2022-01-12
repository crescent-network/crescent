package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetDepositRequest returns the particular deposit request.
func (k Keeper) GetDepositRequest(ctx sdk.Context, poolId, id uint64) (state types.DepositRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDepositRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshalDepositRequest(k.cdc, value)
	return state, true
}

// GetWithdrawRequest returns the particular withdraw request.
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, poolId, id uint64) (state types.WithdrawRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshaWithdrawRequest(k.cdc, value)
	return state, true
}

// GetSwapRequest returns the particular swap request.
func (k Keeper) GetSwapRequest(ctx sdk.Context, poolId, id uint64) (state types.SwapRequest, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetSwapRequestKey(poolId, id)

	value := store.Get(key)
	if value == nil {
		return state, false
	}

	state = types.MustUnmarshaSwapRequest(k.cdc, value)
	return state, true
}

// SetDepositRequest stores deposit request for the batch execution.
func (k Keeper) SetDepositRequest(ctx sdk.Context, poolId uint64, id uint64, state types.DepositRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalDepositRequest(k.cdc, state)
	store.Set(types.GetDepositRequestKey(poolId, id), bz)
}

// SetWithdrawRequest stores withdraw request for the batch execution.
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, poolId uint64, id uint64, state types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaWithdrawRequest(k.cdc, state)
	store.Set(types.GetWithdrawRequestKey(poolId, id), bz)
}

// SetSwapRequest stores swap request for the batch execution.
func (k Keeper) SetSwapRequest(ctx sdk.Context, poolId uint64, id uint64, state types.SwapRequest) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaSwapRequest(k.cdc, state)
	store.Set(types.GetDepositRequestKey(poolId, id), bz)
}
