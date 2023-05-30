package keeper

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
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
		position, found := k.GetPosition(ctx, sdk.BigEndianToUint64(value))
		if !found { // sanity check
			panic("position not found")
		}
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

func (k Querier) PoolPositions(c context.Context, req *types.QueryPoolPositionsRequest) (*types.QueryPoolPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool id must not be 0")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPool(ctx, req.PoolId); !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.GetPositionsByPoolIteratorPrefix(req.PoolId)
	positionStore := prefix.NewStore(store, keyPrefix)
	var positionResps []types.PositionResponse
	pageRes, err := query.Paginate(positionStore, req.Pagination, func(key, _ []byte) error {
		_, positionId := types.ParsePositionsByPoolIndexKey(utils.Key(keyPrefix, key))
		position, found := k.GetPosition(ctx, positionId)
		if !found { // sanity check
			panic("position not found")
		}
		positionResps = append(positionResps, types.NewPositionResponse(position))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolPositionsResponse{
		Positions:  positionResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) AllFarmingPlans(c context.Context, req *types.QueryAllFarmingPlansRequest) (*types.QueryAllFarmingPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	var isPrivate, isTerminated bool
	if req.IsPrivate != "" {
		var err error
		isPrivate, err = strconv.ParseBool(req.IsPrivate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid is_private: %v", err)
		}
	}
	if req.IsTerminated != "" {
		var err error
		isTerminated, err = strconv.ParseBool(req.IsTerminated)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid is_terminated: %v", err)
		}
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	planStore := prefix.NewStore(store, types.FarmingPlanKeyPrefix)
	var plans []types.FarmingPlan
	pageRes, err := query.FilteredPaginate(planStore, req.Pagination, func(_, value []byte, accumulate bool) (hit bool, err error) {
		var plan types.FarmingPlan
		k.cdc.MustUnmarshal(value, &plan)
		if req.IsPrivate != "" && plan.IsPrivate != isPrivate {
			return false, nil
		}
		if req.IsTerminated != "" && plan.IsTerminated != isTerminated {
			return false, nil
		}
		if accumulate {
			plans = append(plans, plan)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllFarmingPlansResponse{
		FarmingPlans: plans,
		Pagination:   pageRes,
	}, nil
}

func (k Querier) MakePoolResponse(ctx sdk.Context, pool types.Pool) types.PoolResponse {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	balances := k.bankKeeper.SpendableCoins(ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress))
	return types.NewPoolResponse(pool, poolState, balances)
}
