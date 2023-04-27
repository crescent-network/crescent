package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExchangeHooks interface {
	AfterOrderExecuted(ctx sdk.Context, order Order, execQty sdk.Int, paid, received, fee sdk.Coin)
}

type MultiExchangeHooks []ExchangeHooks

func NewMultiExchangeHooks(hooks ...ExchangeHooks) MultiExchangeHooks {
	return hooks
}

func (hs MultiExchangeHooks) AfterOrderExecuted(ctx sdk.Context, order Order, execQty sdk.Int, paid, received, fee sdk.Coin) {
	for _, h := range hs {
		h.AfterOrderExecuted(ctx, order, execQty, paid, received, fee)
	}
}
