package marketmaker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v4/x/marketmaker/keeper"
	"github.com/crescent-network/crescent/v4/x/marketmaker/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgApplyMarketMaker:
			res, err := msgServer.ApplyMarketMaker(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgClaimIncentives:
			res, err := msgServer.ClaimIncentives(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

// NewMarketMakerProposalHandler creates a governance handler to manage new proposal types.
// It enables MarketMakerProposal to propose market maker inclusion / exclusion / rejection / distribution.
func NewMarketMakerProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.MarketMakerProposal:
			return keeper.HandleMarketMakerProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized market maker proposal content type: %T", c)
		}
	}
}
