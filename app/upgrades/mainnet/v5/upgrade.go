package v5

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	liquiditykeeper "github.com/crescent-network/crescent/v4/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

const UpgradeName = "v5"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator,
	liquidityKeeper liquiditykeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		liquidityKeeper.SetMaxNumMarketMakingOrdersPerPair(ctx, liquiditytypes.DefaultMaxNumMarketMakingOrdersPerPair)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

var StoreUpgrades = store.StoreUpgrades{}
