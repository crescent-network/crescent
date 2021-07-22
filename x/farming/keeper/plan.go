package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

// NewPlan sets the next plan number to a given plan interface
func (k Keeper) NewPlan(ctx sdk.Context, plan types.PlanI) types.PlanI {
	if err := plan.SetId(k.GetNextPlanIDWithUpdate(ctx)); err != nil {
		panic(err)
	}

	return plan
}

// GetPlan implements PlanI.
func (k Keeper) GetPlan(ctx sdk.Context, id uint64) (plan types.PlanI, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPlanKey(id))
	if bz == nil {
		return plan, false
	}

	return k.decodePlan(bz), true
}

// GetAllPlans returns all plans in the Keeper.
func (k Keeper) GetAllPlans(ctx sdk.Context) (plans []types.PlanI) {
	k.IterateAllPlans(ctx, func(plan types.PlanI) (stop bool) {
		plans = append(plans, plan)
		return false
	})

	return plans
}

// SetPlan implements PlanI.
func (k Keeper) SetPlan(ctx sdk.Context, plan types.PlanI) {
	id := plan.GetId()
	store := ctx.KVStore(k.storeKey)

	bz, err := k.MarshalPlan(plan)
	if err != nil {
		panic(err)
	}

	store.Set(types.GetPlanKey(id), bz)
}

// RemovePlan removes an plan for the plan mapper store.
// NOTE: this will cause supply invariant violation if called
func (k Keeper) RemovePlan(ctx sdk.Context, plan types.PlanI) {
	id := plan.GetId()
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPlanKey(id))
}

// IterateAllPlans iterates over all the stored plans and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllPlans(ctx sdk.Context, cb func(plan types.PlanI) (stop bool)) {
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

// GetPlansByFarmerAddrIndex reads from kvstore and return a specific Plan indexed by given farmer address
func (k Keeper) GetPlansByFarmerAddrIndex(ctx sdk.Context, farmerAcc sdk.AccAddress) (plans []types.PlanI) {
	k.IteratePlansByFarmerAddr(ctx, farmerAcc, func(plan types.PlanI) bool {
		plans = append(plans, plan)
		return false
	})

	return plans
}

// IteratePlansByFarmerAddr iterates over all the stored plans and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IteratePlansByFarmerAddr(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(plan types.PlanI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetPlansByFarmerIndexKey(farmerAcc))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		planID := gogotypes.UInt64Value{}

		err := k.cdc.Unmarshal(iterator.Value(), &planID)
		if err != nil {
			panic(err)
		}
		plan, _ := k.GetPlan(ctx, planID.GetValue())
		if cb(plan) {
			break
		}
	}
}

// SetPlanIdByFarmerAddrIndex sets Index by FarmerAddr
// TODO: need to gas cost check for existing check or update everytime
func (k Keeper) SetPlanIdByFarmerAddrIndex(ctx sdk.Context, farmerAcc sdk.AccAddress, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: planID})
	store.Set(types.GetPlanByFarmerAddrIndexKey(farmerAcc, planID), b)
}

// GetNextPlanIDWithUpdate returns and increments the global Plan ID counter.
// If the global plan number is not set, it initializes it with value 1.
func (k Keeper) GetNextPlanIDWithUpdate(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GlobalPlanIdKey)
	if bz == nil {
		// initialize the PlanId
		id = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

		id = val.GetValue()
	}
	bz = k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id + 1})
	store.Set(types.GlobalPlanIdKey, bz)
	return id
}

func (k Keeper) decodePlan(bz []byte) types.PlanI {
	acc, err := k.UnmarshalPlan(bz)
	if err != nil {
		panic(err)
	}

	return acc
}

// MarshalPlan protobuf serializes an Plan interface
func (k Keeper) MarshalPlan(plan types.PlanI) ([]byte, error) { // nolint:interfacer
	return k.cdc.MarshalInterface(plan)
}

// UnmarshalPlan returns an Plan interface from raw encoded plan
// bytes of a Proto-based Plan type
func (k Keeper) UnmarshalPlan(bz []byte) (plan types.PlanI, err error) {
	return plan, k.cdc.UnmarshalInterface(bz, &plan)
}

// CreateFixedAmountPlan sets fixed amount plan.
func (k Keeper) CreateFixedAmountPlan(ctx sdk.Context, msg *types.MsgCreateFixedAmountPlan, typ types.PlanType) error {
	nextId := k.GetNextPlanIDWithUpdate(ctx)
	farmingPoolAddrAcc, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		return err
	}
	terminationAddrAcc := farmingPoolAddrAcc

	params := k.GetParams(ctx)

	balances := k.bankKeeper.GetAllBalances(ctx, farmingPoolAddrAcc)
	_, hasNeg := balances.SafeSub(params.PrivatePlanCreationFee)
	if hasNeg {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "insufficient balance to pay private plan creation fee")
	}

	farmingFeeCollectorAcc, err := sdk.AccAddressFromBech32(params.FarmingFeeCollector)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoins(ctx, farmingPoolAddrAcc, farmingFeeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
		return err
	}

	basePlan := types.NewBasePlan(
		nextId,
		typ,
		farmingPoolAddrAcc.String(),
		terminationAddrAcc.String(),
		msg.StakingCoinWeights,
		msg.StartTime,
		msg.EndTime,
	)

	fixedPlan := types.NewFixedAmountPlan(basePlan, msg.EpochAmount)

	k.SetPlan(ctx, fixedPlan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFixedAmountPlan,
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, msg.FarmingPoolAddress),
			sdk.NewAttribute(types.AttributeKeyRewardPoolAddress, fixedPlan.RewardPoolAddress),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochAmount, msg.EpochAmount.String()),
		),
	})

	return nil
}

// CreateRatioPlan sets ratio plan.
func (k Keeper) CreateRatioPlan(ctx sdk.Context, msg *types.MsgCreateRatioPlan, typ types.PlanType) error {
	nextId := k.GetNextPlanIDWithUpdate(ctx)
	farmingPoolAddrAcc, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		return err
	}
	terminationAddrAcc := farmingPoolAddrAcc

	params := k.GetParams(ctx)

	balances := k.bankKeeper.GetAllBalances(ctx, farmingPoolAddrAcc)
	_, hasNeg := balances.SafeSub(params.PrivatePlanCreationFee)
	if hasNeg {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "insufficient balance to pay private plan creation fee")
	}

	farmingFeeCollectorAcc, err := sdk.AccAddressFromBech32(params.FarmingFeeCollector)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoins(ctx, farmingPoolAddrAcc, farmingFeeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
		return err
	}

	basePlan := types.NewBasePlan(
		nextId,
		typ,
		farmingPoolAddrAcc.String(),
		terminationAddrAcc.String(),
		msg.StakingCoinWeights,
		msg.StartTime,
		msg.EndTime,
	)

	ratioPlan := types.NewRatioPlan(basePlan, msg.EpochRatio)

	k.SetPlan(ctx, ratioPlan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateRatioPlan,
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, msg.FarmingPoolAddress),
			sdk.NewAttribute(types.AttributeKeyRewardPoolAddress, ratioPlan.RewardPoolAddress),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochRatio, fmt.Sprint(msg.EpochRatio)),
		),
	})

	return nil
}
