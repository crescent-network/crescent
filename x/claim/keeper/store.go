package keeper

import (
	"time"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// GetLastAirdropId returns the last airdrop id.
func (k Keeper) GetLastAirdropId(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastAirdropIdKey)
	if bz == nil {
		id = 0 // initialize the airdrop id
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

// SetAirdropId stores the last airdrop id.
func (k Keeper) SetAirdropId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.LastAirdropIdKey, bz)
}

// GetStartTime returns the start time for the airdrop.
func (k Keeper) GetStartTime(ctx sdk.Context, airdropId uint64) *time.Time {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStartTimeKey(airdropId))
	if bz == nil {
		return nil
	} else {
		ts, err := sdk.ParseTimeBytes(bz)
		if err != nil {
			panic(err)
		}
		return &ts
	}
}

// SetStartTime stores the start time for the airdrop with start time key
func (k Keeper) SetStartTime(ctx sdk.Context, airdropId uint64, startTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.FormatTimeBytes(startTime)
	store.Set(types.GetStartTimeKey(airdropId), bz)
}

// GetEndTime returns the end time for the airdrop.
func (k Keeper) GetEndTime(ctx sdk.Context, airdropId uint64) *time.Time {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetEndTimeKey(airdropId))
	if bz == nil {
		return nil
	} else {
		ts, err := sdk.ParseTimeBytes(bz)
		if err != nil {
			panic(err)
		}
		return &ts
	}
}

// SetEndTime stores the end time for the airdrop with end time key.
func (k Keeper) SetEndTime(ctx sdk.Context, airdropId uint64, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.FormatTimeBytes(endTime)
	store.Set(types.GetEndTimeKey(airdropId), bz)
}

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
	k.SetAirdropId(ctx, airdrop.Id)
	k.SetStartTime(ctx, airdrop.Id, airdrop.StartTime)
	k.SetEndTime(ctx, airdrop.Id, airdrop.EndTime)
	store.Set(types.GetAirdropKey(airdrop.Id), bz)
}

// GetClaimRecordByRecipient returns the claim record for the given airdrop id and the recipient address.
func (k Keeper) GetClaimRecordByRecipient(ctx sdk.Context, airdropId uint64, recipient sdk.AccAddress) (record types.ClaimRecord, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetClaimRecordByRecipientKey(airdropId, recipient))
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
	store.Set(types.GetClaimRecordByRecipientKey(record.AirdropId, record.GetRecipient()), bz)
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
	iter := sdk.KVStorePrefixIterator(store, types.GetClaimRecordKey(airdropId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.ClaimRecord
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}
