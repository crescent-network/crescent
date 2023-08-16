package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
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

// PublicPositions queries all public position objects.
func (k Querier) PublicPositions(c context.Context, req *types.QueryPublicPositionsRequest) (*types.QueryPublicPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	if req.PoolId > 0 {
		if found := k.ammKeeper.LookupPool(ctx, req.PoolId); !found {
			return nil, status.Error(codes.NotFound, "pool not found")
		}
	}

	store := ctx.KVStore(k.storeKey)
	var (
		keyPrefix            []byte
		publicPositionGetter func(key, value []byte) types.PublicPosition
	)
	if req.PoolId > 0 {
		keyPrefix = types.GetPublicPositionsByPoolIteratorPrefix(req.PoolId)
		publicPositionGetter = func(key, _ []byte) types.PublicPosition {
			_, publicPositionId := types.ParsePublicPositionsByPoolIndexKey(utils.Key(keyPrefix, key))
			publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
			if !found { // sanity check
				panic("public position not found")
			}
			return publicPosition
		}
	} else {
		keyPrefix = types.PublicPositionKeyPrefix
		publicPositionGetter = func(_, value []byte) types.PublicPosition {
			var publicPosition types.PublicPosition
			k.cdc.MustUnmarshal(value, &publicPosition)
			return publicPosition
		}
	}
	publicPositionStore := prefix.NewStore(store, keyPrefix)
	var publicPositions []types.PublicPositionResponse
	pageRes, err := query.Paginate(publicPositionStore, req.Pagination, func(key, value []byte) error {
		publicPosition := publicPositionGetter(key, value)
		position, found := k.GetAMMPosition(ctx, publicPosition)
		if !found {
			position.Liquidity = utils.ZeroInt
		}
		shareDenom := types.ShareDenom(publicPosition.Id)
		publicPositions = append(publicPositions, types.PublicPositionResponse{
			Id:                   publicPosition.Id,
			PoolId:               publicPosition.PoolId,
			LowerTick:            publicPosition.LowerTick,
			UpperTick:            publicPosition.UpperTick,
			BidReserveAddress:    publicPosition.BidReserveAddress,
			MinBidAmount:         publicPosition.MinBidAmount,
			FeeRate:              publicPosition.FeeRate,
			LastRewardsAuctionId: publicPosition.LastRewardsAuctionId,
			Liquidity:            position.Liquidity,
			TotalShare:           k.bankKeeper.GetSupply(ctx, shareDenom),
			PositionId:           k.MustGetAMMPosition(ctx, publicPosition).Id,
		})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPublicPositionsResponse{
		PublicPositions: publicPositions,
		Pagination:      pageRes,
	}, nil
}

// PublicPosition queries the particular public position object.
func (k Querier) PublicPosition(c context.Context, req *types.QueryPublicPositionRequest) (*types.QueryPublicPositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.PublicPositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "public position id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	publicPosition, found := k.GetPublicPosition(ctx, req.PublicPositionId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "public position not found")
	}

	ammPosition, found := k.GetAMMPosition(ctx, publicPosition)
	if !found { // sanity check
		ammPosition.Liquidity = utils.ZeroInt
	}
	shareDenom := types.ShareDenom(publicPosition.Id)
	resp := types.PublicPositionResponse{
		Id:                   publicPosition.Id,
		PoolId:               publicPosition.PoolId,
		LowerTick:            publicPosition.LowerTick,
		UpperTick:            publicPosition.UpperTick,
		BidReserveAddress:    publicPosition.BidReserveAddress,
		MinBidAmount:         publicPosition.MinBidAmount,
		FeeRate:              publicPosition.FeeRate,
		LastRewardsAuctionId: publicPosition.LastRewardsAuctionId,
		Liquidity:            ammPosition.Liquidity,
		TotalShare:           k.bankKeeper.GetSupply(ctx, shareDenom),
		PositionId:           k.MustGetAMMPosition(ctx, publicPosition).Id,
	}
	return &types.QueryPublicPositionResponse{PublicPosition: resp}, nil
}

// RewardsAuctions queries all RewardsAuction objects.
func (k Querier) RewardsAuctions(c context.Context, req *types.QueryRewardsAuctionsRequest) (*types.QueryRewardsAuctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.PublicPositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "public position id must not be 0")
	}
	if req.Status != "" && !(req.Status == types.AuctionStatusStarted.String() ||
		req.Status == types.AuctionStatusFinished.String() ||
		req.Status == types.AuctionStatusSkipped.String()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid auction status %s", req.Status)
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPublicPosition(ctx, req.PublicPositionId); !found {
		return nil, status.Error(codes.NotFound, "public position not found")
	}
	store := ctx.KVStore(k.storeKey)
	auctionStore := prefix.NewStore(
		store, types.GetRewardsAuctionsByPublicPositionIteratorPrefix(req.PublicPositionId))
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
	if req.PublicPositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "public position id must not be 0")
	}
	if req.AuctionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "auction id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPublicPosition(ctx, req.PublicPositionId); !found {
		return nil, status.Error(codes.NotFound, "public position not found")
	}
	auction, found := k.GetRewardsAuction(ctx, req.PublicPositionId, req.AuctionId)
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
	if req.PublicPositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "public position id must not be 0")
	}
	if req.AuctionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "auction id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if found := k.LookupPublicPosition(ctx, req.PublicPositionId); !found {
		return nil, status.Error(codes.NotFound, "public position not found")
	}
	if found := k.LookupRewardsAuction(ctx, req.PublicPositionId, req.AuctionId); !found {
		return nil, status.Error(codes.NotFound, "auction not found")
	}
	store := ctx.KVStore(k.storeKey)
	bidStore := prefix.NewStore(
		store,
		types.GetBidsByRewardsAuctionIteratorPrefix(req.PublicPositionId, req.AuctionId))
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

