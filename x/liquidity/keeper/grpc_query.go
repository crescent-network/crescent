package keeper

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
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

// Pairs queries all pairs.
func (k Querier) Pairs(c context.Context, req *types.QueryPairsRequest) (*types.QueryPairsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.Denoms) > 2 {
		return nil, status.Errorf(codes.InvalidArgument, "too many denoms to query: %d", len(req.Denoms))
	}

	for _, denom := range req.Denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	var keyPrefix []byte
	var pairGetter func(key, value []byte) types.Pair
	switch len(req.Denoms) {
	case 0:
		keyPrefix = types.PairKeyPrefix
		pairGetter = func(_, value []byte) types.Pair {
			return types.MustUnmarshalPair(k.cdc, value)
		}
	case 1:
		keyPrefix = types.GetPairsByDenomIndexKeyPrefix(req.Denoms[0])
		pairGetter = func(key, _ []byte) types.Pair {
			_, _, pairId := types.ParsePairsByDenomsIndexKey(append(keyPrefix, key...))
			pair, _ := k.GetPair(ctx, pairId)
			return pair
		}
	case 2:
		keyPrefix = types.GetPairsByDenomsIndexKeyPrefix(req.Denoms[0], req.Denoms[1])
		pairGetter = func(key, _ []byte) types.Pair {
			_, _, pairId := types.ParsePairsByDenomsIndexKey(append(keyPrefix, key...))
			pair, _ := k.GetPair(ctx, pairId)
			return pair
		}
	}
	pairStore := prefix.NewStore(store, keyPrefix)

	var pairs []types.Pair
	pageRes, err := query.FilteredPaginate(pairStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pair := pairGetter(key, value)

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

// Pools queries all pools.
func (k Querier) Pools(c context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var disabled bool
	if req.Disabled != "" {
		var err error
		disabled, err = strconv.ParseBool(req.Disabled)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	var keyPrefix []byte
	var poolGetter func(key, value []byte) types.Pool
	var pairGetter func(id uint64) types.Pair
	pairMap := map[uint64]types.Pair{}
	switch {
	case req.PairId == 0:
		keyPrefix = types.PoolKeyPrefix
		poolGetter = func(_, value []byte) types.Pool {
			return types.MustUnmarshalPool(k.cdc, value)
		}
		pairGetter = func(id uint64) types.Pair {
			pair, ok := pairMap[id]
			if !ok {
				pair, _ = k.GetPair(ctx, id)
				pairMap[id] = pair
			}
			return pair
		}
	default:
		keyPrefix = types.GetPoolsByPairIndexKeyPrefix(req.PairId)
		poolGetter = func(key, _ []byte) types.Pool {
			poolId := types.ParsePoolsByPairIndexKey(append(keyPrefix, key...))
			pool, _ := k.GetPool(ctx, poolId)
			return pool
		}
		pair, _ := k.GetPair(ctx, req.PairId)
		pairGetter = func(_ uint64) types.Pair {
			return pair
		}
	}

	poolStore := prefix.NewStore(store, keyPrefix)

	var poolsRes []types.PoolResponse
	pageRes, err := query.FilteredPaginate(poolStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		pool := poolGetter(key, value)
		if req.Disabled != "" {
			if pool.Disabled != disabled {
				return false, nil
			}
		}

		if accumulate {
			pair := pairGetter(pool.PairId)
			rx, ry := k.getPoolBalances(ctx, pool, pair)
			ps := k.GetPoolCoinSupply(ctx, pool)
			poolsRes = append(poolsRes, types.NewPoolResponse(pool, rx, ry, ps))
		}

		return true, nil
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

	pair, _ := k.GetPair(ctx, pool.PairId)

	rx, ry := k.getPoolBalances(ctx, pool, pair)
	ps := k.GetPoolCoinSupply(ctx, pool)
	return &types.QueryPoolResponse{Pool: types.NewPoolResponse(pool, rx, ry, ps)}, nil
}

// PoolByReserveAddress queries the specific pool by the reserve account address.
func (k Querier) PoolByReserveAddress(c context.Context, req *types.QueryPoolByReserveAddressRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ReserveAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty reserve account address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	reserveAddr, err := sdk.AccAddressFromBech32(req.ReserveAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "reserve account address %s is not valid", req.ReserveAddress)
	}

	pool, found := k.GetPoolByReserveAddress(ctx, reserveAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool by %s doesn't exist", req.ReserveAddress)
	}

	rx, ry := k.GetPoolBalances(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)
	return &types.QueryPoolResponse{Pool: types.NewPoolResponse(pool, rx, ry, ps)}, nil
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

	poolId, err := types.ParsePoolCoinDenom(req.PoolCoinDenom)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse pool coin denom: %v", err)
	}
	pool, found := k.GetPool(ctx, poolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "pool %d doesn't exist", poolId)
	}

	rx, ry := k.GetPoolBalances(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)
	return &types.QueryPoolResponse{Pool: types.NewPoolResponse(pool, rx, ry, ps)}, nil
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

// DepositRequest queries the specific deposit request.
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

// WithdrawRequests queries all withdraw requests.
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

// WithdrawRequest queries the specific withdraw request.
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

// Orders queries all orders.
func (k Querier) Orders(c context.Context, req *types.QueryOrdersRequest) (*types.QueryOrdersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PairId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	drsStore := prefix.NewStore(store, types.OrderKeyPrefix)

	var orders []types.Order
	pageRes, err := query.FilteredPaginate(drsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		order, err := types.UnmarshalOrder(k.cdc, value)
		if err != nil {
			return false, err
		}

		if order.PairId != req.PairId {
			return false, nil
		}

		if accumulate {
			orders = append(orders, order)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOrdersResponse{Orders: orders, Pagination: pageRes}, nil
}

// Order queries the specific order.
func (k Querier) Order(c context.Context, req *types.QueryOrderRequest) (*types.QueryOrderResponse, error) {
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

	order, found := k.GetOrder(ctx, req.PairId, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "order %d in pair %d not found", req.PairId, req.Id)
	}

	return &types.QueryOrderResponse{Order: order}, nil
}

// OrdersByOrderer returns orders made by an orderer.
func (k Querier) OrdersByOrderer(c context.Context, req *types.QueryOrdersByOrdererRequest) (*types.QueryOrdersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	orderer, err := sdk.AccAddressFromBech32(req.Orderer)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "orderer address %s is invalid", req.Orderer)
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	keyPrefix := types.GetOrderIndexKeyPrefix(orderer)
	orderStore := prefix.NewStore(store, keyPrefix)
	var orders []types.Order
	pageRes, err := query.FilteredPaginate(orderStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, pairId, orderId := types.ParseOrderIndexKey(append(keyPrefix, key...))
		if req.PairId != 0 && pairId != req.PairId {
			return false, nil
		}

		order, _ := k.GetOrder(ctx, pairId, orderId)

		if accumulate {
			orders = append(orders, order)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOrdersResponse{Orders: orders, Pagination: pageRes}, nil
}

// OrderBooks queries virtual order books from user orders and pools.
func (k Querier) OrderBooks(c context.Context, req *types.QueryOrderBooksRequest) (*types.QueryOrderBooksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.PairIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "pair ids must not be empty")
	}

	if len(req.PriceUnitPowers) == 0 {
		req.PriceUnitPowers = []uint32{0, 1, 2}
	}

	if req.NumTicks == 0 {
		return nil, status.Error(codes.InvalidArgument, "number of ticks must not be 0")
	}

	pairIdSet := map[uint64]struct{}{}
	for _, pairId := range req.PairIds {
		if _, ok := pairIdSet[pairId]; ok {
			return nil, status.Errorf(codes.InvalidArgument, "duplicate pair id: %d", pairId)
		}
		pairIdSet[pairId] = struct{}{}
	}

	priceUnitPowerSet := map[uint32]struct{}{}
	for _, p := range req.PriceUnitPowers {
		if _, ok := priceUnitPowerSet[p]; ok {
			return nil, status.Errorf(codes.InvalidArgument, "duplicate price unit power: %d", p)
		}
		priceUnitPowerSet[p] = struct{}{}
	}

	ctx := sdk.UnwrapSDKContext(c)

	tickPrec := k.GetTickPrecision(ctx)

	var pairs []types.OrderBookPairResponse
	for _, pairId := range req.PairIds {
		pair, found := k.GetPair(ctx, pairId)
		if !found {
			return nil, status.Errorf(codes.NotFound, "pair %d doesn't exist", pairId)
		}

		if pair.LastPrice == nil {
			return nil, status.Errorf(codes.Unavailable, "pair %d does not have last price", pairId)
		}

		ob := amm.NewOrderBook()
		_ = k.IterateOrdersByPair(ctx, pairId, func(order types.Order) (stop bool, err error) {
			switch order.Status {
			case types.OrderStatusNotExecuted,
				types.OrderStatusNotMatched,
				types.OrderStatusPartiallyMatched:
				ob.AddOrder(types.NewUserOrder(order))
			}
			return false, nil
		})

		lowestPrice, highestPrice := k.PriceLimits(ctx, *pair.LastPrice)
		_ = k.IteratePoolsByPair(ctx, pairId, func(pool types.Pool) (stop bool, err error) {
			if pool.Disabled {
				return false, nil
			}
			rx, ry := k.getPoolBalances(ctx, pool, pair)
			ammPool := pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{})
			ob.AddOrder(amm.PoolOrders(ammPool, amm.DefaultOrderer, lowestPrice, highestPrice, int(tickPrec))...)
			return false, nil
		})

		ov := ob.MakeView()
		ov.Match()

		var configs []types.OrderBookConfig
		for _, p := range req.PriceUnitPowers {
			configs = append(configs, types.OrderBookConfig{
				PriceUnitPower: int(p),
				MaxNumTicks:    int(req.NumTicks),
			})
		}

		pairs = append(
			pairs, types.MakeOrderBookPairResponse(
				pair.Id, ov, lowestPrice, highestPrice, int(tickPrec), configs...))
	}

	return &types.QueryOrderBooksResponse{
		Pairs: pairs,
	}, nil
}
