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
		if pool.TickSpacing%change.TickSpacing != 0 {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "tick spacing for pool %d must be a divisor of %d", change.PoolId, pool.TickSpacing)
		}
		ok := false
		for _, tickSpacing := range types.AllowedTickSpacings {
			if change.TickSpacing == tickSpacing {
				ok = true
				break
			}
		}
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "tick spacing %d is not allowed", change.TickSpacing)
		}
		pool.TickSpacing = change.TickSpacing
		k.SetPool(ctx, pool)
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