// Rewards queries all rewards accumulated for the public position.
func (k Querier) Rewards(c context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.PublicPositionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "public position id must not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)
	publicPosition, found := k.GetPublicPosition(ctx, req.PublicPositionId)
	if !found {
		return nil, status.Error(codes.NotFound, "public position not found")
	}
	ammPosition, found := k.GetAMMPosition(ctx, publicPosition)
	var rewards sdk.Coins
	if found {
		fee, farmingRewards, err := k.ammKeeper.CollectibleCoins(ctx, ammPosition.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		rewards = fee.Add(farmingRewards...)
	}
	return &types.QueryRewardsResponse{Rewards: rewards}, nil
}

// ExchangeRate queries exchange rate, such as mint rate and burn rate per 1 lashare.
func (k Querier) ExchangeRate(c context.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	publicPosition, found := k.GetPublicPosition(ctx, req.PublicPositionId)
	if !found {
		return nil, status.Error(codes.NotFound, "public position not found")
	}
	res := &types.QueryExchangeRateResponse{
		MintRate: sdk.ZeroDec(),
		BurnRate: sdk.ZeroDec(),
	}
	shareSupply := k.bankKeeper.GetSupply(ctx, types.ShareDenom(publicPosition.Id)).Amount
	if !shareSupply.IsZero() {
		position := k.MustGetAMMPosition(ctx, publicPosition)
		var prevWinningBidShareAmt sdk.Int
		prevAuction, found := k.GetPreviousRewardsAuction(ctx, publicPosition)
		if found && prevAuction.WinningBid != nil {
			prevWinningBidShareAmt = prevAuction.WinningBid.Share.Amount
		} else {
			prevWinningBidShareAmt = utils.ZeroInt
		}
		res.MintRate = types.CalculateMintRate(position.Liquidity, shareSupply)
		res.BurnRate = types.CalculateBurnRate(shareSupply, position.Liquidity, prevWinningBidShareAmt)
	}
	return res, nil
}
