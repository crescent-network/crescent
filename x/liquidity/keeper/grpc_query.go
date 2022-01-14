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

	if req.XDenom != "" {
		if err := sdk.ValidateDenom(req.XDenom); err != nil {
			return nil, err
		}
	}

	if req.YDenom != "" {
		if err := sdk.ValidateDenom(req.YDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolKeyPrefix)

	var poolsRes []types.PoolResponse
	pageRes, err := query.FilteredPaginate(poolStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pool, err := types.UnmarshalPool(k.cdc, value)
		if err != nil {
			return false, err
		}

		if req.XDenom != "" {
			if pool.XCoinDenom != req.XDenom {
				return false, nil
			}
		}

		if req.YDenom != "" {
			if pool.YCoinDenom != req.YDenom {
				return false, nil
			}
		}

		rx, ry := k.GetPoolBalance(ctx, pool)
		poolRes := types.PoolResponse{
			Id:                    pool.Id,
			PairId:                pool.PairId,
			XCoinDenom:            pool.XCoinDenom,
			YCoinDenom:            pool.YCoinDenom,
			ReserveAddress:        pool.ReserveAddress,
			PoolCoinDenom:         pool.PoolCoinDenom,
			XCoin:                 sdk.NewCoin(pool.XCoinDenom, rx),
			YCoin:                 sdk.NewCoin(pool.XCoinDenom, ry),
			LastDepositRequestId:  pool.LastDepositRequestId,
			LastWithdrawRequestId: pool.LastWithdrawRequestId,
		}

		if accumulate {
			poolsRes = append(poolsRes, poolRes)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{Pools: poolsRes, Pagination: pageRes}, nil
}

// PoolsByPair queries all pools that correspond to the pair.
func (k Querier) PoolsByPair(c context.Context, req *types.QueryPoolsByPairRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PairId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.GetPoolsByPairIndexKeyPrefix(req.PairId)
	poolsStore := prefix.NewStore(store, keyPrefix)

	var poolsRes []types.PoolResponse
	pageRes, err := query.Paginate(poolsStore, req.Pagination, func(key, value []byte) error {
		poolId := types.ParsePoolsByPairIndexKey(append(keyPrefix, key...))

		pool, found := k.GetPool(ctx, poolId)
		if !found {
			return nil
		}

		rx, ry := k.GetPoolBalance(ctx, pool)
		poolRes := types.PoolResponse{
			Id:                    pool.Id,
			PairId:                pool.PairId,
			XCoinDenom:            pool.XCoinDenom,
			YCoinDenom:            pool.YCoinDenom,
			ReserveAddress:        pool.ReserveAddress,
			PoolCoinDenom:         pool.PoolCoinDenom,
			XCoin:                 sdk.NewCoin(pool.XCoinDenom, rx),
			YCoin:                 sdk.NewCoin(pool.XCoinDenom, ry),
			LastDepositRequestId:  pool.LastDepositRequestId,
			LastWithdrawRequestId: pool.LastWithdrawRequestId,
		}

		poolsRes = append(poolsRes, poolRes)

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPoolsResponse{Pools: poolsRes, Pagination: pageRes}, nil
}

// Pool queries the specific pool.
func (k Querier) Pool(c context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	pool, found := k.GetPool(ctx, req.PoolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool %d doesn't exist", req.PoolId)
	}

	rx, ry := k.GetPoolBalance(ctx, pool)
	poolRes := types.PoolResponse{
		Id:                    pool.Id,
		PairId:                pool.PairId,
		XCoinDenom:            pool.XCoinDenom,
		YCoinDenom:            pool.YCoinDenom,
		ReserveAddress:        pool.ReserveAddress,
		PoolCoinDenom:         pool.PoolCoinDenom,
		XCoin:                 sdk.NewCoin(pool.XCoinDenom, rx),
		YCoin:                 sdk.NewCoin(pool.XCoinDenom, ry),
		LastDepositRequestId:  pool.LastDepositRequestId,
		LastWithdrawRequestId: pool.LastWithdrawRequestId,
	}

	return &types.QueryPoolResponse{Pool: poolRes}, nil
}

// PoolByReserveAcc queries the specific pool by the reserve account address.
func (k Querier) PoolByReserveAcc(c context.Context, req *types.QueryPoolByReserveAccRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ReserveAcc == "" {
		return nil, status.Error(codes.InvalidArgument, "empty reserve account address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	reserveAcc, err := sdk.AccAddressFromBech32(req.ReserveAcc)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "reserve account address %s is not valid", req.ReserveAcc)
	}

	pool, found := k.GetPoolByReserveAcc(ctx, reserveAcc)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool by %s doesn't exist", req.ReserveAcc)
	}

	rx, ry := k.GetPoolBalance(ctx, pool)
	poolRes := types.PoolResponse{
		Id:                    pool.Id,
		PairId:                pool.PairId,
		XCoinDenom:            pool.XCoinDenom,
		YCoinDenom:            pool.YCoinDenom,
		ReserveAddress:        pool.ReserveAddress,
		PoolCoinDenom:         pool.PoolCoinDenom,
		XCoin:                 sdk.NewCoin(pool.XCoinDenom, rx),
		YCoin:                 sdk.NewCoin(pool.XCoinDenom, ry),
		LastDepositRequestId:  pool.LastDepositRequestId,
		LastWithdrawRequestId: pool.LastWithdrawRequestId,
	}

	return &types.QueryPoolResponse{Pool: poolRes}, nil
}

// PoolByPoolCoinDenom queries the specific pool by the pool coin denomination.
func (k Querier) PoolByPoolCoinDenom(c context.Context, req *types.QueryPoolByPoolCoinDenomRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolCoinDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "empty pool coin denom")
	}

	ctx := sdk.UnwrapSDKContext(c)

	poolId := types.ParsePoolCoinDenom(req.PoolCoinDenom)
	pool, found := k.GetPool(ctx, poolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool %d doesn't exist", poolId)
	}

	rx, ry := k.GetPoolBalance(ctx, pool)
	poolRes := types.PoolResponse{
		Id:                    pool.Id,
		PairId:                pool.PairId,
		XCoinDenom:            pool.XCoinDenom,
		YCoinDenom:            pool.YCoinDenom,
		ReserveAddress:        pool.ReserveAddress,
		PoolCoinDenom:         pool.PoolCoinDenom,
		XCoin:                 sdk.NewCoin(pool.XCoinDenom, rx),
		YCoin:                 sdk.NewCoin(pool.XCoinDenom, ry),
		LastDepositRequestId:  pool.LastDepositRequestId,
		LastWithdrawRequestId: pool.LastWithdrawRequestId,
	}

	return &types.QueryPoolResponse{Pool: poolRes}, nil
}

// Pairs queries all pairs.
func (k Querier) Pairs(c context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.XDenom != "" {
		if err := sdk.ValidateDenom(req.XDenom); err != nil {
			return nil, err
		}
	}

	if req.YDenom != "" {
		if err := sdk.ValidateDenom(req.YDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	pairStore := prefix.NewStore(store, types.PairKeyPrefix)

	var pairs []types.Pair
	pageRes, err := query.FilteredPaginate(pairStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pair, err := types.UnmarshalPair(k.cdc, value)
		if err != nil {
			return false, err
		}

		if req.XDenom != "" {
			if pair.XCoinDenom != req.XDenom {
				return false, nil
			}
		}

		if req.YDenom != "" {
			if pair.YCoinDenom != req.YDenom {
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

	if req.PairId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair id cannot be 0")
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	drsStore := prefix.NewStore(store, types.DepositRequestKeyPrefix)

	var drs []types.DepositRequest
	pageRes, err := query.FilteredPaginate(drsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		dr, err := types.UnmarshalDepositRequest(k.cdc, value)
		if err != nil {
			return false, err
		}

		if dr.PoolId != req.PoolId {
			return false, nil
		}

		if accumulate {
			drs = append(drs, dr)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositRequestsResponse{DepositRequests: drs, Pagination: pageRes}, nil
}

// DepositRequest quereis the specific deposit request.
func (k Querier) DepositRequest(c context.Context, req *types.QueryDepositRequestRequest) (*types.QueryDepositRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be 0")
	}

	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	dq, found := k.GetDepositRequest(ctx, req.PoolId, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "deposit request of pool id %d and request id %d doesn't exist or deleted", req.PoolId, req.Id)
	}

	return &types.QueryDepositRequestResponse{DepositRequest: dq}, nil
}

// WithdrawRequests quereis all withdraw requests.
func (k Querier) WithdrawRequests(c context.Context, req *types.QueryWithdrawRequestsRequest) (*types.QueryWithdrawRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	drsStore := prefix.NewStore(store, types.WithdrawRequestKeyPrefix)

	var wrs []types.WithdrawRequest
	pageRes, err := query.FilteredPaginate(drsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		wr, err := types.UnmarshalWithdrawRequest(k.cdc, value)
		if err != nil {
			return false, err
		}

		if wr.PoolId != req.PoolId {
			return false, nil
		}

		if accumulate {
			wrs = append(wrs, wr)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawRequestsResponse{WithdrawRequests: wrs, Pagination: pageRes}, nil
}

// WithdrawRequest quereis the specific withdraw request.
func (k Querier) WithdrawRequest(c context.Context, req *types.QueryWithdrawRequestRequest) (*types.QueryWithdrawRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id cannot be 0")
	}

	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	wq, found := k.GetWithdrawRequest(ctx, req.PoolId, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "withdraw request of pool id %d and request id %d doesn't exist or deleted", req.PoolId, req.Id)
	}

	return &types.QueryWithdrawRequestResponse{WithdrawRequest: wq}, nil
}

// SwapRequests queries all swap requests.
func (k Querier) SwapRequests(c context.Context, req *types.QuerySwapRequestsRequest) (*types.QuerySwapRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PairId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	drsStore := prefix.NewStore(store, types.SwapRequestKeyPrefix)

	var srs []types.SwapRequest
	pageRes, err := query.FilteredPaginate(drsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		sr, err := types.UnmarshalSwapRequest(k.cdc, value)
		if err != nil {
			return false, err
		}

		if sr.PairId != req.PairId {
			return false, nil
		}

		if accumulate {
			srs = append(srs, sr)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapRequestsResponse{SwapRequests: srs, Pagination: pageRes}, nil
}

// SwapRequest queries the specific swap request.
func (k Querier) SwapRequest(c context.Context, req *types.QuerySwapRequestRequest) (*types.QuerySwapRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PairId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair id cannot be 0")
	}

	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	sq, found := k.GetSwapRequest(ctx, req.PairId, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "swap request of pair id %d and request id %d doesn't exist or deleted", req.PairId, req.Id)
	}

	return &types.QuerySwapRequestResponse{SwapRequest: sq}, nil
}
