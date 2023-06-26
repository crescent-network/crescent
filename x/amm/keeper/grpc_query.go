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
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	minttypes "github.com/crescent-network/crescent/v5/x/mint/types"
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
	var (
		keyPrefix  []byte
		poolGetter func(key, value []byte) types.Pool
	)
	if req.MarketId > 0 {
		if found := k.exchangeKeeper.LookupMarket(ctx, req.MarketId); !found {
			return nil, status.Error(codes.NotFound, "market not found")
		}
		keyPrefix = types.GetPoolByMarketIndexKey(req.MarketId)
		poolGetter = func(_, value []byte) types.Pool {
			pool, found := k.GetPool(ctx, sdk.BigEndianToUint64(value))
			if !found { // sanity check
				panic("pool not found")
			}
			return pool
		}
	} else {
		keyPrefix = types.PoolKeyPrefix
		poolGetter = func(_, value []byte) types.Pool {
			var pool types.Pool
			k.cdc.MustUnmarshal(value, &pool)
			return pool
		}
	}
	poolStore := prefix.NewStore(store, keyPrefix)
	var poolResps []types.PoolResponse
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key, value []byte) error {
		pool := poolGetter(key, value)
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
	var ownerAddr sdk.AccAddress
	if req.Owner != "" {
		var err error
		ownerAddr, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid owner: %v", err)
		}
	}
	if req.PoolId > 0 {
		if found := k.LookupPool(ctx, req.PoolId); !found {
			return nil, status.Error(codes.NotFound, "pool not found")
		}
	}
	store := ctx.KVStore(k.storeKey)
	// TODO: filter by pool id and owner
	var (
		keyPrefix      []byte
		positionGetter func(key, value []byte) types.Position
	)
	getPositionFromPositionByParamsIndexKey := func(_, value []byte) types.Position {
		position, found := k.GetPosition(ctx, sdk.BigEndianToUint64(value))
		if !found { // sanity check
			panic("position not found")
		}
		return position
	}
	if req.PoolId > 0 && req.Owner != "" {
		keyPrefix = types.GetPositionsByPoolAndOwnerIteratorPrefix(ownerAddr, req.PoolId)
		positionGetter = getPositionFromPositionByParamsIndexKey
	} else if req.PoolId > 0 {
		keyPrefix = types.GetPositionsByPoolIteratorPrefix(req.PoolId)
		positionGetter = func(key, _ []byte) types.Position {
			_, positionId := types.ParsePositionsByPoolIndexKey(utils.Key(keyPrefix, key))
			position, found := k.GetPosition(ctx, positionId)
			if !found { // sanity check
				panic("position not found")
			}
			return position
		}
	} else if req.Owner != "" {
		keyPrefix = types.GetPositionsByOwnerIteratorPrefix(ownerAddr)
		positionGetter = getPositionFromPositionByParamsIndexKey
	} else {
		keyPrefix = types.PositionKeyPrefix
		positionGetter = func(_, value []byte) types.Position {
			var position types.Position
			k.cdc.MustUnmarshal(value, &position)
			return position
		}
	}
	positionStore := prefix.NewStore(store, keyPrefix)
	var positionResps []types.PositionResponse
	pageRes, err := query.Paginate(positionStore, req.Pagination, func(key, value []byte) error {
		position := positionGetter(key, value)
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

func (k Querier) Position(c context.Context, req *types.QueryPositionRequest) (*types.QueryPositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	position, found := k.GetPosition(ctx, req.PositionId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	return &types.QueryPositionResponse{Position: types.NewPositionResponse(position)}, nil
}

func (k Querier) AddLiquiditySimulation(c context.Context, req *types.QueryAddLiquiditySimulationRequest) (*types.QueryAddLiquiditySimulationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	desiredAmt, err := sdk.ParseCoinsNormalized(req.DesiredAmount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid desired amount: %v", err)
	}
	lowerPrice, err := sdk.NewDecFromStr(req.LowerPrice)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lower price: %v", err)
	}
	upperPrice, err := sdk.NewDecFromStr(req.UpperPrice)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid upper price: %v", err)
	}
	ctx := sdk.UnwrapSDKContext(c)
	// Create temporary account with sufficient funds.
	ownerAddr := sdk.AccAddress(crypto.AddressHash([]byte("simaccount")))
	if err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, desiredAmt); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, ownerAddr, desiredAmt); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := types.NewMsgAddLiquidity(
		ownerAddr, req.PoolId, lowerPrice, upperPrice, desiredAmt).ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, liquidity, amt, err := k.AddLiquidity(
		ctx, ownerAddr, ownerAddr, req.PoolId, lowerPrice, upperPrice, desiredAmt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &types.QueryAddLiquiditySimulationResponse{
		Liquidity: liquidity,
		Amount:    amt,
	}, nil
}

