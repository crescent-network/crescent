package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/x/mint/legacy/v2"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := app.MakeTestEncodingConfig()
	key := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(key, tKey)
	paramSpace := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, key, tKey, types.ModuleName)

	// Check no params
	require.False(t, paramSpace.Has(ctx, types.KeyMintPoolAddress))

	// Run migrations.
	paramSpace.WithKeyTable(types.ParamKeyTable())
	err := v2.MigrateStore(ctx, paramSpace)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramSpace.Has(ctx, types.KeyMintPoolAddress))
}
