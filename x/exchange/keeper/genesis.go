package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.LastSpotMarketId > 0 {
		k.SetLastSpotMarketId(ctx, genState.LastSpotMarketId)
	}
	if genState.LastSpotOrderId > 0 {
		k.SetLastSpotOrderId(ctx, genState.LastSpotOrderId)
	}
	for _, marketRecord := range genState.SpotMarketRecords {
		k.SetSpotMarket(ctx, marketRecord.Market)
		k.SetSpotMarketByDenomsIndex(ctx, marketRecord.Market)
		k.SetSpotMarketState(ctx, marketRecord.Market.Id, marketRecord.State)
	}
	for _, order := range genState.SpotOrders {
		k.SetSpotOrder(ctx, order)
		k.SetSpotOrderBookOrder(ctx, order)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	marketRecords := []types.SpotMarketRecord{}
	k.IterateAllSpotMarkets(ctx, func(market types.SpotMarket) (stop bool) {
		marketRecords = append(marketRecords, types.SpotMarketRecord{
			Market: market,
			State:  k.MustGetSpotMarketState(ctx, market.Id),
		})
		return false
	})
	orders := []types.SpotOrder{}
	k.IterateAllSpotOrders(ctx, func(order types.SpotOrder) (stop bool) {
		orders = append(orders, order)
		return false
	})
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetLastSpotMarketId(ctx),
		k.GetLastSpotOrderId(ctx),
		marketRecords,
		orders)
}
