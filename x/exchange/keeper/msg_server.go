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

func (k msgServer) CreateMarket(goCtx context.Context, msg *types.MsgCreateMarket) (*types.MsgCreateMarketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	market, err := k.Keeper.CreateMarket(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.BaseDenom, msg.QuoteDenom)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateMarketResponse{MarketId: market.Id}, nil
}

func (k msgServer) PlaceLimitOrder(goCtx context.Context, msg *types.MsgPlaceLimitOrder) (*types.MsgPlaceLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	orderId, _, execQty, paid, received, err := k.Keeper.PlaceLimitOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceLimitOrderResponse{
		OrderId:          orderId,
		ExecutedQuantity: execQty,
		Paid:             paid,
		Received:         received,
	}, nil
}

func (k msgServer) PlaceMarketOrder(goCtx context.Context, msg *types.MsgPlaceMarketOrder) (*types.MsgPlaceMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	orderId, execQty, paid, received, err := k.Keeper.PlaceMarketOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceMarketOrderResponse{
		OrderId:          orderId,
		ExecutedQuantity: execQty,
		Paid:             paid,
		Received:         received,
	}, nil
}

func (k msgServer) PlaceMMLimitOrder(goCtx context.Context, msg *types.MsgPlaceMMLimitOrder) (*types.MsgPlaceMMLimitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	orderId, _, execQty, paid, received, err := k.Keeper.PlaceMMLimitOrder(
		ctx, msg.MarketId, sdk.MustAccAddressFromBech32(msg.Sender),
		msg.IsBuy, msg.Price, msg.Quantity, msg.Lifespan)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceMMLimitOrderResponse{
		OrderId:          orderId,
		ExecutedQuantity: execQty,
		Paid:             paid,
		Received:         received,
	}, nil
}

func (k msgServer) CancelOrder(goCtx context.Context, msg *types.MsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, refundedDeposit, err := k.Keeper.CancelOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelOrderResponse{
		RefundedDeposit: refundedDeposit,
	}, nil
}

func (k msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	output, _, err := k.Keeper.SwapExactAmountIn(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.Routes, msg.Input, msg.MinOutput, false)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapExactAmountInResponse{
		Output: output,
	}, nil
}
