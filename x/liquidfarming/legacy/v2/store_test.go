package v2_test

import (
	"testing"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v5/app"
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

	endTime := utils.ParseTime("2023-01-01T08:00:00Z")
	ts, err := gogotypes.TimestampProto(endTime)
	require.NoError(t, err)
	cdc := chain.MakeTestEncodingConfig().Marshaler
	bz := cdc.MustMarshal(ts)
	store.Set(v1liquidfarming.LastRewardsAuctionEndTimeKey, bz)
	require.NoError(t, v2liquidfarming.MigrateStore(ctx, storeKey, cdc))
	require.Equal(t, sdk.FormatTimeBytes(endTime), store.Get(v2liquidfarming.NextRewardsAuctionEndTimeKey))
}
