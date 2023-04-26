package keeper

import (
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
