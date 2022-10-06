package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
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

// LiquidFarm defines a method for farming pool coin and mint LFCoin for the farmer.
func (m msgServer) LiquidFarm(goCtx context.Context, msg *types.MsgLiquidFarm) (*types.MsgLiquidFarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.LiquidFarm(ctx, msg.PoolId, msg.GetFarmer(), msg.FarmingCoin); err != nil {
		return nil, err
	}

	return &types.MsgLiquidFarmResponse{}, nil
}

// LiquidUnfarm defines a method for unfarming LFCoin to return the corresponding amount of pool coin.
func (m msgServer) LiquidUnfarm(goCtx context.Context, msg *types.MsgLiquidUnfarm) (*types.MsgLiquidUnfarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.LiquidUnfarm(ctx, msg.PoolId, msg.GetFarmer(), msg.UnfarmingCoin); err != nil {
		return nil, err
	}

	return &types.MsgLiquidUnfarmResponse{}, nil
}

// LiquidUnfarmAndWithdraw defines a method for unfarming LFCoin and withdraw the corresponding amount of pool coin
// from the pool in the liquidity module.
// This is a convenient transaction message for a bidder to use when they participate in rewards auction.
func (m msgServer) LiquidUnfarmAndWithdraw(goCtx context.Context, msg *types.MsgLiquidUnfarmAndWithdraw) (*types.MsgLiquidUnfarmAndWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.LiquidUnfarmAndWithdraw(ctx, msg.PoolId, msg.GetFarmer(), msg.UnfarmingCoin); err != nil {
		return nil, err
	}

	return &types.MsgLiquidUnfarmAndWithdrawResponse{}, nil
}

// PlaceBid defines a method for placing a bid for a rewards auction.
func (m msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.PlaceBid(ctx, msg.AuctionId, msg.PoolId, msg.GetBidder(), msg.BiddingCoin); err != nil {
		return nil, err
	}

	return &types.MsgPlaceBidResponse{}, nil
}

// RefundBid defines a method for refunding the bid for the auction.
func (m msgServer) RefundBid(goCtx context.Context, msg *types.MsgRefundBid) (*types.MsgRefundBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.RefundBid(ctx, msg.AuctionId, msg.PoolId, msg.GetBidder()); err != nil {
		return nil, err
	}

	return &types.MsgRefundBidResponse{}, nil
}
