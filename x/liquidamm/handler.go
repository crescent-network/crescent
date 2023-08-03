package liquidamm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// NewHandler returns a new msg handler.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgMintShare:
			res, err := msgServer.MintShare(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgBurnShare:
			res, err := msgServer.BurnShare(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgPlaceBid:
			res, err := msgServer.PlaceBid(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func NewProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.PublicPositionCreateProposal:
			return keeper.HandlePublicPositionCreateProposal(ctx, k, c)
		case *types.PublicPositionParameterChangeProposal:
			return keeper.HandlePublicPositionParameterChangeProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}
