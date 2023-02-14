package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/marker/types"
)

// GetLastBlockTime returns the last block time, if present.
func (k Keeper) GetLastBlockTime(ctx sdk.Context) (t time.Time, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastBlockTimeKey)
	if bz == nil {
		return
	}
	t, err := sdk.ParseTimeBytes(bz)
	if err != nil {
		panic(err)
	}
	return t, true
}

// SetLastBlockTime sets the last block time.
func (k Keeper) SetLastBlockTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastBlockTimeKey, sdk.FormatTimeBytes(t))
}
