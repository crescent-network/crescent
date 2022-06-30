package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
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

func (m msgServer) CreateRangedPool(goCtx context.Context, msg *types.MsgCreateRangedPool) (*types.MsgCreateRangedPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.CreateRangedPool(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCreateRangedPoolResponse{}, nil
}

// Deposit defines a method to deposit coins to the pool.
func (m msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.Deposit(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{}, nil
}

// Withdraw defines a method to withdraw pool coin from the pool.
func (m msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.Withdraw(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawResponse{}, nil
}

// LimitOrder defines a method to making a limit order.
func (m msgServer) LimitOrder(goCtx context.Context, msg *types.MsgLimitOrder) (*types.MsgLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.LimitOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgLimitOrderResponse{}, nil
}

// MarketOrder defines a method to making a market order.
func (m msgServer) MarketOrder(goCtx context.Context, msg *types.MsgMarketOrder) (*types.MsgMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.MarketOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgMarketOrderResponse{}, nil
}

// CancelOrder defines a method to cancel an order.
func (m msgServer) CancelOrder(goCtx context.Context, msg *types.MsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CancelOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCancelOrderResponse{}, nil
}

// CancelAllOrders defines a method to cancel all orders.
func (m msgServer) CancelAllOrders(goCtx context.Context, msg *types.MsgCancelAllOrders) (*types.MsgCancelAllOrdersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CancelAllOrders(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgCancelAllOrdersResponse{}, nil
}
