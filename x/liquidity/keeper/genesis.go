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
	for _, record := range genState.Pairs {
		k.SetPair(ctx, record.Pair)
		k.SetPairIndex(ctx, record.Pair.BaseCoinDenom, record.Pair.QuoteCoinDenom, record.Pair.Id)
		k.SetPairLookupIndex(ctx, record.Pair.BaseCoinDenom, record.Pair.QuoteCoinDenom, record.Pair.Id)
		k.SetPairLookupIndex(ctx, record.Pair.QuoteCoinDenom, record.Pair.BaseCoinDenom, record.Pair.Id)
		k.SetPairState(ctx, record.Pair.Id, record.State)
	}
	for _, record := range genState.Pools {
		k.SetPool(ctx, record.Pool)
		k.SetPoolByReserveIndex(ctx, record.Pool)
		k.SetPoolsByPairIndex(ctx, record.Pool)
		k.SetPoolState(ctx, record.Pool.Id, record.State)
	}
	for _, req := range genState.DepositRequests {
		k.SetDepositRequest(ctx, req)
		k.SetDepositRequestIndex(ctx, req)
	}
	for _, req := range genState.WithdrawRequests {
		k.SetWithdrawRequest(ctx, req)
		k.SetWithdrawRequestIndex(ctx, req)
	}
	for _, record := range genState.Orders {
		k.SetOrder(ctx, record.Order)
		k.SetOrderIndex(ctx, record.Order)
		k.SetOrderState(ctx, record.Order.PairId, record.Order.Id, record.State)
	}
	for _, record := range genState.NumMarketMakingOrdersRecords {
		ordererAddr := sdk.MustAccAddressFromBech32(record.Orderer)
		k.SetNumMMOrders(ctx, ordererAddr, record.PairId, record.NumMarketMakingOrders)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	pairRecords := []types.PairRecord{}
	_ = k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool, err error) {
		pairState, _ := k.GetPairState(ctx, pair.Id)
		pairRecords = append(pairRecords, types.PairRecord{
			Pair:  pair,
			State: pairState,
		})
		return false, nil
	})
	poolRecords := []types.PoolRecord{}
	_ = k.IterateAllPools(ctx, func(pool types.Pool) (stop bool, err error) {
		poolState, _ := k.GetPoolState(ctx, pool.Id)
		poolRecords = append(poolRecords, types.PoolRecord{
			Pool:  pool,
			State: poolState,
		})
		return false, nil
	})
	orderRecords := []types.OrderRecord{}
	_ = k.IterateAllOrders(ctx, func(order types.Order) (stop bool, err error) {
		orderState, _ := k.GetOrderState(ctx, order.PairId, order.Id)
		orderRecords = append(orderRecords, types.OrderRecord{
			Order: order,
			State: orderState,
		})
		return false, nil
	})
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
		Pairs:                        pairRecords,
		Pools:                        poolRecords,
		DepositRequests:              k.GetAllDepositRequests(ctx),
		WithdrawRequests:             k.GetAllWithdrawRequests(ctx),
		Orders:                       orderRecords,
		NumMarketMakingOrdersRecords: numMMOrdersRecords,
	}
}
