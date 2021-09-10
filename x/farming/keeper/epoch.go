package keeper

import (
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (k Keeper) GetLastEpochTime(ctx sdk.Context) (time.Time, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GlobalLastEpochTimeKey)
	if bz == nil {
		return time.Time{}, false
	}
	var ts gogotypes.Timestamp
	k.cdc.MustUnmarshal(bz, &ts)
	t, err := gogotypes.TimestampFromProto(&ts)
	if err != nil {
		panic(err)
	}
	return t, true
}

func (k Keeper) SetLastEpochTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	ts, err := gogotypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	bz := k.cdc.MustMarshal(ts)
	store.Set(types.GlobalLastEpochTimeKey, bz)
}

func (k Keeper) AdvanceEpoch(ctx sdk.Context) error {
	if err := k.AllocateRewards(ctx); err != nil {
		return err
	}
	k.ProcessQueuedCoins(ctx)
	k.SetLastEpochTime(ctx, ctx.BlockTime())

	return nil
}
