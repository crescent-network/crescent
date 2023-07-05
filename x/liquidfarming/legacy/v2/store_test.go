package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	v1liquidfarming "github.com/crescent-network/crescent/v5/x/liquidfarming/legacy/v1"
	v2liquidfarming "github.com/crescent-network/crescent/v5/x/liquidfarming/legacy/v2"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestMigrateStore(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	endTimeBytes := sdk.FormatTimeBytes(utils.ParseTime("2023-01-01T08:00:00Z"))
	store.Set(v1liquidfarming.LastRewardsAuctionEndTimeKey, endTimeBytes)
	require.NoError(t, v2liquidfarming.MigrateStore(ctx, storeKey))
	require.Equal(t, endTimeBytes, store.Get(v2liquidfarming.NextRewardsAuctionEndTimeKey))
}
