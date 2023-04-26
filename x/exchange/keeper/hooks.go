package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.ExchangeHooks = Keeper{}

func (k Keeper) AfterOrderExecuted(ctx sdk.Context, order types.Order, qty sdk.Int) {
	if k.hooks != nil {
		k.hooks.AfterOrderExecuted(ctx, order, qty)
	}
}
