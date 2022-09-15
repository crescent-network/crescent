package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

// CreatePrivatePlan creates a new private farming plan.
func (k Keeper) CreatePrivatePlan(
	ctx sdk.Context, creatorAddr sdk.AccAddress, description string,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time,
) (types.Plan, error) {
	fee := k.GetPrivatePlanCreationFee(ctx)
	feeCollectorAddr, err := sdk.AccAddressFromBech32(k.GetFeeCollector(ctx))
	if err != nil {
		return types.Plan{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, creatorAddr, feeCollectorAddr, fee); err != nil {
		return types.Plan{}, err
	}

	farmingPoolAddr := sdk.AccAddress{} // TODO: derive correct address
	return k.createPlan(
		ctx, description, farmingPoolAddr, farmingPoolAddr, rewardAllocs, startTime, endTime, true)
}

func (k Keeper) createPlan(
	ctx sdk.Context, description string, srcAddr, termAddr sdk.AccAddress,
	rewardAllocs []types.RewardAllocation, startTime, endTime time.Time, isPrivate bool,
) (types.Plan, error) {
	// TODO: validate reward allocations and start/end time

	// Generate next plan id and update the last plan id.
	id, _ := k.GetLastPlanId(ctx)
	id++
	k.SetLastPlanId(ctx, id)

	plan := types.NewPlan(
		id, description, srcAddr, termAddr, rewardAllocs, startTime, endTime, isPrivate)
	k.SetPlan(ctx, plan)
	return plan, nil
}
