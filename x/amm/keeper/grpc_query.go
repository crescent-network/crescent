package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// Querier is used as Keeper will have duplicate methods if used directly,
// and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.Keeper.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Querier) AllPools(c context.Context, req *types.QueryAllPoolsRequest) (*types.QueryAllPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolKeyPrefix)
	var poolResps []types.PoolResponse
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(_, value []byte) error {
		var pool types.Pool
		k.cdc.MustUnmarshal(value, &pool)
		poolResps = append(poolResps, k.MakePoolResponse(ctx, pool))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllPoolsResponse{
		Pools:      poolResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Pool(c context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	pool, found := k.GetPool(ctx, req.PoolId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
	}
	return &types.QueryPoolResponse{Pool: k.MakePoolResponse(ctx, pool)}, nil
}

func (k Querier) AllPositions(c context.Context, req *types.QueryAllPositionsRequest) (*types.QueryAllPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	positionStore := prefix.NewStore(store, types.PositionKeyPrefix)
	var positionResps []types.PositionResponse
	pageRes, err := query.Paginate(positionStore, req.Pagination, func(_, value []byte) error {
		var position types.Position
		k.cdc.MustUnmarshal(value, &position)
		positionResps = append(positionResps, types.NewPositionResponse(position))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllPositionsResponse{
		Positions:  positionResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Positions(c context.Context, req *types.QueryPositionsRequest) (*types.QueryPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ownerAddr, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner address: %v", err)
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	positionStore := prefix.NewStore(store, types.GetPositionsByOwnerIteratorPrefix(ownerAddr))
	var positionResps []types.PositionResponse
	pageRes, err := query.Paginate(positionStore, req.Pagination, func(_, value []byte) error {
		var position types.Position
		k.cdc.MustUnmarshal(value, &position)
		positionResps = append(positionResps, types.NewPositionResponse(position))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPositionsResponse{
		Positions:  positionResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) MakePoolResponse(ctx sdk.Context, pool types.Pool) types.PoolResponse {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	balances := k.bankKeeper.SpendableCoins(ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress))
	return types.NewPoolResponse(pool, poolState, balances)
}
