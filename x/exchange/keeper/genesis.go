package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.LastMarketId > 0 {
		k.SetLastMarketId(ctx, genState.LastMarketId)
	}
	if genState.LastOrderId > 0 {
		k.SetLastOrderId(ctx, genState.LastOrderId)
	}
	for _, marketRecord := range genState.MarketRecords {
		k.SetMarket(ctx, marketRecord.Market)
		k.SetMarketByDenomsIndex(ctx, marketRecord.Market)
		k.SetMarketState(ctx, marketRecord.Market.Id, marketRecord.State)
	}
	for _, order := range genState.Orders {
		k.SetOrder(ctx, order)
		k.SetOrderBookOrderIndex(ctx, order)
		k.SetOrdersByOrdererIndex(ctx, order)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	marketRecords := []types.MarketRecord{}
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		marketRecords = append(marketRecords, types.MarketRecord{
			Market: market,
			State:  k.MustGetMarketState(ctx, market.Id),
		})
		return false
	})
	orders := []types.Order{}
	k.IterateAllOrders(ctx, func(order types.Order) (stop bool) {
		orders = append(orders, order)
		return false
	})
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetLastMarketId(ctx),
		k.GetLastOrderId(ctx),
		marketRecords,
		orders)
}
