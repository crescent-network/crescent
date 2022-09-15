package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (k Keeper) GetLastPlanId(ctx sdk.Context) (id uint64, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPlanIdKey)
	if bz == nil {
		return
	}
	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetLastPlanId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastPlanIdKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) GetPlan(ctx sdk.Context, id uint64) (plan types.Plan, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPlanKey(id))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &plan)
	return plan, true
}

func (k Keeper) SetPlan(ctx sdk.Context, plan types.Plan) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPlanKey(plan.Id), k.cdc.MustMarshal(&plan))
}

func (k Keeper) IterateAllPlans(ctx sdk.Context, cb func(plan types.Plan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PlanKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var plan types.Plan
		k.cdc.MustUnmarshal(iter.Value(), &plan)
		if cb(plan) {
			break
		}
	}
}

func (k Keeper) GetFarm(ctx sdk.Context, denom string) (farm types.Farm, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFarmKey(denom))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &farm)
	return farm, true
}

func (k Keeper) SetFarm(ctx sdk.Context, denom string, farm types.Farm) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFarmKey(denom), k.cdc.MustMarshal(&farm))
}

func (k Keeper) IterateAllFarms(ctx sdk.Context, cb func(denom string, farm types.Farm) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FarmKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom := types.ParseFarmKey(iter.Key())
		var farm types.Farm
		k.cdc.MustUnmarshal(iter.Value(), &farm)
		if cb(denom, farm) {
			break
		}
	}
}

func (k Keeper) GetPosition(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (position types.Position, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPositionKey(farmerAddr, denom))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &position)
	return position, true
}

func (k Keeper) SetPosition(ctx sdk.Context, position types.Position) {
	store := ctx.KVStore(k.storeKey)
	farmerAddr, err := sdk.AccAddressFromBech32(position.Farmer)
	if err != nil {
		panic(err)
	}
	store.Set(types.GetPositionKey(farmerAddr, position.Denom), k.cdc.MustMarshal(&position))
}

func (k Keeper) IterateAllPositions(ctx sdk.Context, cb func(position types.Position) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PositionKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var position types.Position
		k.cdc.MustUnmarshal(iter.Value(), &position)
		if cb(position) {
			break
		}
	}
}

func (k Keeper) GetHistoricalRewards(ctx sdk.Context, denom string, period uint64) (hist types.HistoricalRewards, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoricalRewardsKey(denom, period))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &hist)
	return hist, true
}

func (k Keeper) SetHistoricalRewards(ctx sdk.Context, denom string, period uint64, hist types.HistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetHistoricalRewardsKey(denom, period), k.cdc.MustMarshal(&hist))
}

func (k Keeper) DeleteHistoricalRewards(ctx sdk.Context, denom string, period uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetHistoricalRewardsKey(denom, period))
}

func (k Keeper) IterateAllHistoricalRewards(ctx sdk.Context, cb func(denom string, period uint64, hist types.HistoricalRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.HistoricalRewardsKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom, period := types.ParseHistoricalRewardsKey(iter.Key())
		var hist types.HistoricalRewards
		k.cdc.MustUnmarshal(iter.Value(), &hist)
		if cb(denom, period, hist) {
			break
		}
	}
}
