package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the module.
func (k Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// LiquidFarms queries all LiquidFarm objects.
func (k Querier) LiquidFarms(c context.Context, req *types.QueryLiquidFarmsRequest) (*types.QueryLiquidFarmsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	liquidFarmStore := prefix.NewStore(store, types.LiquidFarmKeyPrefix)
	var liquidFarms []types.LiquidFarmResponse
	pageRes, err := query.Paginate(liquidFarmStore, req.Pagination, func(_, value []byte) error {
		var liquidFarm types.LiquidFarm
		k.cdc.MustUnmarshal(value, &liquidFarm)
		position, found := k.GetLiquidFarmPosition(ctx, liquidFarm)
		if !found {
			position.Liquidity = utils.ZeroInt
		}
		shareDenom := types.ShareDenom(liquidFarm.Id)
		liquidFarms = append(liquidFarms, types.LiquidFarmResponse{
			Id:                   liquidFarm.Id,
			PoolId:               liquidFarm.PoolId,
			LowerTick:            liquidFarm.LowerTick,
			UpperTick:            liquidFarm.UpperTick,
			BidReserveAddress:    liquidFarm.BidReserveAddress,
			MinBidAmount:         liquidFarm.MinBidAmount,
			FeeRate:              liquidFarm.FeeRate,
			LastRewardsAuctionId: liquidFarm.LastRewardsAuctionId,
			Liquidity:            position.Liquidity,
			TotalShare:           k.bankKeeper.GetSupply(ctx, shareDenom),
		})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryLiquidFarmsResponse{
		LiquidFarms: liquidFarms,
		Pagination:  pageRes,
	}, nil
}

// LiquidFarm queries the particular LiquidFarm object.
func (k Querier) LiquidFarm(c context.Context, req *types.QueryLiquidFarmRequest) (*types.QueryLiquidFarmResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LiquidFarmId == 0 {
		return nil, status.Error(codes.InvalidArgument, "liquid farm id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	liquidFarm, found := k.GetLiquidFarm(ctx, req.LiquidFarmId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "liquid farm not found")
	}

	position, found := k.GetLiquidFarmPosition(ctx, liquidFarm)
	if !found { // sanity check
		position.Liquidity = utils.ZeroInt
	}
	shareDenom := types.ShareDenom(liquidFarm.Id)
	liquidFarmResp := types.LiquidFarmResponse{
		Id:                   liquidFarm.Id,
		PoolId:               liquidFarm.PoolId,
		LowerTick:            liquidFarm.LowerTick,
		UpperTick:            liquidFarm.UpperTick,
		BidReserveAddress:    liquidFarm.BidReserveAddress,
		MinBidAmount:         liquidFarm.MinBidAmount,
		FeeRate:              liquidFarm.FeeRate,
		LastRewardsAuctionId: liquidFarm.LastRewardsAuctionId,
		Liquidity:            position.Liquidity,
		TotalShare:           k.bankKeeper.GetSupply(ctx, shareDenom),
	}
	return &types.QueryLiquidFarmResponse{LiquidFarm: liquidFarmResp}, nil
}

// RewardsAuctions queries all RewardsAuction objects.
func (k Querier) RewardsAuctions(c context.Context, req *types.QueryRewardsAuctionsRequest) (*types.QueryRewardsAuctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LiquidFarmId == 0 {
		return nil, status.Error(codes.InvalidArgument, "liquid farm id must not be 0")
	}
	if req.Status != "" && !(req.Status == types.AuctionStatusStarted.String() ||
		req.Status == types.AuctionStatusFinished.String() ||
		req.Status == types.AuctionStatusSkipped.String()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid auction status %s", req.Status)
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupLiquidFarm(ctx, req.LiquidFarmId); !found {
		return nil, status.Error(codes.NotFound, "liquid farm not found")
	}
	store := ctx.KVStore(k.storeKey)
	auctionStore := prefix.NewStore(
		store, types.GetRewardsAuctionsByLiquidFarmIteratorPrefix(req.LiquidFarmId))
	var auctions []types.RewardsAuction
	pageRes, err := query.FilteredPaginate(auctionStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var auction types.RewardsAuction
		k.cdc.MustUnmarshal(value, &auction)
		if req.Status != "" && auction.Status.String() != req.Status {
			return false, nil
		}
		if accumulate {
			auctions = append(auctions, auction)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryRewardsAuctionsResponse{
		RewardsAuctions: auctions,
		Pagination:      pageRes,
	}, nil
}

// RewardsAuction queries the particular RewardsAuction object.
func (k Querier) RewardsAuction(c context.Context, req *types.QueryRewardsAuctionRequest) (*types.QueryRewardsAuctionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LiquidFarmId == 0 {
		return nil, status.Error(codes.InvalidArgument, "liquid farm id must not be 0")
	}
	if req.AuctionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "auction id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupLiquidFarm(ctx, req.LiquidFarmId); !found {
		return nil, status.Error(codes.NotFound, "liquid farm not found")
	}
	auction, found := k.GetRewardsAuction(ctx, req.LiquidFarmId, req.AuctionId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "auction not found")
	}
	return &types.QueryRewardsAuctionResponse{RewardsAuction: auction}, nil
}

// Bids queries all Bid objects.
func (k Querier) Bids(c context.Context, req *types.QueryBidsRequest) (*types.QueryBidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LiquidFarmId == 0 {
		return nil, status.Error(codes.InvalidArgument, "liquid farm id must not be 0")
	}
	if req.AuctionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "auction id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupLiquidFarm(ctx, req.LiquidFarmId); !found {
		return nil, status.Error(codes.NotFound, "liquid farm not found")
	}
	if found := k.LookupRewardsAuction(ctx, req.LiquidFarmId, req.AuctionId); !found {
		return nil, status.Error(codes.NotFound, "auction not found")
	}
	store := ctx.KVStore(k.storeKey)
	bidStore := prefix.NewStore(
		store,
		types.GetBidsByRewardsAuctionIteratorPrefix(req.LiquidFarmId, req.AuctionId))
	var bids []types.Bid
	pageRes, err := query.FilteredPaginate(bidStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var bid types.Bid
		k.cdc.MustUnmarshal(value, &bid)
		if accumulate {
			bids = append(bids, bid)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryBidsResponse{Bids: bids, Pagination: pageRes}, nil
}

// Rewards queries all rewards accumulated for the liquid farm.
func (k Querier) Rewards(c context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LiquidFarmId == 0 {
		return nil, status.Error(codes.InvalidArgument, "liquid farm id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	liquidFarm, found := k.GetLiquidFarm(ctx, req.LiquidFarmId)
	if !found {
		return nil, status.Error(codes.NotFound, "liquid farm not found")
	}
	position, found := k.GetLiquidFarmPosition(ctx, liquidFarm)
	var rewards sdk.Coins
	if found {
		fee, farmingRewards, err := k.ammKeeper.CollectibleCoins(ctx, position.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		rewards = fee.Add(farmingRewards...)
	}
	return &types.QueryRewardsResponse{Rewards: rewards}, nil
}
