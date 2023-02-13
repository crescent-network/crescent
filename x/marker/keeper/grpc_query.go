package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/marker/types"
)

// Querier is used as Keeper will have duplicate methods if used directly,
// and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the lending module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.Keeper.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// LastBlockTime queries the last block time.
func (k Querier) LastBlockTime(c context.Context, _ *types.QueryLastBlockTimeRequest) (*types.QueryLastBlockTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var lastBlockTimePtr *time.Time
	lastBlockTime, found := k.GetLastBlockTime(ctx)
	if found {
		lastBlockTimePtr = &lastBlockTime
	}
	return &types.QueryLastBlockTimeResponse{LastBlockTime: lastBlockTimePtr}, nil
}
