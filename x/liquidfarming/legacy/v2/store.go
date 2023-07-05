package v2

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/crescent-network/crescent/v5/x/liquidfarming/legacy/v1"
)

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey) error {
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
	return nil
}
