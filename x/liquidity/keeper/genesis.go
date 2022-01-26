package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
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
	}
	for _, pool := range genState.Pools {
		k.SetPool(ctx, pool)
	}
	for _, req := range genState.DepositRequests {
		k.SetDepositRequest(ctx, req)
	}
	for _, req := range genState.WithdrawRequests {
		k.SetWithdrawRequest(ctx, req)
	}
	for _, req := range genState.SwapRequests {
		k.SetSwapRequest(ctx, req)
	}
	for _, req := range genState.CancelSwapRequests {
		k.SetCancelSwapRequest(ctx, req)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var pairs []types.Pair
	k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool) {
		pairs = append(pairs, pair)
		return false
	})
	var pools []types.Pool
	k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
		pools = append(pools, pool)
		return false
	})
	var depositReqs []types.DepositRequest
	k.IterateAllDepositRequests(ctx, func(req types.DepositRequest) (stop bool) {
		depositReqs = append(depositReqs, req)
		return false
	})
	var withdrawReqs []types.WithdrawRequest
	k.IterateAllWithdrawRequests(ctx, func(req types.WithdrawRequest) (stop bool) {
		withdrawReqs = append(withdrawReqs, req)
		return false
	})
	var swapReqs []types.SwapRequest
	k.IterateAllSwapRequests(ctx, func(req types.SwapRequest) (stop bool) {
		swapReqs = append(swapReqs, req)
		return false
	})
	var cancelSwapReqs []types.CancelSwapRequest
	k.IterateAllCancelSwapRequests(ctx, func(req types.CancelSwapRequest) (stop bool) {
		cancelSwapReqs = append(cancelSwapReqs, req)
		return false
	})
	return &types.GenesisState{
		Params:             k.GetParams(ctx),
		LastPairId:         k.GetLastPairId(ctx),
		LastPoolId:         k.GetLastPoolId(ctx),
		Pairs:              pairs,
		Pools:              pools,
		DepositRequests:    depositReqs,
		WithdrawRequests:   withdrawReqs,
		SwapRequests:       swapReqs,
		CancelSwapRequests: cancelSwapReqs,
	}
}
