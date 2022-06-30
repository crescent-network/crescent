package keeper

import (
	"fmt"
	"strconv"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// GetPlan returns a plan for a given plan id.
func (k Keeper) GetPlan(ctx sdk.Context, id uint64) (plan types.PlanI, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPlanKey(id))
	if bz == nil {
		return plan, false
	}

	return k.decodePlan(bz), true
}

// GetPlans returns all plans in the store.
func (k Keeper) GetPlans(ctx sdk.Context) (plans []types.PlanI) {
	k.IteratePlans(ctx, func(plan types.PlanI) (stop bool) {
		plans = append(plans, plan)
		return false
	})

	return plans
}

// SetPlan sets a plan for a given plan id.
func (k Keeper) SetPlan(ctx sdk.Context, plan types.PlanI) {
	id := plan.GetId()
	store := ctx.KVStore(k.storeKey)

	bz, err := k.MarshalPlan(plan)
	if err != nil {
		panic(err)
	}

	store.Set(types.GetPlanKey(id), bz)
}

// DeletePlan deletes a plan from the store.
// NOTE: this will cause supply invariant violation if called
func (k Keeper) DeletePlan(ctx sdk.Context, plan types.PlanI) {
	id := plan.GetId()
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPlanKey(id))
}

// IteratePlans iterates over all the stored plans and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IteratePlans(ctx sdk.Context, cb func(plan types.PlanI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PlanKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		plan := k.decodePlan(iterator.Value())

		if cb(plan) {
			break
		}
	}
}

// GetNextPlanIdWithUpdate returns and increments the global Plan ID counter.
// If the global plan number is not set, it initializes it with value 0.
func (k Keeper) GetNextPlanIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetGlobalPlanId(ctx) + 1
	k.SetGlobalPlanId(ctx, id)
	return id
}

// SetGlobalPlanId sets the global Plan ID counter.
func (k Keeper) SetGlobalPlanId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.GlobalPlanIdKey, bz)
}

// GetGlobalPlanId returns the global Plan ID counter.
func (k Keeper) GetGlobalPlanId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GlobalPlanIdKey)
	if bz == nil {
		// initialize the PlanId
		id = 0
	} else {
		val := gogotypes.UInt64Value{}
		k.cdc.MustUnmarshal(bz, &val)
		id = val.GetValue()
	}
	return id
}

func (k Keeper) decodePlan(bz []byte) types.PlanI {
	acc, err := k.UnmarshalPlan(bz)
	if err != nil {
		panic(err)
	}

	return acc
}

// GetNumActivePrivatePlans returns the number of active(non-terminated)
// private plans.
func (k Keeper) GetNumActivePrivatePlans(ctx sdk.Context) int {
	num := 0
	k.IteratePlans(ctx, func(plan types.PlanI) (stop bool) {
		if plan.GetType() == types.PlanTypePrivate && !plan.IsTerminated() {
			num++
		}
		return false
	})
	return num
}

// MarshalPlan serializes a plan.
func (k Keeper) MarshalPlan(plan types.PlanI) ([]byte, error) { // nolint:interfacer
	return k.cdc.MarshalInterface(plan)
}

// UnmarshalPlan returns a plan from raw serialized
// bytes of a Proto-based Plan type.
func (k Keeper) UnmarshalPlan(bz []byte) (plan types.PlanI, err error) {
	return plan, k.cdc.UnmarshalInterface(bz, &plan)
}

