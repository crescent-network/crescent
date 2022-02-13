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
	for _, req := range genState.Orders {
		k.SetOrder(ctx, req)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:           k.GetParams(ctx),
		LastPairId:       k.GetLastPairId(ctx),
		LastPoolId:       k.GetLastPoolId(ctx),
		Pairs:            k.GetAllPairs(ctx),
		Pools:            k.GetAllPools(ctx),
		DepositRequests:  k.GetAllDepositRequests(ctx),
		WithdrawRequests: k.GetAllWithdrawRequests(ctx),
		Orders:           k.GetAllOrders(ctx),
	}
}
