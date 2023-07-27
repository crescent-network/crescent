package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.LastPublicPositionId > 0 {
		k.SetLastPublicPositionId(ctx, genState.LastPublicPositionId)
	}
	for _, publicPosition := range genState.PublicPositions {
		k.SetPublicPosition(ctx, publicPosition)
	}
	for _, auction := range genState.RewardsAuctions {
		k.SetRewardsAuction(ctx, auction)
	}
	for _, bid := range genState.Bids {
		k.SetBid(ctx, bid)
	}
	if genState.NextRewardsAuctionEndTime != nil {
		k.SetNextRewardsAuctionEndTime(ctx, *genState.NextRewardsAuctionEndTime)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var nextAuctionEndTime *time.Time
	if t, found := k.GetNextRewardsAuctionEndTime(ctx); found {
		nextAuctionEndTime = &t
	}
	return types.NewGenesisState(
		k.GetParams(ctx), k.GetLastPublicPositionId(ctx),
		k.GetAllPublicPositions(ctx), k.GetAllRewardsAuctions(ctx),
		k.GetAllBids(ctx), nextAuctionEndTime)
}
