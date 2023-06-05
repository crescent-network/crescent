package keeper

import (
	"context"

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

// MintShare defines a method for adding liquidity to the public position and
// minting liquid farm share.
func (m msgServer) MintShare(goCtx context.Context, msg *types.MsgMintShare) (*types.MsgMintShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	mintedShare, _, liquidity, amt, err := m.Keeper.MintShare(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.DesiredAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintShareResponse{
		MintedShare: mintedShare,
		Liquidity:   liquidity,
		Amount:      amt,
	}, nil
}

// BurnShare defines a method for burning liquid farm share to withdraw underlying pool assets.
func (m msgServer) BurnShare(goCtx context.Context, msg *types.MsgBurnShare) (*types.MsgBurnShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	removedLiquidity, _, amt, err := m.Keeper.BurnShare(ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.LiquidFarmId, msg.Share)
	if err != nil {
		return nil, err
	}

	return &types.MsgBurnShareResponse{
		RemovedLiquidity: removedLiquidity,
		Amount:           amt,
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
