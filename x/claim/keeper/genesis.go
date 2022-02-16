package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	totalClaimableCoinsMap := make(map[uint64]sdk.Coins) // map(airdropId => totalClaimableCoins)
	for _, r := range genState.ClaimRecords {
		totalClaimableCoinsMap[r.AirdropId] = totalClaimableCoinsMap[r.AirdropId].Add(r.ClaimableCoins...)

		k.SetClaimRecord(ctx, r)
	}

	for _, airdrop := range genState.Airdrops {
		_, found := k.GetAirdrop(ctx, airdrop.AirdropId)
		if !found {
			k.SetAirdrop(ctx, airdrop)
		}

		// The source account balances must be greater than or equal to the total claimable amounts
		balances := k.bankKeeper.GetAllBalances(ctx, airdrop.GetSourceAddress())
		if !balances.IsAllGTE(totalClaimableCoinsMap[airdrop.AirdropId]) {
			panic("source account balances must be equal to total claimable coins")
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	airdrops := k.GetAllAirdrops(ctx)

	records := []types.ClaimRecord{}
	for _, airdrop := range airdrops {
		records = append(records, k.GetAllClaimRecordsByAirdropId(ctx, airdrop.AirdropId)...)
	}

	return &types.GenesisState{
		Airdrops:     airdrops,
		ClaimRecords: records,
	}
}