func (k Querier) RemoveLiquiditySimulation(c context.Context, req *types.QueryRemoveLiquiditySimulationRequest) (*types.QueryRemoveLiquiditySimulationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	liquidity, ok := sdk.NewIntFromString(req.Liquidity)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid liquidity: %s", req.Liquidity)
	}
	ctx := sdk.UnwrapSDKContext(c)
	position, found := k.GetPosition(ctx, req.PositionId)
	if !found {
		return nil, status.Error(codes.NotFound, "position not found")
	}
	ownerAddr := sdk.MustAccAddressFromBech32(position.Owner)
	if err := types.NewMsgRemoveLiquidity(
		ownerAddr, req.PositionId, liquidity).ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, amt, err := k.RemoveLiquidity(
		ctx, ownerAddr, ownerAddr, req.PositionId, liquidity)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return &types.QueryRemoveLiquiditySimulationResponse{
		Amount: amt,
	}, nil
}

func (k Querier) CollectibleCoins(c context.Context, req *types.QueryCollectibleCoinsRequest) (*types.QueryCollectibleCoinsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Owner != "" && req.PositionId > 0 {
		return nil, status.Error(codes.InvalidArgument, "owner and position id must not be specified at the same time")
	} else if req.Owner == "" && req.PositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "owner or position id must be specified")
	}
	var ownerAddr sdk.AccAddress
	if req.Owner != "" {
		var err error
		ownerAddr, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid owner: %v", err)
		}
	}
	ctx := sdk.UnwrapSDKContext(c)
	if req.PositionId > 0 {
		fee, farmingRewards, err := k.Keeper.CollectibleCoins(ctx, req.PositionId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return &types.QueryCollectibleCoinsResponse{
			Fee:            fee,
			FarmingRewards: farmingRewards,
		}, nil
	} else {
		var totalFee, totalFarmingRewards sdk.Coins
		k.IteratePositionsByOwner(ctx, ownerAddr, func(position types.Position) (stop bool) {
			fee, farmingRewards, err := k.Keeper.CollectibleCoins(ctx, position.Id)
			if err != nil { // sanity check
				panic(err)
			}
			totalFee = totalFee.Add(fee...)
			totalFarmingRewards = totalFarmingRewards.Add(farmingRewards...)
			return false
		})
		return &types.QueryCollectibleCoinsResponse{
			Fee:            totalFee,
			FarmingRewards: totalFarmingRewards,
		}, nil
	}
}

func (k Querier) AllTickInfos(c context.Context, req *types.QueryAllTickInfosRequest) (*types.QueryAllTickInfosResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPool(ctx, req.PoolId); !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	var lowerTick, upperTick int64
	if req.LowerTick != "" {
		var err error
		lowerTick, err = strconv.ParseInt(req.LowerTick, 10, 32)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid lower tick: %v", err)
		}
	}
	if req.UpperTick != "" {
		var err error
		upperTick, err = strconv.ParseInt(req.UpperTick, 10, 32)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid upper tick: %v", err)
		}
	}
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.GetTickInfosByPoolIteratorPrefix(req.PoolId)
	tickInfoStore := prefix.NewStore(store, keyPrefix)
	var tickInfoResps []types.TickInfoResponse
	pageRes, err := query.FilteredPaginate(tickInfoStore, req.Pagination, func(key, value []byte, accumulate bool) (hit bool, err error) {
		var tickInfo types.TickInfo
		k.cdc.MustUnmarshal(value, &tickInfo)
		_, tick := types.ParseTickInfoKey(utils.Key(keyPrefix, key))
		if req.LowerTick != "" && tick < int32(lowerTick) {
			return false, nil
		}
		if req.UpperTick != "" && tick > int32(upperTick) {
			return false, nil
		}
		if accumulate {
			tickInfoResps = append(
				tickInfoResps, types.NewTickInfoResponse(tick, tickInfo))
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllTickInfosResponse{
		TickInfos:  tickInfoResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) TickInfo(c context.Context, req *types.QueryTickInfoRequest) (*types.QueryTickInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPool(ctx, req.PoolId); !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	tickInfo, found := k.GetTickInfo(ctx, req.PoolId, req.Tick)
	if !found {
		return nil, status.Error(codes.NotFound, "tick info not found")
	}
	return &types.QueryTickInfoResponse{
		TickInfo: types.NewTickInfoResponse(req.Tick, tickInfo),
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

func (k Querier) FarmingPlan(c context.Context, req *types.QueryFarmingPlanRequest) (*types.QueryFarmingPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	plan, found := k.GetFarmingPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "farming plan not found")
	}
	return &types.QueryFarmingPlanResponse{FarmingPlan: plan}, nil
}

func (k Querier) MakePoolResponse(ctx sdk.Context, pool types.Pool) types.PoolResponse {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	balances := k.bankKeeper.SpendableCoins(ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress))
	return types.NewPoolResponse(pool, poolState, balances)
}
