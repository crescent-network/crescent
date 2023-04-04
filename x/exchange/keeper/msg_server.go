package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	market, err := k.Keeper.CreateSpotMarket(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.BaseDenom, msg.QuoteDenom)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateSpotMarketResponse{MarketId: market.Id}, nil
}

func (k msgServer) PlaceSpotLimitOrder(goCtx context.Context, msg *types.MsgPlaceSpotLimitOrder) (*types.MsgPlaceSpotLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	order, _, err := k.Keeper.PlaceSpotOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.MarketId,
		msg.IsBuy, &msg.Price, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotLimitOrderResponse{OrderId: order.Id}, nil
}

func (k msgServer) PlaceSpotMarketOrder(goCtx context.Context, msg *types.MsgPlaceSpotMarketOrder) (*types.MsgPlaceSpotMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, _, err := k.Keeper.PlaceSpotOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.MarketId,
		msg.IsBuy, nil, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotMarketOrderResponse{}, nil
}
