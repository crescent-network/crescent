package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

	market, found := k.Keeper.GetSpotMarket(ctx, msg.MarketId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}

	order, execQuote, err := k.Keeper.PlaceSpotLimitOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), market,
		msg.IsBuy, msg.Price, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotLimitOrderResponse{
		Order: order,
		Quote: execQuote,
	}, nil
}

func (k msgServer) PlaceSpotMarketOrder(goCtx context.Context, msg *types.MsgPlaceSpotMarketOrder) (*types.MsgPlaceSpotMarketOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	market, found := k.Keeper.GetSpotMarket(ctx, msg.MarketId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}

	order, execQuote, err := k.Keeper.PlaceSpotMarketOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), market,
		msg.IsBuy, msg.Quantity)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceSpotMarketOrderResponse{
		Order: order,
		Quote: execQuote,
	}, nil
}

func (k msgServer) CancelSpotOrder(goCtx context.Context, msg *types.MsgCancelSpotOrder) (*types.MsgCancelSpotOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.Keeper.CancelSpotOrder(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.MarketId, msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelSpotOrderResponse{}, nil
}
