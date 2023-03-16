package v4

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
)

const UpgradeName = "v4"

func UpgradeHandler(mm *module.Manager, configurator module.Configurator, icaModule ica.AppModule) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = icaModule.ConsensusVersion()

		// Create ICS27 Host submodule params
		icaHostParams := icahosttypes.Params{
			HostEnabled: false,
			AllowMessages: []string{
				// bank module
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				// staking module
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				// distribution module
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				// liquidity module
				sdk.MsgTypeURL(&liquiditytypes.MsgLimitOrder{}),
				sdk.MsgTypeURL(&liquiditytypes.MsgMarketOrder{}),
			},
		}
		// Pass empty controller params since we don't use ica controller submodule
		icaModule.InitModule(ctx, icacontrollertypes.Params{}, icaHostParams)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// Add store upgrades for new modules
var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		icahosttypes.StoreKey,
	},
}
