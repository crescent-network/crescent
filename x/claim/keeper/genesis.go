package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// TODO: mint source coins to the source address

	for _, a := range genState.Airdrops {
		_, found := k.GetAirdrop(ctx, a.AirdropId)
		if found {
			panic("airdrop already exists")
		}
		k.SetAirdrop(ctx, a)
	}

	for _, r := range genState.ClaimRecords {
		k.SetClaimRecord(ctx, r)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	airdrops := k.GetAllAirdrops(ctx)

	records := []types.ClaimRecord{}
	for _, a := range airdrops {
		records = append(records, k.GetAllClaimRecords(ctx, a.AirdropId)...)
	}

	return &types.GenesisState{
		Airdrops:     airdrops,
		ClaimRecords: records,
	}
}
