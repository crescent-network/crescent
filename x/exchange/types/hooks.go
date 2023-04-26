package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExchangeHooks interface {
	AfterOrderExecuted(ctx sdk.Context, order Order, qty sdk.Int)
}

type MultiExchangeHooks []ExchangeHooks

func NewMultiExchangeHooks(hooks ...ExchangeHooks) MultiExchangeHooks {
	return hooks
}

func (hs MultiExchangeHooks) AfterOrderExecuted(ctx sdk.Context, order Order, qty sdk.Int) {
	for _, h := range hs {
		h.AfterOrderExecuted(ctx, order, qty)
	}
}
