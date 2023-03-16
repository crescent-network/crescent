package v4_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	v4liquidity "github.com/crescent-network/crescent/v5/x/liquidity/legacy/v4"
	"github.com/crescent-network/crescent/v5/x/liquidity/types"
)

func TestMigrateStore(t *testing.T) {
	encCfg := chain.MakeTestEncodingConfig()
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)
	paramSpace := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, storeKey, tKey, types.ModuleName)
	paramSpace.WithKeyTable(types.ParamKeyTable())

	ordererAddr := utils.TestAddress(0)
	// We're setting dummy value cause only the existence of the key is important.
	store.Set(v4liquidity.GetMMOrderIndexKey(ordererAddr, 1), []byte("foo"))

	require.NoError(t, v4liquidity.MigrateStore(ctx, storeKey, paramSpace))

	require.Nil(t, store.Get(v4liquidity.GetMMOrderIndexKey(ordererAddr, 1)))
}
