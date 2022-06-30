package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/mint/types"
)

// GetLastBlockTime returns the last block time.
func (k Keeper) GetLastBlockTime(ctx sdk.Context) *time.Time {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastBlockTimeKey)
	if bz == nil {
		return nil
	}
	ts, err := sdk.ParseTimeBytes(bz)
	if err != nil {
		panic(err)
	}
	return &ts
}

// SetLastBlockTime stores the last block time.
func (k Keeper) SetLastBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.FormatTimeBytes(blockTime)
	store.Set(types.LastBlockTimeKey, bz)
}
