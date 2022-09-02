package v3

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	liquiditykeeper "github.com/crescent-network/crescent/v3/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	marketmakerkeeper "github.com/crescent-network/crescent/v3/x/marketmaker/keeper"
	marketmakertypes "github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

const UpgradeName = "v3"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, marketmakerkeeper marketmakerkeeper.Keeper, liquiditykeeper liquiditykeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVM, err
		}

		// Set newly added liquidity param
		liquiditykeeper.SetMaxNumMarketMakingOrderTicks(ctx, liquiditytypes.DefaultMaxNumMarketMakingOrderTicks)

		// Set Default param for new market maker module
		marketmakerkeeper.SetParams(ctx, marketmakertypes.DefaultParams())
		return newVM, err

	}
}

var StoreUpgrades = store.StoreUpgrades{
	// Add newly added market maker module
	Added: []string{marketmakertypes.ModuleName},
}
