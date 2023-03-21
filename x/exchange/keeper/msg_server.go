package keeper

import (
	"context"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
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

func (k msgServer) CreateSpotMarket(goCtx context.Context, msg *types.MsgCreateSpotMarket) (*types.MsgCreateSpotMarketResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgCreateSpotMarketResponse{}, nil
}

func (k msgServer) PlaceSpotLimitOrder(goCtx context.Context, msg *types.MsgPlaceSpotLimitOrder) (*types.MsgPlaceSpotLimitOrderResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgPlaceSpotLimitOrderResponse{}, nil
}

func (k msgServer) PlaceSpotMarketOrder(goCtx context.Context, msg *types.MsgPlaceSpotMarketOrder) (*types.MsgPlaceSpotMarketOrderResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgPlaceSpotMarketOrderResponse{}, nil
}
