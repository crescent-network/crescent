package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/marker/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.LastBlockTime != nil {
		k.SetLastBlockTime(ctx, *genState.LastBlockTime)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var lastBlockTimePtr *time.Time
	lastBlockTime, found := k.GetLastBlockTime(ctx)
	if found {
		lastBlockTimePtr = &lastBlockTime
	}
	return types.NewGenesisState(k.GetParams(ctx), lastBlockTimePtr)
}
