package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/mint/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// LastBlockTime returns last block time.
func (k Keeper) LastBlockTime(c context.Context, _ *types.QueryLastBlockTimeRequest) (*types.QueryLastBlockTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryLastBlockTimeResponse{LastBlockTime: k.GetLastBlockTime(ctx)}, nil
}
