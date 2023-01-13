package v5

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const UpgradeName = "v5"

func UpgradeHandler(mm *module.Manager, configurator module.Configurator, wasmKeeper wasm.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		newVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return newVM, err
		}

		// Set the code upload access permission to nobody for permissioned contract platform
		// Note that governance proposals bypass the existing authorization policy
		params := wasmKeeper.GetParams(ctx)
		params.CodeUploadAccess = wasmtypes.AllowNobody
		wasmKeeper.SetParams(ctx, params)

		return newVM, nil
	}
}

// Add store upgrades for new modules
var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		wasm.StoreKey,
	},
}
