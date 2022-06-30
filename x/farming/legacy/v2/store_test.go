package v2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	v1farming "github.com/crescent-network/crescent/v2/x/farming/legacy/v1"
	v2farming "github.com/crescent-network/crescent/v2/x/farming/legacy/v2"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func TestMigrateQueuedStaking(t *testing.T) {
	now := utils.ParseTime("2022-01-01T09:00:00Z")

	encCfg := chain.MakeTestEncodingConfig()
	storeKey := sdk.NewKVStoreKey(v1farming.ModuleName)

	for _, currentEpochDays := range []uint32{1, 2, 3} {
		t.Run("", func(t *testing.T) {
			ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))
			ctx = ctx.WithBlockTime(now)
			store := ctx.KVStore(storeKey)

			stakingCoinDenom := "denom1"
			farmerAcc := sdk.AccAddress(crypto.AddressHash([]byte("farmer")))

			queuedStaking := types.QueuedStaking{Amount: sdk.NewInt(1000000)}
			bz := encCfg.Marshaler.MustMarshal(&queuedStaking)

			oldKey := v1farming.GetQueuedStakingKey(stakingCoinDenom, farmerAcc)
			store.Set(oldKey, bz)

			require.NoError(t, v2farming.MigrateStore(ctx, storeKey, currentEpochDays))

			newKey := types.GetQueuedStakingKey(now.AddDate(0, 0, int(currentEpochDays)), stakingCoinDenom, farmerAcc)
			require.Equal(t, bz, store.Get(newKey))
			require.False(t, store.Has(oldKey))
		})
	}
}

func TestMigrateQueuedStakingIndex(t *testing.T) {
	now := utils.ParseTime("2022-01-01T09:00:00Z")

	storeKey := sdk.NewKVStoreKey(v1farming.ModuleName)

	for _, currentEpochDays := range []uint32{1, 2, 3} {
		ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))
		ctx = ctx.WithBlockTime(now)
		store := ctx.KVStore(storeKey)

		stakingCoinDenom := "denom1"
		farmerAcc := sdk.AccAddress(crypto.AddressHash([]byte("farmer")))

		oldKey := v1farming.GetQueuedStakingIndexKey(farmerAcc, stakingCoinDenom)
		store.Set(oldKey, []byte{})

		require.NoError(t, v2farming.MigrateStore(ctx, storeKey, currentEpochDays))

		newKey := types.GetQueuedStakingIndexKey(farmerAcc, stakingCoinDenom, now.AddDate(0, 0, int(currentEpochDays)))
		require.Equal(t, []byte{}, store.Get(newKey))
		require.False(t, store.Has(oldKey))
	}
}
