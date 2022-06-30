package keeper

import (
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// GetLastEpochTime returns the last time the epoch ended.
func (k Keeper) GetLastEpochTime(ctx sdk.Context) (t time.Time, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastEpochTimeKey)
	if bz == nil {
		return
	}
	var ts gogotypes.Timestamp
	k.cdc.MustUnmarshal(bz, &ts)
	var err error
	t, err = gogotypes.TimestampFromProto(&ts)
	if err != nil {
		panic(err)
	}
	found = true
	return
}

// SetLastEpochTime sets the last time the epoch ended.
func (k Keeper) SetLastEpochTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	ts, err := gogotypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	bz := k.cdc.MustMarshal(ts)
	store.Set(types.LastEpochTimeKey, bz)
}

// AdvanceEpoch allocates rewards and advances epoch by one.
// AdvanceEpoch also forcefully makes queued stakings to be staked.
// Use this only for simulation and testing purpose.
func (k Keeper) AdvanceEpoch(ctx sdk.Context) error {
	currentEpochDays := k.GetCurrentEpochDays(ctx)
	k.ProcessQueuedCoins(ctx, ctx.BlockTime().Add(time.Duration(currentEpochDays)*types.Day))
	if err := k.TerminateEndedPlans(ctx); err != nil {
		return err
	}
	if err := k.AllocateRewards(ctx); err != nil {
		return err
	}
	k.SetLastEpochTime(ctx, ctx.BlockTime())
	if params := k.GetParams(ctx); params.NextEpochDays != currentEpochDays {
		k.SetCurrentEpochDays(ctx, params.NextEpochDays)
	}

	return nil
}

// GetCurrentEpochDays returns the current epoch days(period).
func (k Keeper) GetCurrentEpochDays(ctx sdk.Context) uint32 {
	var epochDays uint32
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CurrentEpochDaysKey)
	if bz == nil {
		// initialize with next epoch days
		epochDays = k.GetParams(ctx).NextEpochDays
	} else {
		val := gogotypes.UInt32Value{}
		if err := k.cdc.Unmarshal(bz, &val); err != nil {
			panic(err)
		}
		epochDays = val.GetValue()
	}
	return epochDays
}

// SetCurrentEpochDays sets the current epoch days(period).
func (k Keeper) SetCurrentEpochDays(ctx sdk.Context, epochDays uint32) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt32Value{Value: epochDays})
	store.Set(types.CurrentEpochDaysKey, bz)
}
