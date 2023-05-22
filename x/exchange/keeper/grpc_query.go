package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
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
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Querier) AllMarkets(c context.Context, req *types.QueryAllMarketsRequest) (*types.QueryAllMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	marketStore := prefix.NewStore(store, types.MarketKeyPrefix)
	var marketResps []types.MarketResponse
	pageRes, err := query.Paginate(marketStore, req.Pagination, func(key, value []byte) error {
		var market types.Market
		k.cdc.MustUnmarshal(value, &market)
		marketResps = append(marketResps, k.MakeMarketResponse(ctx, market))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllMarketsResponse{
		Markets:    marketResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Market(c context.Context, req *types.QueryMarketRequest) (*types.QueryMarketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	market, found := k.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "market not found")
	}
	return &types.QueryMarketResponse{Market: k.MakeMarketResponse(ctx, market)}, nil
}

func (k Querier) Order(c context.Context, req *types.QueryOrderRequest) (*types.QueryOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	order, found := k.GetOrder(ctx, req.OrderId)
	if !found {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return &types.QueryOrderResponse{Order: order}, nil
}

func (k Querier) BestSwapExactAmountInRoutes(c context.Context, req *types.QueryBestSwapExactAmountInRoutesRequest) (*types.QueryBestSwapExactAmountInRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	allRoutes := k.FindAllRoutes(ctx, req.Input.Denom, req.OutputDenom, 3) // TODO: remove hard-coded limit
	if len(allRoutes) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "no possible routes")
	}
	var (
		bestOutput = utils.ZeroInt
		bestRoutes []uint64
	)
	for _, routes := range allRoutes {
		output, err := k.SwapExactAmountIn(
			ctx, sdk.AccAddress{}, routes, req.Input, sdk.NewCoin(req.OutputDenom, utils.ZeroInt), true)
		if err != nil && !errors.Is(err, types.ErrInsufficientOutput) { // sanity check
			panic(err)
		}
		if err == nil {
			if output.Amount.GT(bestOutput) {
				bestOutput = output.Amount
				bestRoutes = routes
			}
		}
	}
	return &types.QueryBestSwapExactAmountInRoutesResponse{
		Routes: bestRoutes,
		Output: sdk.NewCoin(req.OutputDenom, bestOutput),
	}, nil
}

func (k Querier) MakeMarketResponse(ctx sdk.Context, market types.Market) types.MarketResponse {
	marketState := k.MustGetMarketState(ctx, market.Id)
	return types.NewMarketResponse(market, marketState)
}
