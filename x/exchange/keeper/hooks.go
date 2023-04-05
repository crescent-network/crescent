package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.ExchangeHooks = Keeper{}

func (k Keeper) AfterRestingSpotOrderExecuted(ctx sdk.Context, order types.SpotLimitOrder, qty sdk.Int) {
	if k.hooks != nil {
		k.hooks.AfterRestingSpotOrderExecuted(ctx, order, qty)
	}
}

func (k Keeper) AfterSpotOrderExecuted(ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	if k.hooks != nil {
		k.hooks.AfterSpotOrderExecuted(ctx, market, ordererAddr, isBuy, lastPrice, qty, quoteAmt)
	}
}
