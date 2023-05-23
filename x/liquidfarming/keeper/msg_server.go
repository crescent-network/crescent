package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
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

// MintShare defines a method for farming pool coin and mint LFCoin for the farmer.
func (m msgServer) MintShare(goCtx context.Context, msg *types.MsgMintShare) (*types.MsgMintShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	position, liquidity, amt, err := m.Keeper.MintShare(ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.DesiredAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintShareResponse{
		PositionId: position.Id,
		Liquidity:  liquidity,
		Amount:     amt,
	}, nil
}

// BurnShare defines a method for unfarming LFCoin to return the corresponding amount of pool coin.
func (m msgServer) BurnShare(goCtx context.Context, msg *types.MsgBurnShare) (*types.MsgBurnShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amt, err := m.Keeper.BurnShare(ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.Share)
	if err != nil {
		return nil, err
	}

	return &types.MsgBurnShareResponse{
		Amount: amt,
	}, nil
}

// PlaceBid defines a method for placing a bid for a rewards auction.
func (m msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.PlaceBid(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.RewardsAuctionId, msg.Share); err != nil {
		return nil, err
	}

	return &types.MsgPlaceBidResponse{}, nil
}

// RefundBid defines a method for refunding the bid for the auction.
func (m msgServer) RefundBid(goCtx context.Context, msg *types.MsgRefundBid) (*types.MsgRefundBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.Keeper.RefundBid(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.RewardsAuctionId); err != nil {
		return nil, err
	}

	return &types.MsgRefundBidResponse{}, nil
}

// FinishRewardsAuctions defines a method for finishing all rewards auctions.
// This message is just for testing purpose and it shouldn't be used in production.
func (k msgServer) FinishRewardsAuctions(goCtx context.Context, msg *types.MsgFinishRewardsAuctions) (*types.MsgFinishRewardsAuctionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if EnableAdvanceAuction {
		endTime, _ := k.GetNextRewardsAuctionEndTime(ctx)
		ctx = ctx.WithBlockTime(endTime)

		duration := k.GetRewardsAuctionDuration(ctx)
		nextEndTime := endTime.Add(duration)

		for _, l := range k.GetAllLiquidFarms(ctx) {
			auction, found := k.GetLastRewardsAuction(ctx, l.PoolId)
			if !found {
				k.CreateRewardsAuction(ctx, l.PoolId, nextEndTime)
			} else {
				if err := k.FinishRewardsAuction(ctx, auction, l.FeeRate); err != nil {
					panic(err)
				}
				k.CreateRewardsAuction(ctx, l.PoolId, nextEndTime)
			}
		}
		k.SetNextRewardsAuctionEndTime(ctx, nextEndTime)
	} else {
		return nil, fmt.Errorf("FinishAuctions is disabled")
	}

	return &types.MsgFinishRewardsAuctionsResponse{}, nil
}
