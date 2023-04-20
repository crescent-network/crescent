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

	order, execQty, execQuote, err := k.Keeper.PlaceSpotLimitOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Price, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotLimitOrderResponse{
		Rested:           order.Id > 0,
		OrderId:          order.Id,
		ExecutedQuantity: execQty,
		ExecutedQuote:    execQuote,
	}, nil
}

func (k msgServer) PlaceSpotMarketOrder(goCtx context.Context, msg *types.MsgPlaceSpotMarketOrder) (*types.MsgPlaceSpotMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	execQty, execQuote, err := k.Keeper.PlaceSpotMarketOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotMarketOrderResponse{
		ExecutedQuantity: execQty,
		ExecutedQuote:    execQuote,
	}, nil
}

func (k msgServer) CancelSpotOrder(goCtx context.Context, msg *types.MsgCancelSpotOrder) (*types.MsgCancelSpotOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Keeper.CancelSpotOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelSpotOrderResponse{}, nil
}

func (k msgServer) SwapExactIn(goCtx context.Context, msg *types.MsgSwapExactIn) (*types.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	output, err := k.Keeper.SwapExactIn(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.Routes, msg.Input, msg.MinOutput)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapExactInResponse{
		Output: output,
	}, nil
}
