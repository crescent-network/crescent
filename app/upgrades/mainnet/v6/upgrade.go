package v6

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	ammkeeper "github.com/crescent-network/crescent/v5/x/amm/keeper"
	claimkeeper "github.com/crescent-network/crescent/v5/x/claim/keeper"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	farmingkeeper "github.com/crescent-network/crescent/v5/x/farming/keeper"
	liquiditykeeper "github.com/crescent-network/crescent/v5/x/liquidity/keeper"
	lpfarmkeeper "github.com/crescent-network/crescent/v5/x/lpfarm/keeper"
	markerkeeper "github.com/crescent-network/crescent/v5/x/marker/keeper"
)

const UpgradeName = "v6"

var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		evmtypes.StoreKey,
		feemarkettypes.StoreKey,
		// TODO: add erc20
		//erc20types.StoreKey
	},
}

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper, distrKeeper distrkeeper.Keeper, liquidityKeeper liquiditykeeper.Keeper,
	lpFarmKeeper lpfarmkeeper.Keeper, exchangeKeeper exchangekeeper.Keeper, ammKeeper ammkeeper.Keeper,
	markerKeeper markerkeeper.Keeper, farmingKeeper farmingkeeper.Keeper, claimKeeper claimkeeper.Keeper,
	disableUpgradeEvents bool) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// TODO: Set evm, fee market params

		// TODO: migration account type if needed

		return vm, nil
	}
}
