package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Airdrops queries all the existing airdrops.
func (k Querier) Airdrops(c context.Context, req *types.QueryAirdropsRequest) (*types.QueryAirdropsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	airdropStore := prefix.NewStore(store, types.AirdropKeyPrefix)

	airdrops := []types.Airdrop{}
	pageRes, err := query.FilteredPaginate(airdropStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		var airdrop types.Airdrop
		k.cdc.MustUnmarshal(value, &airdrop)

		if accumulate {
			airdrops = append(airdrops, airdrop)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAirdropsResponse{Airdrops: airdrops, Pagination: pageRes}, nil
}

// Airdrop queries the specific airdrop.
func (k Querier) Airdrop(c context.Context, req *types.QueryAirdropRequest) (*types.QueryAirdropResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	airdrop, found := k.Keeper.GetAirdrop(ctx, req.AirdropId)
	if !found {
		return nil, status.Error(codes.NotFound, "airdrop not found")
	}

	return &types.QueryAirdropResponse{Airdrop: airdrop}, nil
}

// ClaimRecord queries the specific claim record.
func (k Querier) ClaimRecord(c context.Context, req *types.QueryClaimRecordRequest) (*types.QueryClaimRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	recipientAddr, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	record, found := k.GetClaimRecordByRecipient(ctx, req.AirdropId, recipientAddr)
	if !found {
		return nil, status.Error(codes.NotFound, "claim record not found")
	}

	return &types.QueryClaimRecordResponse{ClaimRecord: record}, nil
}
