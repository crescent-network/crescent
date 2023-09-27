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

func (k Keeper) GetFees(ctx sdk.Context) (fees types.Fees) {
	k.paramSpace.Get(ctx, types.KeyFees, &fees)
	return
}

func (k Keeper) SetFees(ctx sdk.Context, fees types.Fees) {
	k.paramSpace.Set(ctx, types.KeyFees, fees)
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

func (k Keeper) GetDefaultMinOrderQuantity(ctx sdk.Context) (minOrderQty sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyDefaultMinOrderQuantity, &minOrderQty)
	return
}

func (k Keeper) SetDefaultMinOrderQuantity(ctx sdk.Context, minOrderQty sdk.Int) {
	k.paramSpace.Set(ctx, types.KeyDefaultMinOrderQuantity, minOrderQty)
}

func (k Keeper) GetDefaultMinOrderQuote(ctx sdk.Context) (minOrderQuote sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyDefaultMinOrderQuote, &minOrderQuote)
	return
}

func (k Keeper) SetDefaultMinOrderQuote(ctx sdk.Context, minOrderQuote sdk.Int) {
	k.paramSpace.Set(ctx, types.KeyDefaultMinOrderQuote, minOrderQuote)
}

func (k Keeper) GetDefaultMaxOrderQuantity(ctx sdk.Context) (maxOrderQty sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyDefaultMaxOrderQuantity, &maxOrderQty)
	return
}

func (k Keeper) SetDefaultMaxOrderQuantity(ctx sdk.Context, maxOrderQty sdk.Int) {
	k.paramSpace.Set(ctx, types.KeyDefaultMaxOrderQuantity, maxOrderQty)
}

func (k Keeper) GetDefaultMaxOrderQuote(ctx sdk.Context) (maxOrderQuote sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyDefaultMaxOrderQuote, &maxOrderQuote)
	return
}

func (k Keeper) SetDefaultMaxOrderQuote(ctx sdk.Context, maxOrderQuote sdk.Int) {
	k.paramSpace.Set(ctx, types.KeyDefaultMaxOrderQuote, maxOrderQuote)
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
