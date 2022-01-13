package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the liquidity module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.Keeper.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// Pools queries all pools.
func (k Querier) Pools(c context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPoolsResponse{}, nil
}

// ...
func (k Querier) PoolsByPair(c context.Context, req *types.QueryPoolsByPairRequest) (*types.QueryPoolsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPoolsResponse{}, nil
}

// Pool queries all pool.
func (k Querier) Pool(c context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPoolResponse{}, nil
}

// ...
func (k Querier) PoolByReserveAcc(c context.Context, req *types.QueryPoolByReserveAccRequest) (*types.QueryPoolResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPoolResponse{}, nil
}

// ...
func (k Querier) PoolByPoolCoinDenom(c context.Context, req *types.QueryPoolByPoolCoinDenomRequest) (*types.QueryPoolResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPoolResponse{}, nil
}

// ...
func (k Querier) Pairs(c context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPairsResponse{}, nil
}

// ...
func (k Querier) Pair(c context.Context, req *types.QueryPairRequest) (*types.QueryPairResponse, error) {
	// TODO: not implemented yet
	return &types.QueryPairResponse{}, nil
}

// ...
func (k Querier) Deposits(c context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryDepositsResponse{}, nil
}

// ...
func (k Querier) Deposit(c context.Context, req *types.QueryDepositRequest) (*types.QueryDepositResponse, error) {
	// TODO: not implemented yet
	return &types.QueryDepositResponse{}, nil
}

// ...
func (k Querier) Withdraws(c context.Context, req *types.QueryWithdrawsRequest) (*types.QueryWithdrawsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryWithdrawsResponse{}, nil
}

// ...
func (k Querier) Withdraw(c context.Context, req *types.QueryWithdrawRequest) (*types.QueryWithdrawResponse, error) {
	// TODO: not implemented yet
	return &types.QueryWithdrawResponse{}, nil
}

// ...
func (k Querier) Swaps(c context.Context, req *types.QuerySwapsRequest) (*types.QuerySwapsResponse, error) {
	// TODO: not implemented yet
	return &types.QuerySwapsResponse{}, nil
}

// ...
func (k Querier) Swap(c context.Context, req *types.QuerySwapRequest) (*types.QuerySwapResponse, error) {
	// TODO: not implemented yet
	return &types.QuerySwapResponse{}, nil
}
