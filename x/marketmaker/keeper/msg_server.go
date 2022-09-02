package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the marketmaker MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// ApplyMarketMaker defines a method for apply to be market maker.
func (k msgServer) ApplyMarketMaker(goCtx context.Context, msg *types.MsgApplyMarketMaker) (*types.MsgApplyMarketMakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.ApplyMarketMaker(ctx, msg.GetAddress(), msg.PairIds); err != nil {
		return nil, err
	}

	return &types.MsgApplyMarketMakerResponse{}, nil
}

// ClaimIncentives defines a method for claim all claimable incentives of the market maker.
func (k msgServer) ClaimIncentives(goCtx context.Context, msg *types.MsgClaimIncentives) (*types.MsgClaimIncentivesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.ClaimIncentives(ctx, msg.GetAddress()); err != nil {
		return nil, err
	}

	return &types.MsgClaimIncentivesResponse{}, nil
}
