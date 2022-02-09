package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/claim/types"
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

// Claim defines a method to claim coins.
func (m msgServer) Claim(goCtx context.Context, msg *types.MsgClaim) (*types.MsgClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	claimant := msg.GetClaimant()

	record, found := m.Keeper.GetClaimRecord(ctx, claimant)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "claim record not found")
	}

	notClaimedActions := int64(0)
	for _, b := range []bool{record.SwapActionClaimed, record.DepositActionClaimed, record.FarmingActionClaimed} {
		if !b {
			notClaimedActions++
		}
	}
	if notClaimedActions == 0 {
		// TODO: emit an event
		return &types.MsgClaimResponse{}, nil
	}
	divisor := sdk.NewDec(notClaimedActions)
	amt, _ := sdk.NewDecCoinsFromCoins(record.InitialClaimableCoins...).QuoDecTruncate(divisor).TruncateDecimal()

	skip := true
	switch msg.ActionType {
	case types.ActionTypeSwap:
		if !record.SwapActionClaimed {
			record.SwapActionClaimed = true
			skip = false
		}
	case types.ActionTypeDeposit:
		if !record.DepositActionClaimed {
			record.DepositActionClaimed = true
			skip = false
		}
	case types.ActionTypeFarming:
		if !record.FarmingActionClaimed {
			record.FarmingActionClaimed = true
			skip = false
		}
	}
	if skip {
		return &types.MsgClaimResponse{}, nil
	}

	// TODO: send coins
	_ = amt

	return &types.MsgClaimResponse{}, nil
}
