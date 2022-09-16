package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
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

// Farm defines a method for farming pool coin to get minted LFCoin.
func (m msgServer) Farm(goCtx context.Context, msg *types.MsgFarm) (*types.MsgFarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.Farm(ctx, msg.PoolId, msg.GetFarmer(), msg.FarmingCoin); err != nil {
		return nil, err
	}

	return &types.MsgFarmResponse{}, nil
}

// Unfarm defines a method for unfarming LFCoin to receive the corresponding amount of pool coin.
func (m msgServer) Unfarm(goCtx context.Context, msg *types.MsgUnfarm) (*types.MsgUnfarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.Unfarm(ctx, msg.PoolId, msg.GetFarmer(), msg.BurningCoin); err != nil {
		return nil, err
	}

	return &types.MsgUnfarmResponse{}, nil
}

// UnfarmAndWithdraw defines a method for unfarming LFCoin and withdraw the corresponding amount of pool coin
// from the pool in the liquidity module.
// This is a convenient transaction message for a bidder to use when they participate in rewards auction.
func (m msgServer) UnfarmAndWithdraw(goCtx context.Context, msg *types.MsgUnfarmAndWithdraw) (*types.MsgUnfarmAndWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.UnfarmAndWithdraw(ctx, msg.PoolId, msg.GetFarmer(), msg.UnfarmingCoin); err != nil {
		return nil, err
	}

	return &types.MsgUnfarmAndWithdrawResponse{}, nil
}

// PlaceBid defines a method for placing a bid for a rewards auction.
func (m msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.PlaceBid(ctx, msg.PoolId, msg.GetBidder(), msg.BiddingCoin); err != nil {
		return nil, err
	}

	return &types.MsgPlaceBidResponse{}, nil
}

// RefundBid defines a method for refunding the bid for the auction.
func (m msgServer) RefundBid(goCtx context.Context, msg *types.MsgRefundBid) (*types.MsgRefundBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.RefundBid(ctx, msg.PoolId, msg.GetBidder()); err != nil {
		return nil, err
	}

	return &types.MsgRefundBidResponse{}, nil
}
