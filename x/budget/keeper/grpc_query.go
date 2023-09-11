package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the budget module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// Budgets queries all budgets.
func (k Querier) Budgets(c context.Context, req *types.QueryBudgetsRequest) (*types.QueryBudgetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.SourceAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.SourceAddress); err != nil {
			return nil, err
		}
	}

	if req.DestinationAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.DestinationAddress); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	var budgets []types.BudgetResponse
	for _, b := range params.Budgets {
		if req.Name != "" && b.Name != req.Name ||
			req.SourceAddress != "" && b.SourceAddress != req.SourceAddress ||
			req.DestinationAddress != "" && b.DestinationAddress != req.DestinationAddress {
			continue
		}

		collectedCoins := k.GetTotalCollectedCoins(ctx, b.Name)
		budgets = append(budgets, types.BudgetResponse{
			Budget:              b,
			TotalCollectedCoins: collectedCoins,
		})
	}

	return &types.QueryBudgetsResponse{Budgets: budgets}, nil
}

// Addresses queries an address that can be used as source and destination is derived according to the given name, module name and address type.
func (k Querier) Addresses(_ context.Context, req *types.QueryAddressesRequest) (*types.QueryAddressesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Name == "" && req.ModuleName == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one input of name or module name is required")
	}

	if req.ModuleName == "" && req.Type == types.AddressType32Bytes {
		req.ModuleName = types.ModuleName
	}

	addr := types.DeriveAddress(req.Type, req.ModuleName, req.Name)
	if addr.Empty() {
		return nil, status.Error(codes.InvalidArgument, "invalid names with address type")
	}

	return &types.QueryAddressesResponse{Address: addr.String()}, nil
}
