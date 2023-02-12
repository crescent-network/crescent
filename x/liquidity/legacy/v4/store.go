package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v4/x/liquidity/types"
)

func DeleteMMOrderIndexes(store sdk.KVStore) {
	iter := sdk.KVStorePrefixIterator(store, MMOrderIndexKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, paramSpace paramstypes.Subspace) error {
	store := ctx.KVStore(storeKey)
	DeleteMMOrderIndexes(store)
	paramSpace.Set(ctx, types.KeyMaxNumMarketMakingOrdersPerPair, uint32(types.DefaultMaxNumMarketMakingOrdersPerPair))
	return nil
}
