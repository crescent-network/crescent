package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// GetClaimRecord returns the types.ClaimRecord for given recipient.
func (k Keeper) GetClaimRecord(ctx sdk.Context, recipient sdk.AccAddress) (record types.ClaimRecord, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetClaimRecordKey(recipient))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &record)
	return record, true
}

// SetClaimRecord stores a types.ClaimRecord.
func (k Keeper) SetClaimRecord(ctx sdk.Context, record types.ClaimRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.GetClaimRecordKey(record.GetAddress()), bz)
}

// GetAllClaimRecords returns all types.ClaimRecord stored.
func (k Keeper) GetAllClaimRecords(ctx sdk.Context) (records []types.ClaimRecord) {
	k.IterateAllClaimRecords(ctx, func(record types.ClaimRecord) (stop bool) {
		records = append(records, record)
		return false
	})
	return
}

// IterateAllClaimRecords iterates over all types.ClaimRecord stored.
func (k Keeper) IterateAllClaimRecords(ctx sdk.Context, cb func(record types.ClaimRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ClaimRecordKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.ClaimRecord
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// DeleteClaimRecord deletes a types.ClaimRecord.
func (k Keeper) DeleteClaimRecord(ctx sdk.Context, recipient sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetClaimRecordKey(recipient))
}
