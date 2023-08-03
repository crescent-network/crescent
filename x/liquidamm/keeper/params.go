package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// GetParams returns the parameters for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetRewardsAuctionDuration(ctx sdk.Context) (duration time.Duration) {
	k.paramSpace.Get(ctx, types.KeyRewardsAuctionDuration, &duration)
	return
}

func (k Keeper) SetRewardsAuctionDuration(ctx sdk.Context, duration time.Duration) {
	k.paramSpace.Set(ctx, types.KeyRewardsAuctionDuration, duration)
}

func (k Keeper) GetMaxNumRecentRewardsAuctions(ctx sdk.Context) (maxNum uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxNumRecentRewardsAuctions, &maxNum)
	return
}

func (k Keeper) SetMaxNumRecentRewardsAuctions(ctx sdk.Context, maxNum uint32) {
	k.paramSpace.Set(ctx, types.KeyMaxNumRecentRewardsAuctions, maxNum)
}