// CreateFixedAmountPlan sets fixed amount plan.
func (k Keeper) CreateFixedAmountPlan(ctx sdk.Context, msg *types.MsgCreateFixedAmountPlan, farmingPoolAcc, terminationAcc sdk.AccAddress, typ types.PlanType) (types.PlanI, error) {
	if !ctx.BlockTime().Before(msg.EndTime) { // EndTime <= BlockTime
		return nil, sdkerrors.Wrap(types.ErrInvalidPlanEndTime, "end time has already passed")
	}

	for _, coin := range msg.StakingCoinWeights {
		if k.bankKeeper.GetSupply(ctx, coin.Denom).Amount.IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrInvalidStakingCoinWeights, "denom %s has no supply", coin.Denom)
		}
	}
	for _, coin := range msg.EpochAmount {
		if k.bankKeeper.GetSupply(ctx, coin.Denom).Amount.IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrInvalidEpochAmount, "denom %s has no supply", coin.Denom)
		}
	}

	var maxNumDenoms int
	switch typ {
	case types.PlanTypePrivate:
		maxNumDenoms = types.PrivatePlanMaxNumDenoms
	case types.PlanTypePublic:
		maxNumDenoms = types.PublicPlanMaxNumDenoms
	}
	if len(msg.StakingCoinWeights) > maxNumDenoms {
		return nil, sdkerrors.Wrapf(
			types.ErrNumMaxDenomsLimit,
			"number of denoms in staking coin weights is %d, which exceeds the limit %d",
			len(msg.StakingCoinWeights), maxNumDenoms)
	}
	if len(msg.EpochAmount) > maxNumDenoms {
		return nil, sdkerrors.Wrapf(
			types.ErrNumMaxDenomsLimit,
			"number of denoms in epoch amount is %d, which exceeds the limit %d",
			len(msg.EpochAmount), maxNumDenoms)
	}

	params := k.GetParams(ctx)

	if typ == types.PlanTypePrivate {
		if uint32(k.GetNumActivePrivatePlans(ctx)) >= params.MaxNumPrivatePlans {
			return nil, types.ErrNumPrivatePlansLimit
		}

		feeCollectorAcc, _ := sdk.AccAddressFromBech32(params.FarmingFeeCollector) // Already validated
		if err := k.bankKeeper.SendCoins(ctx, msg.GetCreator(), feeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to pay private plan creation fee")
		}
	}

	nextId := k.GetNextPlanIdWithUpdate(ctx)
	basePlan := types.NewBasePlan(
		nextId,
		msg.Name,
		typ,
		farmingPoolAcc.String(),
		terminationAcc.String(),
		msg.StakingCoinWeights,
		msg.StartTime,
		msg.EndTime,
	)

	fixedPlan := types.NewFixedAmountPlan(basePlan, msg.EpochAmount)

	k.SetPlan(ctx, fixedPlan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFixedAmountPlan,
			sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(nextId, 10)),
			sdk.NewAttribute(types.AttributeKeyPlanName, msg.Name),
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, farmingPoolAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochAmount, msg.EpochAmount.String()),
		),
	})

	return fixedPlan, nil
}

// CreateRatioPlan sets ratio plan.
func (k Keeper) CreateRatioPlan(ctx sdk.Context, msg *types.MsgCreateRatioPlan, farmingPoolAcc, terminationAcc sdk.AccAddress, typ types.PlanType) (types.PlanI, error) {
	if !ctx.BlockTime().Before(msg.EndTime) { // EndTime <= BlockTime
		return nil, sdkerrors.Wrap(types.ErrInvalidPlanEndTime, "end time has already passed")
	}

	for _, coin := range msg.StakingCoinWeights {
		if k.bankKeeper.GetSupply(ctx, coin.Denom).Amount.IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrInvalidStakingCoinWeights, "denom %s has no supply", coin.Denom)
		}
	}

	var maxNumDenoms int
	switch typ {
	case types.PlanTypePrivate:
		maxNumDenoms = types.PrivatePlanMaxNumDenoms
	case types.PlanTypePublic:
		maxNumDenoms = types.PublicPlanMaxNumDenoms
	}
	if len(msg.StakingCoinWeights) > maxNumDenoms {
		return nil, sdkerrors.Wrapf(
			types.ErrNumMaxDenomsLimit,
			"number of denoms in staking coin weights is %d, which exceeds the limit %d",
			len(msg.StakingCoinWeights), maxNumDenoms)
	}

	params := k.GetParams(ctx)

	if typ == types.PlanTypePrivate {
		if uint32(k.GetNumActivePrivatePlans(ctx)) >= params.MaxNumPrivatePlans {
			return nil, types.ErrNumPrivatePlansLimit
		}

		feeCollectorAcc, _ := sdk.AccAddressFromBech32(params.FarmingFeeCollector) // Already validated
		if err := k.bankKeeper.SendCoins(ctx, msg.GetCreator(), feeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to pay private plan creation fee")
		}
	}

	nextId := k.GetNextPlanIdWithUpdate(ctx)
	basePlan := types.NewBasePlan(
		nextId,
		msg.Name,
		typ,
		farmingPoolAcc.String(),
		terminationAcc.String(),
		msg.StakingCoinWeights,
		msg.StartTime,
		msg.EndTime,
	)

	ratioPlan := types.NewRatioPlan(basePlan, msg.EpochRatio)

	k.SetPlan(ctx, ratioPlan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateRatioPlan,
			sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(nextId, 10)),
			sdk.NewAttribute(types.AttributeKeyPlanName, msg.Name),
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, farmingPoolAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochRatio, msg.EpochRatio.String()),
		),
	})

	return ratioPlan, nil
}

