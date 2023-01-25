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

	if err := k.Keeper.LimitOrder(ctx, msg.GetAddress(), msg.BootstrapPoolId, msg.Direction, msg.OfferCoin, msg.Price); err != nil {
		return nil, err
	}

	return &types.MsgLimitOrderResponse{}, nil
}
