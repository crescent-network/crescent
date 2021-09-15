package keeper

import (
	"strconv"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

// NewPlan sets the next plan number to a given plan interface
func (k Keeper) NewPlan(ctx sdk.Context, plan types.PlanI) types.PlanI {
	if err := plan.SetId(k.GetNextPlanIdWithUpdate(ctx)); err != nil {
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

// GetNextPlanIdWithUpdate returns and increments the global Plan ID counter.
// If the global plan number is not set, it initializes it with value 0.
func (k Keeper) GetNextPlanIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetGlobalPlanId(ctx) + 1
	k.SetGlobalPlanId(ctx, id)
	return id
}

// SetGlobalPlanId set the global Plan ID counter.
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

		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

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
func (k Keeper) CreateFixedAmountPlan(ctx sdk.Context, msg *types.MsgCreateFixedAmountPlan, farmingPoolAcc, terminationAcc sdk.AccAddress, typ types.PlanType) (types.PlanI, error) {
	nextId := k.GetNextPlanIdWithUpdate(ctx)
	if typ == types.PlanTypePrivate {
		params := k.GetParams(ctx)
		balances := k.bankKeeper.GetAllBalances(ctx, msg.GetCreator())
		diffs, hasNeg := balances.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "lack of %s coins to pay private plan creation fee", diffs.String())
		}

		farmingFeeCollectorAcc, err := sdk.AccAddressFromBech32(params.FarmingFeeCollector)
		if err != nil {
			return nil, err
		}

		if err := k.bankKeeper.SendCoins(ctx, msg.GetCreator(), farmingFeeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
			return nil, err
		}
	}

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
	nextId := k.GetNextPlanIdWithUpdate(ctx)
	if typ == types.PlanTypePrivate {
		params := k.GetParams(ctx)
		balances := k.bankKeeper.GetAllBalances(ctx, msg.GetCreator())
		diffs, hasNeg := balances.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "lack of %s coins to pay private plan createion fee", diffs.String())
		}

		farmingFeeCollectorAcc, err := sdk.AccAddressFromBech32(params.FarmingFeeCollector)
		if err != nil {
			return nil, err
		}

		if err := k.bankKeeper.SendCoins(ctx, msg.GetCreator(), farmingFeeCollectorAcc, params.PrivatePlanCreationFee); err != nil {
			return nil, err
		}
	}

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

// TerminatePlan sends all remaining coins in the plan's farming pool to
// the termination address and mark the plan as terminated.
func (k Keeper) TerminatePlan(ctx sdk.Context, plan types.PlanI) error {
	balances := k.bankKeeper.GetAllBalances(ctx, plan.GetFarmingPoolAddress())
	if balances.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(ctx, plan.GetFarmingPoolAddress(), plan.GetTerminationAddress(), balances); err != nil {
			return err
		}
	}

	if err := plan.SetTerminated(true); err != nil {
		return err
	}
	k.SetPlan(ctx, plan)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePlanTerminated,
			sdk.NewAttribute(types.AttributeKeyPlanId, strconv.FormatUint(plan.GetId(), 10)),
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, plan.GetFarmingPoolAddress().String()),
			sdk.NewAttribute(types.AttributeKeyTerminationAddress, plan.GetTerminationAddress().String()),
			// TODO: add refunded coins?
		),
	})

	return nil
}

func (k Keeper) GeneratePrivatePlanFarmingPoolAddress(ctx sdk.Context, name string) (sdk.AccAddress, error) {
	nextPlanId := k.GetGlobalPlanId(ctx) + 1
	poolAcc := types.PrivatePlanFarmingPoolAddress(name, nextPlanId)
	if !k.bankKeeper.GetAllBalances(ctx, poolAcc).Empty() {
		return nil, types.ErrConflictPrivatePlanFarmingPool
	}
	return poolAcc, nil
}

// ValidateStakingReservedAmount checks that the balance of StakingReserveAcc greater than the amount of staked, queued coins in all staking objects.
func (k Keeper) ValidateStakingReservedAmount(ctx sdk.Context) error {
	var totalStakingCoins sdk.Coins
	k.IterateStakings(ctx, func(stakingCoinDenom string, _ sdk.AccAddress, staking types.Staking) (stop bool) {
		totalStakingCoins = totalStakingCoins.Add(sdk.NewCoin(stakingCoinDenom, staking.Amount))
		return false
	})
	k.IterateQueuedStakings(ctx, func(stakingCoinDenom string, _ sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		totalStakingCoins = totalStakingCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
		return false
	})

	balanceStakingReserveAcc := k.bankKeeper.GetAllBalances(ctx, types.StakingReserveAcc)
	if !balanceStakingReserveAcc.IsAllGTE(totalStakingCoins) {
		return types.ErrInvalidStakingReservedAmount
	}

	return nil
}

// ValidateRemainingRewardsAmount checks that the balance of the RewardPoolAddresses of all plans greater than the total amount of unwithdrawn reward coins in all reward objects
func (k Keeper) ValidateRemainingRewardsAmount(ctx sdk.Context) error {
	var totalRemainingRewards sdk.Coins
	totalBalancesRewardPool := k.bankKeeper.GetAllBalances(ctx, k.GetRewardsReservePoolAcc(ctx))
	// TODO: apply f1 logic
	//k.IterateAllRewards(ctx, func(reward types.Reward) (stop bool) {
	//	totalRemainingRewards = totalRemainingRewards.Add(reward.RewardCoins...)
	//	return false
	//})

	if !totalBalancesRewardPool.IsAllGTE(totalRemainingRewards) {
		return types.ErrInvalidRemainingRewardsAmount
	}
	return nil
}
