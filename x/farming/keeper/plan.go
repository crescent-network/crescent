package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	iterator := sdk.KVStorePrefixIterator(store, types.GetPlansByFarmerAddrIndexKey(farmerAcc))

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

// SetPlanIDByFarmerAddrIndex sets Index by FarmerAddr
// TODO: need to gas cost check for existing check or update everytime
func (k Keeper) SetPlanIDByFarmerAddrIndex(ctx sdk.Context, planID uint64, farmerAcc sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: planID})
	store.Set(types.GetPlanByFarmerAddrIndexKey(farmerAcc, planID), b)
}

// CreateFixedAmountPlan sets fixed amount plan.
func (k Keeper) CreateFixedAmountPlan(ctx sdk.Context, msg *types.MsgCreateFixedAmountPlan, typ types.PlanType) *types.FixedAmountPlan {
	nextId := k.GetNextPlanIDWithUpdate(ctx)
	farmingPoolAddr := msg.GetFarmingPoolAddress()
	terminationAddr := farmingPoolAddr

	basePlan := types.NewBasePlan(
		nextId,
		typ,
		farmingPoolAddr,
		terminationAddr,
		msg.GetStakingCoinWeights(),
		msg.StartTime,
		msg.EndTime,
		msg.GetEpochDays(),
	)

	fixedPlan := types.NewFixedAmountPlan(basePlan, msg.EpochAmount)

	k.SetPlan(ctx, fixedPlan)

	return fixedPlan
}

// CreateRatioPlan sets ratio plan.
func (k Keeper) CreateRatioPlan(ctx sdk.Context, msg *types.MsgCreateRatioPlan, typ types.PlanType) *types.RatioPlan {
	nextId := k.GetNextPlanIDWithUpdate(ctx)
	farmingPoolAddr := msg.GetFarmingPoolAddress()
	terminationAddr := farmingPoolAddr

	basePlan := types.NewBasePlan(
		nextId,
		typ,
		farmingPoolAddr,
		terminationAddr,
		msg.GetStakingCoinWeights(),
		msg.StartTime,
		msg.EndTime,
		msg.GetEpochDays(),
	)

	ratioPlan := types.NewRatioPlan(basePlan, msg.EpochRatio)

	k.SetPlan(ctx, ratioPlan)

	return ratioPlan
}
