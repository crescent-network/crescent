package v2

import (
	"bytes"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/crescent-network/crescent/v5/x/liquidfarming/legacy/v1"
)

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	// Delete all previous states except for LastRewardsAuctionEndTime.
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		if bytes.Equal(iter.Key(), v1.LastRewardsAuctionEndTimeKey) {
			continue
		}
		store.Delete(iter.Key())
	}
	// Migrate NextRewardsAuctionEndTime.
	bz := store.Get(v1.LastRewardsAuctionEndTimeKey)
	if bz != nil {
		var ts gogotypes.Timestamp
		if err := cdc.Unmarshal(bz, &ts); err != nil {
			return err
		}
		endTime, err := gogotypes.TimestampFromProto(&ts)
		if err != nil {
			return err
		}
		store.Set(NextRewardsAuctionEndTimeKey, sdk.FormatTimeBytes(endTime))
	}
	// TODO: Delete old parameters
	return nil
}
