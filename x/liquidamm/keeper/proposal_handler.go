package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func HandlePublicPositionCreateProposal(ctx sdk.Context, k Keeper, p *types.PublicPositionCreateProposal) error {
	if _, err := k.CreatePublicPosition(ctx, p.PoolId, p.LowerPrice, p.UpperPrice, p.MinBidAmount, p.FeeRate); err != nil {
		return err
	}
	return nil
}

func HandlePublicPositionParameterChangeProposal(ctx sdk.Context, k Keeper, p *types.PublicPositionParameterChangeProposal) error {
	for _, change := range p.Changes {
		publicPosition, found := k.GetPublicPosition(ctx, change.PublicPositionId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "public position %d not found", change.PublicPositionId)
		}
		publicPosition.MinBidAmount = change.MinBidAmount
		publicPosition.FeeRate = change.FeeRate
		k.SetPublicPosition(ctx, publicPosition)
		if err := ctx.EventManager().EmitTypedEvent(&types.EventPublicPositionParameterChanged{
			PublicPositionId: change.PublicPositionId,
			MinBidAmount:     change.MinBidAmount,
			FeeRate:          change.FeeRate,
		}); err != nil {
			return err
		}
	}
	return nil
}
