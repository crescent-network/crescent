package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetXDenom() != "" {
		if err := sdk.ValidateDenom(req.XDenom); err != nil {
			return nil, err
		}
	}

	if req.GetYDenom() != "" {
		if err := sdk.ValidateDenom(req.YDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolKeyPrefix)

	pools := []types.Pool{}
	pageRes, err := query.FilteredPaginate(poolStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pool, err := types.UnmarshalPool(k.cdc, value)
		if err != nil {
			return false, err
		}

		if req.GetXDenom() != "" {
			if pool.ReserveCoinDenoms[0] != req.GetXDenom() {
				return false, nil
			}
		}

		if req.GetYDenom() != "" {
			if pool.ReserveCoinDenoms[1] != req.GetYDenom() {
				return false, nil
			}
		}

		if accumulate {
			pools = append(pools, pool)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{Pools: pools, Pagination: pageRes}, nil
}

// PoolsByPair queries all pools that correspond to the pair.
func (k Querier) PoolsByPair(c context.Context, req *types.QueryPoolsByPairRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	poolsStore := prefix.NewStore(store, types.GetPoolsByPairKey(req.PairId))

	pools := []types.Pool{}
	pageRes, err := query.Paginate(poolsStore, req.Pagination, func(key []byte, value []byte) error {
		pool, err := types.UnmarshalPool(k.cdc, value)
		if err != nil {
			return err
		}

		pools = append(pools, pool)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{Pools: pools, Pagination: pageRes}, nil
}

// Pool queries the specific pool.
func (k Querier) Pool(c context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	pool, found := k.GetPool(ctx, req.PoolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool %d doesn't exist", req.PoolId)
	}

	return &types.QueryPoolResponse{Pool: pool}, nil
}

// PoolByReserveAcc queries the specific pool by the reserve account address.
func (k Querier) PoolByReserveAcc(c context.Context, req *types.QueryPoolByReserveAccRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	reserveAcc, err := sdk.AccAddressFromBech32(req.ReserveAcc)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "the reserve account address %s is not valid", req.ReserveAcc)
	}

	pool, found := k.GetPoolByReserveAcc(ctx, reserveAcc)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool by %s doesn't exist", req.ReserveAcc)
	}

	return &types.QueryPoolResponse{Pool: pool}, nil
}

// PoolByPoolCoinDenom queries the specific pool by the pool coin denomination.
func (k Querier) PoolByPoolCoinDenom(c context.Context, req *types.QueryPoolByPoolCoinDenomRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	poolId := types.TrimPoolPrefix(req.PoolCoinDenom)

	pool, found := k.GetPool(ctx, poolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool %d doesn't exist", poolId)
	}

	return &types.QueryPoolResponse{Pool: pool}, nil
}

// Pairs queries all pairs.
func (k Querier) Pairs(c context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.GetXDenom() != "" {
		if err := sdk.ValidateDenom(req.XDenom); err != nil {
			return nil, err
		}
	}

	if req.GetYDenom() != "" {
		if err := sdk.ValidateDenom(req.YDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	pairStore := prefix.NewStore(store, types.PairKeyPrefix)

	pairs := []types.Pair{}
	pageRes, err := query.FilteredPaginate(pairStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pair, err := types.UnmarshalPair(k.cdc, value)
		if err != nil {
			return false, err
		}

		if req.GetXDenom() != "" {
			if pair.XCoinDenom != req.GetXDenom() {
				return false, nil
			}
		}

		if req.GetYDenom() != "" {
			if pair.YCoinDenom != req.GetYDenom() {
				return false, nil
			}
		}

		if accumulate {
			pairs = append(pairs, pair)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPairsResponse{Pairs: pairs, Pagination: pageRes}, nil
}

// Pair queries the specific pair.
func (k Querier) Pair(c context.Context, req *types.QueryPairRequest) (*types.QueryPairResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	pair, found := k.GetPair(ctx, req.PairId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pair %d doesn't exist", req.PairId)
	}

	return &types.QueryPairResponse{Pair: pair}, nil
}

// DepositRequests queries all deposit requests.
func (k Querier) DepositRequests(c context.Context, req *types.QueryDepositRequestsRequest) (*types.QueryDepositRequestsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryDepositRequestsResponse{}, nil
}

// DepositRequest quereis the specific deposit request.
func (k Querier) DepositRequest(c context.Context, req *types.QueryDepositRequestRequest) (*types.QueryDepositRequestResponse, error) {
	// TODO: not implemented yet
	return &types.QueryDepositRequestResponse{}, nil
}

// WithdrawRequests quereis all withdraw requests.
func (k Querier) WithdrawRequests(c context.Context, req *types.QueryWithdrawRequestsRequest) (*types.QueryWithdrawRequestsResponse, error) {
	// TODO: not implemented yet
	return &types.QueryWithdrawRequestsResponse{}, nil
}

// WithdrawRequest quereis the specific withdraw request.
func (k Querier) WithdrawRequest(c context.Context, req *types.QueryWithdrawRequestRequest) (*types.QueryWithdrawRequestResponse, error) {
	// TODO: not implemented yet
	return &types.QueryWithdrawRequestResponse{}, nil
}

// SwapRequests queries all swap requests.
func (k Querier) SwapRequests(c context.Context, req *types.QuerySwapRequestsRequest) (*types.QuerySwapRequestsResponse, error) {
	// TODO: not implemented yet
	return &types.QuerySwapRequestsResponse{}, nil
}

// SwapRequest queries the specific swap request.
func (k Querier) SwapRequest(c context.Context, req *types.QuerySwapRequestRequest) (*types.QuerySwapRequestResponse, error) {
	// TODO: not implemented yet
	return &types.QuerySwapRequestResponse{}, nil
}
