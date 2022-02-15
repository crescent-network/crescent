package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) ClaimRecord(c context.Context, req *types.QueryClaimRecordRequest) (*types.QueryClaimRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	recipientAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	record, found := k.GetClaimRecordByRecipient(ctx, req.AirdropId, recipientAddr)
	if !found {
		return nil, status.Error(codes.NotFound, "claim record not found")
	}

	return &types.QueryClaimRecordResponse{ClaimRecord: record}, nil
}
