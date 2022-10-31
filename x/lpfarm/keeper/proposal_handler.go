package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

// HandleFarmingPlanProposal is a handler for executing a farming plan proposal.
func HandleFarmingPlanProposal(ctx sdk.Context, k Keeper, p *types.FarmingPlanProposal) error {
	for _, req := range p.CreatePlanRequests {
		farmingPoolAddr, _ := sdk.AccAddressFromBech32(req.FarmingPoolAddress)
		if _, err := k.CreatePublicPlan(
			ctx, req.Description, farmingPoolAddr,
			req.RewardAllocations, req.StartTime, req.EndTime); err != nil {
			return err
		}
	}
	for _, req := range p.TerminatePlanRequests {
		plan, found := k.GetPlan(ctx, req.PlanId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "plan %d not found", req.PlanId)
		}
		// TODO: do we actually need this restriction?
		if plan.IsPrivate {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot terminate private plans")
		}
		if err := k.TerminatePlan(ctx, plan); err != nil {
			return err
		}
	}
	return nil
}
