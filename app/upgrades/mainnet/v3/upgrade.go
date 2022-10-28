package v3

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	liquidfarmingtypes "github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditykeeper "github.com/crescent-network/crescent/v3/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v3/x/lpfarm/types"
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

		// Set param for new market maker module
		marketmakerparams := marketmakertypes.DefaultParams()
		marketmakerparams.DepositAmount = sdk.NewCoins(sdk.NewCoin("ucre", sdk.NewInt(1000000000)))
		marketmakerkeeper.SetParams(ctx, marketmakerparams)
		return newVM, err

	}
}

// Add new modules
var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		marketmakertypes.ModuleName,
		lpfarmtypes.ModuleName,
		liquidfarmingtypes.ModuleName,
	},
}
