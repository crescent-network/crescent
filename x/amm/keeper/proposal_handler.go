package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func HandlePoolParameterChangeProposal(ctx sdk.Context, k Keeper, p *types.PoolParameterChangeProposal) error {
	for _, change := range p.Changes {
		pool, found := k.GetPool(ctx, change.PoolId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", change.PoolId)
		}
		if change.TickSpacing != 0 {
			if pool.TickSpacing == change.TickSpacing {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "tick spacing is not changed: %d", pool.TickSpacing)
			}
			if err := types.ValidateTickSpacing(pool.TickSpacing, change.TickSpacing); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}
			pool.TickSpacing = change.TickSpacing
		}
		if change.MinOrderQuantity != nil {
			pool.MinOrderQuantity = *change.MinOrderQuantity
		}
		k.SetPool(ctx, pool)
		if err := ctx.EventManager().EmitTypedEvent(&types.EventPoolParameterChanged{
			PoolId:      change.PoolId,
			TickSpacing: change.TickSpacing,
		}); err != nil {
			return err
		}
	}
	return nil
}

func HandlePublicFarmingPlanProposal(ctx sdk.Context, k Keeper, p *types.PublicFarmingPlanProposal) error {
	for _, req := range p.CreateRequests {
		farmingPoolAddr := sdk.MustAccAddressFromBech32(req.FarmingPoolAddress)
		termAddr := sdk.MustAccAddressFromBech32(req.TerminationAddress)
		if _, err := k.CreatePublicFarmingPlan(
			ctx, req.Description, farmingPoolAddr, termAddr,
			req.RewardAllocations, req.StartTime, req.EndTime); err != nil {
			return err
		}
	}
	for _, req := range p.TerminateRequests {
		plan, found := k.GetFarmingPlan(ctx, req.FarmingPlanId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "farming plan %d not found", req.FarmingPlanId)
		}
		if err := k.TerminateFarmingPlan(ctx, plan); err != nil {
			return err
		}
	}
	return nil
}
