package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidity/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	k.SetParams(ctx, genState.Params)
	k.SetLastPairId(ctx, genState.LastPairId)
	k.SetLastPoolId(ctx, genState.LastPoolId)
	for _, pair := range genState.Pairs {
		k.SetPair(ctx, pair)
		k.SetPairIndex(ctx, pair.BaseCoinDenom, pair.QuoteCoinDenom, pair.Id)
		k.SetPairLookupIndex(ctx, pair.BaseCoinDenom, pair.QuoteCoinDenom, pair.Id)
		k.SetPairLookupIndex(ctx, pair.QuoteCoinDenom, pair.BaseCoinDenom, pair.Id)
	}
	for _, pool := range genState.Pools {
		k.SetPool(ctx, pool)
		k.SetPoolByReserveIndex(ctx, pool)
		k.SetPoolsByPairIndex(ctx, pool)
	}
	for _, req := range genState.DepositRequests {
		k.SetDepositRequest(ctx, req)
		k.SetDepositRequestIndex(ctx, req)
	}
	for _, req := range genState.WithdrawRequests {
		k.SetWithdrawRequest(ctx, req)
		k.SetWithdrawRequestIndex(ctx, req)
	}
	for _, order := range genState.Orders {
		k.SetOrder(ctx, order)
		k.SetOrderIndex(ctx, order)
	}
	for _, record := range genState.NumMarketMakingOrdersRecords {
		ordererAddr := sdk.MustAccAddressFromBech32(record.Orderer)
		k.SetNumMMOrders(ctx, ordererAddr, record.PairId, record.NumMarketMakingOrders)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	numMMOrdersRecords := []types.NumMMOrdersRecord{}
	k.IterateAllNumMMOrders(ctx, func(ordererAddr sdk.AccAddress, pairId uint64, numMMOrders uint32) (stop bool) {
		numMMOrdersRecords = append(numMMOrdersRecords, types.NumMMOrdersRecord{
			Orderer:               ordererAddr.String(),
			PairId:                pairId,
			NumMarketMakingOrders: numMMOrders,
		})
		return false
	})
	return &types.GenesisState{
		Params:                       k.GetParams(ctx),
		LastPairId:                   k.GetLastPairId(ctx),
		LastPoolId:                   k.GetLastPoolId(ctx),
		Pairs:                        k.GetAllPairs(ctx),
		Pools:                        k.GetAllPools(ctx),
		DepositRequests:              k.GetAllDepositRequests(ctx),
		WithdrawRequests:             k.GetAllWithdrawRequests(ctx),
		Orders:                       k.GetAllOrders(ctx),
		NumMarketMakingOrdersRecords: numMMOrdersRecords,
	}
}
