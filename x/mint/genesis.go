package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/mint/keeper"
	"github.com/cosmosquad-labs/squad/x/mint/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	if err := types.ValidateGenesis(*data); err != nil {
		panic(err)
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
	return types.NewGenesisState(params, lastBlockTime)
}
