package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bootstrap MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// LimitOrder defines a method for limit order
func (k msgServer) LimitOrder(goCtx context.Context, msg *types.MsgLimitOrder) (*types.MsgLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.Keeper.LimitOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgLimitOrderResponse{}, nil
}

// ModifyOrder defines a method for modify order
func (k msgServer) ModifyOrder(goCtx context.Context, msg *types.MsgModifyOrder) (*types.MsgModifyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.ModifyOrder(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgModifyOrderResponse{}, nil
}
