package liquidfarming

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// NewHandler returns a new msg handler.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgFarm:
			res, err := msgServer.Farm(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUnfarm:
			res, err := msgServer.Unfarm(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUnfarmAndWithdraw:
			res, err := msgServer.UnfarmAndWithdraw(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgPlaceBid:
			res, err := msgServer.PlaceBid(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRefundBid:
			res, err := msgServer.RefundBid(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}
