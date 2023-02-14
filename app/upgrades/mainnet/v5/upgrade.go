package v5

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	markertypes "github.com/crescent-network/crescent/v4/x/marker/types"
)

const UpgradeName = "v5"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

var StoreUpgrades = store.StoreUpgrades{
	Added: []string{markertypes.StoreKey},
}
