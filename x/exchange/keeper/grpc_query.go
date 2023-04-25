package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

func (k Querier) BestSwapExactInRoutes(c context.Context, req *types.QueryBestSwapExactInRoutesRequest) (*types.QueryBestSwapExactInRoutesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	allRoutes := k.FindAllRoutes(ctx, req.Input.Denom, req.MinOutput.Denom)
	if len(allRoutes) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "no possible routes")
	}

	var (
		bestOutput = utils.ZeroInt
		bestRoutes []uint64
	)
	for _, routes := range allRoutes {
		output, err := k.SwapExactIn(ctx, sdk.AccAddress{}, routes, req.Input, req.MinOutput, true)
		if err != nil && !errors.Is(err, types.ErrInsufficientOutput) { // sanity check
			panic(err)
		}
		fmt.Println(routes, output)
		if err == nil {
			if output.Amount.GT(bestOutput) {
				bestOutput = output.Amount
				bestRoutes = routes
			}
		}
	}

	if bestOutput.LT(req.MinOutput.Amount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "no possible routes") // TODO: use different error
	}

	return &types.QueryBestSwapExactInRoutesResponse{
		Routes: bestRoutes,
		Output: sdk.NewCoin(req.MinOutput.Denom, bestOutput),
	}, nil
}
