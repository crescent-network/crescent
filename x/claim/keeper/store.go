package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// GetAirdrop returns the airdrop object from the airdrop id.
func (k Keeper) GetAirdrop(ctx sdk.Context, airdropId uint64) (airdrop types.Airdrop, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetAirdropKey(airdropId))
	if bz == nil {
		return airdrop, false
	}
	k.cdc.MustUnmarshal(bz, &airdrop)
	return airdrop, true
}

// SetAirdrop sets start and end times and stores the airdrop.
func (k Keeper) SetAirdrop(ctx sdk.Context, airdrop types.Airdrop) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&airdrop)
	store.Set(types.GetAirdropKey(airdrop.Id), bz)
}

// GetClaimRecordByRecipient returns the claim record for the given airdrop id and the recipient address.
func (k Keeper) GetClaimRecordByRecipient(ctx sdk.Context, airdropId uint64, recipient sdk.AccAddress) (record types.ClaimRecord, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetClaimRecordKey(airdropId, recipient))
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
	store.Set(types.GetClaimRecordKey(record.AirdropId, record.GetRecipient()), bz)
}

// GetAllAirdrops returns all types.Airdrop stored.
func (k Keeper) GetAllAirdrops(ctx sdk.Context) []types.Airdrop {
	airdrops := []types.Airdrop{}
	k.IterateAllAirdrops(ctx, func(airdrop types.Airdrop) (stop bool) {
		airdrops = append(airdrops, airdrop)
		return false
	})
	return airdrops
}

// GetAllClaimRecordsByAirdropId returns all types.ClaimRecord stored.
func (k Keeper) GetAllClaimRecordsByAirdropId(ctx sdk.Context, airdropId uint64) (records []types.ClaimRecord) {
	k.IterateAllClaimRecordsByAirdropId(ctx, airdropId, func(record types.ClaimRecord) (stop bool) {
		records = append(records, record)
		return false
	})
	return
}

func (k Keeper) IterateAllAirdrops(ctx sdk.Context, cb func(airdrop types.Airdrop) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.AirdropKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var airdrop types.Airdrop
		k.cdc.MustUnmarshal(iter.Value(), &airdrop)
		if cb(airdrop) {
			break
		}
	}
}

// IterateAllClaimRecordsByAirdropId iterates over all types.ClaimRecord stored.
func (k Keeper) IterateAllClaimRecordsByAirdropId(ctx sdk.Context, airdropId uint64, cb func(record types.ClaimRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetClaimRecordsByAirdropKeyPrefix(airdropId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.ClaimRecord
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}
