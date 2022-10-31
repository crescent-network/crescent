package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

// GetParams returns the parameters for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

// SetParams sets the parameters for the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetPrivatePlanCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPrivatePlanCreationFee, &fee)
	return
}

func (k Keeper) SetPrivatePlanCreationFee(ctx sdk.Context, fee sdk.Coins) {
	k.paramSpace.Set(ctx, types.KeyPrivatePlanCreationFee, fee)
}

func (k Keeper) GetFeeCollector(ctx sdk.Context) (feeCollector string) {
	k.paramSpace.Get(ctx, types.KeyFeeCollector, &feeCollector)
	return
}

func (k Keeper) SetFeeCollector(ctx sdk.Context, feeCollector string) {
	k.paramSpace.Set(ctx, types.KeyFeeCollector, feeCollector)
}

func (k Keeper) GetMaxNumPrivatePlans(ctx sdk.Context) (num uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxNumPrivatePlans, &num)
	return
}

func (k Keeper) SetMaxNumPrivatePlans(ctx sdk.Context, num uint32) {
	k.paramSpace.Set(ctx, types.KeyMaxNumPrivatePlans, num)
}

func (k Keeper) GetMaxBlockDuration(ctx sdk.Context) (d time.Duration) {
	k.paramSpace.Get(ctx, types.KeyMaxBlockDuration, &d)
	return
}

func (k Keeper) SetMaxBlockDuration(ctx sdk.Context, d time.Duration) {
	k.paramSpace.Set(ctx, types.KeyMaxBlockDuration, d)
}
