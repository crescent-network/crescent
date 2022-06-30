package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/mint/keeper"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	if err := types.ValidateGenesis(*data); err != nil {
		panic(err)
	}
	// init to prevent nil slice, []types.InflationSchedule(nil)
	if data.Params.InflationSchedules == nil || len(data.Params.InflationSchedules) == 0 {
		data.Params.InflationSchedules = []types.InflationSchedule{}
	}
	keeper.SetParams(ctx, data.Params)
	if data.LastBlockTime != nil {
		keeper.SetLastBlockTime(ctx, *data.LastBlockTime)
	}
	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	lastBlockTime := keeper.GetLastBlockTime(ctx)
	params := keeper.GetParams(ctx)
	// init to prevent nil slice, []types.InflationSchedule(nil)
	if params.InflationSchedules == nil || len(params.InflationSchedules) == 0 {
		params.InflationSchedules = []types.InflationSchedule{}
	}
	return types.NewGenesisState(params, lastBlockTime)
}
