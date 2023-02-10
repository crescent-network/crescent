package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DeleteMMOrderIndexes(store sdk.KVStore) {
	iter := sdk.KVStorePrefixIterator(store, MMOrderIndexKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	DeleteMMOrderIndexes(store)
	return nil
}
