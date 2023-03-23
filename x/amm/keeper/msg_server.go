package keeper

import (
	"context"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgCreatePoolResponse{}, nil
}

func (k msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgAddLiquidityResponse{}, nil
}

func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgRemoveLiquidityResponse{}, nil
}
