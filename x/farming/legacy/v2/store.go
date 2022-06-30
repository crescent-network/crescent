package v2

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1farming "github.com/crescent-network/crescent/v2/x/farming/legacy/v1"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func MigrateQueuedStaking(store sdk.KVStore, endTime time.Time) {
	oldStore := prefix.NewStore(store, v1farming.QueuedStakingKeyPrefix)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		fullKey := append(v1farming.QueuedStakingKeyPrefix, oldStoreIter.Key()...)
		stakingCoinDenom, farmerAcc := v1farming.ParseQueuedStakingKey(fullKey)
		store.Set(types.GetQueuedStakingKey(endTime, stakingCoinDenom, farmerAcc), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

func MigrateQueuedStakingIndex(store sdk.KVStore, endTime time.Time) {
	oldStore := prefix.NewStore(store, v1farming.QueuedStakingIndexKeyPrefix)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		fullKey := append(v1farming.QueuedStakingIndexKeyPrefix, oldStoreIter.Key()...)
		farmerAcc, stakingCoinDenom := v1farming.ParseQueuedStakingIndexKey(fullKey)
		store.Set(types.GetQueuedStakingIndexKey(farmerAcc, stakingCoinDenom, endTime), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, currentEpochDays uint32) error {
	store := ctx.KVStore(storeKey)

	endTime := ctx.BlockTime().Add(time.Duration(currentEpochDays) * types.Day)
	MigrateQueuedStaking(store, endTime)
	MigrateQueuedStakingIndex(store, endTime)

	return nil
}
