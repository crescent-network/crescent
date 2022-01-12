package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/mint/keeper"
	"github.com/crescent-network/crescent/x/mint/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetParams(ctx, data.Params)
	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	params := keeper.GetParams(ctx)
	return types.NewGenesisState(params)
}
