package keeper

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/marketmaker/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the marketmaker module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.Keeper.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// MarketMakers queries all market makers.
func (k Querier) MarketMakers(c context.Context, req *types.QueryMarketMakersRequest) (*types.QueryMarketMakersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var mmAddr sdk.AccAddress
	var eligible bool
	var err error
	if req.Address != "" {
		mmAddr, err = sdk.AccAddressFromBech32(req.Address)
		if err != nil {
			return nil, err
		}
	}

	if req.Eligible != "" {
		eligible, err = strconv.ParseBool(req.Eligible)
		if err != nil {
			return nil, err
		}
	}

	// query specific market maker case
	if !mmAddr.Empty() && req.PairId != 0 {
		mm, found := k.GetMarketMaker(ctx, mmAddr, req.PairId)
		if !found {
			return &types.QueryMarketMakersResponse{}, nil
		}
		return &types.QueryMarketMakersResponse{
			Marketmakers: []types.MarketMaker{
				mm,
			},
		}, nil
	}

	store := ctx.KVStore(k.storeKey)

	var keyPrefix = types.MarketMakerKeyPrefix
	switch {
	case req.PairId != 0:
		keyPrefix = types.GetMarketMakerByPairIdPrefix(req.PairId)
	case !mmAddr.Empty():
		keyPrefix = types.GetMarketMakerByAddrPrefix(mmAddr)
	}

	mmStore := prefix.NewStore(store, keyPrefix)

	var mmsRes []types.MarketMaker
	pageRes, err := query.FilteredPaginate(mmStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		var mm types.MarketMaker

		switch {
		case req.PairId != 0:
			pairId, mmAddr := types.ParseMarketMakerIndexByPairIdKey(append(keyPrefix, key...))
			mm, _ = k.GetMarketMaker(ctx, mmAddr, pairId)
		default:
			mm, err = types.UnmarshalMarketMaker(k.cdc, value)
			if err != nil {
				return false, err
			}
		}

		if req.Eligible != "" && mm.Eligible != eligible {
			return false, nil
		}

		if accumulate {
			mmsRes = append(mmsRes, mm)
		}
		return true, nil
	})

	return &types.QueryMarketMakersResponse{Marketmakers: mmsRes, Pagination: pageRes}, nil
}

// Incentive queries all queued stakings of the farmer.
func (k Querier) Incentive(c context.Context, req *types.QueryIncentiveRequest) (*types.QueryIncentiveResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	var mmAddr sdk.AccAddress
	var err error
	mmAddr, err = sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	incentive, found := k.GetIncentive(ctx, mmAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "incentive for %s doesn't exist", req.Address)
	}

	return &types.QueryIncentiveResponse{Incentive: incentive}, nil
}
