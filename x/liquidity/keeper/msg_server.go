package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreatePool defines a method to create a liquidity pool.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// TODO : not implemented yet

	return &types.MsgCreatePoolResponse{}, nil
}

// DepositBatch defines a method to deposit coins to the pool.
func (m msgServer) DepositBatch(goCtx context.Context, msg *types.MsgDepositBatch) (*types.MsgDepositBatchResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// TODO : not implemented yet

	return &types.MsgDepositBatchResponse{}, nil
}

// WithdrawBatch defines a method to withdraw pool coin from the pool.
func (m msgServer) WithdrawBatch(goCtx context.Context, msg *types.MsgWithdrawBatch) (*types.MsgWithdrawBatchResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// TODO : not implemented yet

	return &types.MsgWithdrawBatchResponse{}, nil
}

// SwapBatch defines a method to swap coin X to Y from the pool.
func (m msgServer) SwapBatch(goCtx context.Context, msg *types.MsgSwapBatch) (*types.MsgSwapBatchResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// TODO : not implemented yet

	return &types.MsgSwapBatchResponse{}, nil
}

// CancelSwapBatch defines a method to cancel the swap request.
func (m msgServer) CancelSwapBatch(goCtx context.Context, msg *types.MsgCancelSwapBatch) (*types.MsgCancelSwapBatchResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	// TODO : not implemented yet

	return &types.MsgCancelSwapBatchResponse{}, nil
}
