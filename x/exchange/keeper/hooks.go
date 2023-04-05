package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.ExchangeHooks = Keeper{}

func (k Keeper) AfterSpotOrderExecuted(ctx sdk.Context, order types.SpotLimitOrder, qty sdk.Int) {
	if k.hooks != nil {
		k.hooks.AfterSpotOrderExecuted(ctx, order, qty)
	}
}
