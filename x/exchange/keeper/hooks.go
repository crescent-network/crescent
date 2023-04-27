package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.ExchangeHooks = Keeper{}

func (k Keeper) AfterOrderExecuted(ctx sdk.Context, order types.Order, execQty sdk.Int, paid, received, fee sdk.Coin) {
	if k.hooks != nil {
		k.hooks.AfterOrderExecuted(ctx, order, execQty, paid, received, fee)
	}
}
