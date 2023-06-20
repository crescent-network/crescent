package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func HandleLiquidFarmCreateProposal(ctx sdk.Context, k Keeper, p *types.LiquidFarmCreateProposal) error {
	if _, err := k.CreateLiquidFarm(ctx, p.PoolId, p.LowerPrice, p.UpperPrice, p.MinBidAmount, p.FeeRate); err != nil {
		return err
	}
	return nil
}

func HandleLiquidFarmParameterChangeProposal(ctx sdk.Context, k Keeper, p *types.LiquidFarmParameterChangeProposal) error {
	for _, change := range p.Changes {
		liquidFarm, found := k.GetLiquidFarm(ctx, change.LiquidFarmId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "liquid farm %d not found", change.LiquidFarmId)
		}
		liquidFarm.MinBidAmount = change.MinBidAmount
		liquidFarm.FeeRate = change.FeeRate
		k.SetLiquidFarm(ctx, liquidFarm)
		if err := ctx.EventManager().EmitTypedEvent(&types.EventLiquidFarmParameterChanged{
			LiquidFarmId: change.LiquidFarmId,
			MinBidAmount: change.MinBidAmount,
			FeeRate:      change.FeeRate,
		}); err != nil {
			return err
		}
	}
	return nil
}
