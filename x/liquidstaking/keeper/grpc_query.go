package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the liquidstaking module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// LiquidValidators queries all liquid validators.
func (k Querier) LiquidValidators(c context.Context, req *types.QueryLiquidValidatorsRequest) (*types.QueryLiquidValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryLiquidValidatorsResponse{LiquidValidators: k.GetAllLiquidValidatorStates(ctx)}, nil
}

// States queries states of liquid staking module.
func (k Querier) States(c context.Context, req *types.QueryStatesRequest) (*types.QueryStatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryStatesResponse{NetAmountState: k.GetNetAmountState(ctx)}, nil
}

// VotingPower queries voting power of staking, liquid staking module's for the voter.
func (k Querier) VotingPower(c context.Context, req *types.QueryVotingPowerRequest) (*types.QueryVotingPowerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	addr, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, err
	}
	return &types.QueryVotingPowerResponse{VotingPower: k.GetVotingPower(ctx, addr)}, nil
}
