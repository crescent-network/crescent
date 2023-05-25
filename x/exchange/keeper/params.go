package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
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

func (k Keeper) GetMarketCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyMarketCreationFee, &fee)
	return
}

func (k Keeper) SetMarketCreationFee(ctx sdk.Context, fee sdk.Coins) {
	k.paramSpace.Set(ctx, types.KeyMarketCreationFee, fee)
}

func (k Keeper) GetDefaultMakerFeeRate(ctx sdk.Context) (feeRate sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyDefaultMakerFeeRate, &feeRate)
	return
}

func (k Keeper) SetDefaultMakerFeeRate(ctx sdk.Context, feeRate sdk.Dec) {
	k.paramSpace.Set(ctx, types.KeyDefaultMakerFeeRate, feeRate)
}

func (k Keeper) GetDefaultTakerFeeRate(ctx sdk.Context) (feeRate sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyDefaultTakerFeeRate, &feeRate)
	return
}

func (k Keeper) SetDefaultTakerFeeRate(ctx sdk.Context, feeRate sdk.Dec) {
	k.paramSpace.Set(ctx, types.KeyDefaultTakerFeeRate, feeRate)
}

func (k Keeper) GetMaxOrderLifespan(ctx sdk.Context) (maxLifespan time.Duration) {
	k.paramSpace.Get(ctx, types.KeyMaxOrderLifespan, &maxLifespan)
	return
}

func (k Keeper) SetMaxOrderLifespan(ctx sdk.Context, maxLifespan time.Duration) {
	k.paramSpace.Set(ctx, types.KeyMaxOrderLifespan, maxLifespan)
}

func (k Keeper) GetMaxOrderPriceRatio(ctx sdk.Context) (maxRatio sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyMaxOrderPriceRatio, &maxRatio)
	return
}

func (k Keeper) SetMaxOrderPriceRatio(ctx sdk.Context, maxRatio sdk.Dec) {
	k.paramSpace.Set(ctx, types.KeyMaxOrderPriceRatio, maxRatio)
}

func (k Keeper) GetMaxSwapRoutesLen(ctx sdk.Context) (maxLen uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxSwapRoutesLen, &maxLen)
	return
}

func (k Keeper) SetMaxSwapRouteLen(ctx sdk.Context, maxLen uint32) {
	k.paramSpace.Set(ctx, types.KeyMaxSwapRoutesLen, maxLen)
}