// TerminatePlan marks the plan as terminated.
// It moves the plan under different store key, which is for terminated plans.
func (k Keeper) TerminatePlan(ctx sdk.Context, plan types.PlanI) error {
	if plan.GetFarmingPoolAddress().String() != plan.GetTerminationAddress().String() {
		balances := k.bankKeeper.SpendableCoins(ctx, plan.GetFarmingPoolAddress())
		if !balances.IsZero() {
			if err := k.bankKeeper.SendCoins(ctx, plan.GetFarmingPoolAddress(), plan.GetTerminationAddress(), balances); err != nil {
				return err
			}
		}
	}

	switch plan.GetType() {
	case types.PlanTypePrivate:
		// For private plans, mark the plan as terminated so that it can be removed
		// later by the creator.
		_ = plan.SetTerminated(true)
		k.SetPlan(ctx, plan)
	case types.PlanTypePublic:
		// Delete the public plan immediately after terminating it.
		k.DeletePlan(ctx, plan)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePlanTerminated,
			sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(plan.GetId(), 10)),
		),
	})

	return nil
}

// TerminateEndedPlans terminates plans that have been ended.
func (k Keeper) TerminateEndedPlans(ctx sdk.Context) error {
	for _, plan := range k.GetPlans(ctx) {
		if !plan.IsTerminated() && !ctx.BlockTime().Before(plan.GetEndTime()) {
			if err := k.TerminatePlan(ctx, plan); err != nil {
				return fmt.Errorf("terminate plan %d: %w", plan.GetId(), err)
			}
		}
	}
	return nil
}

// RemovePlan removes a terminated plan and sends all remaining coins in the
// farming pool address to the termination address.
func (k Keeper) RemovePlan(ctx sdk.Context, creator sdk.AccAddress, planId uint64) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "plan %d not found", planId)
	}

	if !plan.IsTerminated() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "plan %d is not terminated yet", planId)
	}

	if plan.GetType() != types.PlanTypePrivate {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "plan %d is not a private plan", planId)
	}

	if !plan.GetTerminationAddress().Equals(creator) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only the plan creator can remove the plan")
	}

	// Refund private plan creation fee.
	params := k.GetParams(ctx)
	feeCollectorAcc, _ := sdk.AccAddressFromBech32(params.FarmingFeeCollector) // Already validated
	if err := k.bankKeeper.SendCoins(ctx, feeCollectorAcc, creator, params.PrivatePlanCreationFee); err != nil {
		return sdkerrors.Wrap(err, "failed to refund private plan creation fee")
	}

	k.DeletePlan(ctx, plan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemovePlan,
			sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(plan.GetId(), 10)),
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, plan.GetFarmingPoolAddress().String()),
			sdk.NewAttribute(types.AttributeKeyTerminationAddress, plan.GetTerminationAddress().String()),
		),
	})

	return nil
}

// DerivePrivatePlanFarmingPoolAcc returns a unique account address
// of a farming pool for a private plan.
func (k Keeper) DerivePrivatePlanFarmingPoolAcc(ctx sdk.Context, name string) (sdk.AccAddress, error) {
	nextPlanId := k.GetGlobalPlanId(ctx) + 1
	poolAcc := types.PrivatePlanFarmingPoolAcc(name, nextPlanId)
	if !k.bankKeeper.GetAllBalances(ctx, poolAcc).Empty() {
		return nil, types.ErrConflictPrivatePlanFarmingPool
	}
	return poolAcc, nil
}
