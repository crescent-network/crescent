package v2_0_0

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	budgetkeeper "github.com/tendermint/budget/x/budget/keeper"
	budgettypes "github.com/tendermint/budget/x/budget/types"

	utils "github.com/crescent-network/crescent/v2/types"
	liquiditykeeper "github.com/crescent-network/crescent/v2/x/liquidity/keeper"
	mintkeeper "github.com/crescent-network/crescent/v2/x/mint/keeper"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
)

const UpgradeName = "v2.0.0"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator,
	mintKeeper mintkeeper.Keeper, budgetKeeper budgetkeeper.Keeper, liquidityKeeper liquiditykeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVM, err
		}

		// upgrade mint pool address to mint module account from fee collector
		mintparams := mintKeeper.GetParams(ctx)
		mintparams.MintPoolAddress = minttypes.MintModuleAcc.String()
		mintKeeper.SetParams(ctx, mintparams)

		// upgrade fee collector source address to mint pool
		budgetparams := budgetKeeper.GetParams(ctx)
		for i, budget := range budgetparams.Budgets {
			if budget.SourceAddress == minttypes.DefaultMintPoolAddress.String() {
				budgetparams.Budgets[i].SourceAddress = minttypes.MintModuleAcc.String()
			}
		}

		// add budget for staking reward and community fund
		genesisTime := utils.ParseTime("2022-04-13T00:00:00Z")
		budgetStakingRewardCommFund := budgettypes.Budget{
			Name:               "budget-staking-reward-and-community-fund",
			Rate:               sdk.MustNewDecFromStr("0.0875"),
			SourceAddress:      minttypes.MintModuleAcc.String(),          // mint module account
			DestinationAddress: minttypes.DefaultMintPoolAddress.String(), // fee collector
			StartTime:          genesisTime,
			EndTime:            genesisTime.AddDate(10, 0, 0),
		}
		budgetparams.Budgets = append([]budgettypes.Budget{budgetStakingRewardCommFund}, budgetparams.Budgets...)
		budgetKeeper.SetParams(ctx, budgetparams)

		// Increment tick precision to 4.
		liquidityParams := liquidityKeeper.GetParams(ctx)
		liquidityParams.TickPrecision = 4
		liquidityKeeper.SetParams(ctx, liquidityParams)

		return newVM, err
	}
}

var StoreUpgrades store.StoreUpgrades
