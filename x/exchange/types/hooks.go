package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExchangeHooks interface {
	AfterSpotOrderExecuted(ctx sdk.Context, order SpotLimitOrder, qty sdk.Int)
}

type MultiExchangeHooks []ExchangeHooks

func NewMultiExchangeHooks(hooks ...ExchangeHooks) MultiExchangeHooks {
	return hooks
}

func (hs MultiExchangeHooks) AfterSpotOrderExecuted(ctx sdk.Context, order SpotLimitOrder, qty sdk.Int) {
	for _, h := range hs {
		h.AfterSpotOrderExecuted(ctx, order, qty)
	}
}
