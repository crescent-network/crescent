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

func (k Keeper) GetDefaultFees(ctx sdk.Context) (fees types.Fees) {
	k.paramSpace.Get(ctx, types.KeyDefaultFees, &fees)
	return
}

func (k Keeper) SetDefaultFees(ctx sdk.Context, fees types.Fees) {
	k.paramSpace.Set(ctx, types.KeyDefaultFees, fees)
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

func (k Keeper) GetDefaultOrderQuantityLimits(ctx sdk.Context) (orderQtyLimits types.AmountLimits) {
	k.paramSpace.Get(ctx, types.KeyDefaultOrderQuantityLimits, &orderQtyLimits)
	return
}

func (k Keeper) SetDefaultOrderQuantityLimits(ctx sdk.Context, orderQtyLimits types.AmountLimits) {
	k.paramSpace.Set(ctx, types.KeyDefaultOrderQuantityLimits, orderQtyLimits)
}

func (k Keeper) GetDefaultOrderQuoteLimits(ctx sdk.Context) (orderQuoteLimits types.AmountLimits) {
	k.paramSpace.Get(ctx, types.KeyDefaultOrderQuoteLimits, &orderQuoteLimits)
	return
}

func (k Keeper) SetDefaultOrderQuoteLimits(ctx sdk.Context, orderQuoteLimits sdk.Int) {
	k.paramSpace.Set(ctx, types.KeyDefaultOrderQuoteLimits, orderQuoteLimits)
}


func (k Keeper) GetMaxSwapRoutesLen(ctx sdk.Context) (maxLen uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxSwapRoutesLen, &maxLen)
	return
}

func (k Keeper) SetMaxSwapRouteLen(ctx sdk.Context, maxLen uint32) {
	k.paramSpace.Set(ctx, types.KeyMaxSwapRoutesLen, maxLen)
}

func (k Keeper) GetMaxNumMMOrders(ctx sdk.Context) (maxNum uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxNumMMOrders, &maxNum)
	return
}

func (k Keeper) SetMaxNumMMOrders(ctx sdk.Context, maxNum uint32) {
	k.paramSpace.Set(ctx, types.KeyMaxNumMMOrders, maxNum)
}
