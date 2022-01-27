package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
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

// CreatePair defines a method to create a pair.
func (m msgServer) CreatePair(goCtx context.Context, msg *types.MsgCreatePair) (*types.MsgCreatePairResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.CreatePair(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCreatePairResponse{}, nil
}

// CreatePool defines a method to create a liquidity pool.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.CreatePool(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCreatePoolResponse{}, nil
}

// DepositBatch defines a method to deposit coins to the pool.
func (m msgServer) DepositBatch(goCtx context.Context, msg *types.MsgDepositBatch) (*types.MsgDepositBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.DepositBatch(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgDepositBatchResponse{}, nil
}

// WithdrawBatch defines a method to withdraw pool coin from the pool.
func (m msgServer) WithdrawBatch(goCtx context.Context, msg *types.MsgWithdrawBatch) (*types.MsgWithdrawBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.WithdrawBatch(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawBatchResponse{}, nil
}

// LimitOrderBatch defines a method to making a limit order.
func (m msgServer) LimitOrderBatch(goCtx context.Context, msg *types.MsgLimitOrderBatch) (*types.MsgLimitOrderBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.LimitOrderBatch(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgLimitOrderBatchResponse{}, nil
}

// MarketOrderBatch defines a method to making a market order.
func (m msgServer) MarketOrderBatch(goCtx context.Context, msg *types.MsgMarketOrderBatch) (*types.MsgMarketOrderBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.MarketOrderBatch(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgMarketOrderBatchResponse{}, nil
}

// CancelOrder defines a method to cancel an order.
func (m msgServer) CancelOrder(goCtx context.Context, msg *types.MsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CancelOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCancelOrderResponse{}, nil
}
