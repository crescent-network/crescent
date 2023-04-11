package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExchangeHooks interface {
	AfterRestingSpotOrderExecuted(ctx sdk.Context, order SpotLimitOrder, qty sdk.Int)
	AfterSpotOrderExecuted(ctx sdk.Context, market SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, firstPrice, lastPrice sdk.Dec, qty, quoteAmt sdk.Int)
}

type MultiExchangeHooks []ExchangeHooks

func NewMultiExchangeHooks(hooks ...ExchangeHooks) MultiExchangeHooks {
	return hooks
}

func (hs MultiExchangeHooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order SpotLimitOrder, qty sdk.Int) {
	for _, h := range hs {
		h.AfterRestingSpotOrderExecuted(ctx, order, qty)
	}
}

func (hs MultiExchangeHooks) AfterSpotOrderExecuted(ctx sdk.Context, market SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, firstPrice, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	for _, h := range hs {
		h.AfterSpotOrderExecuted(ctx, market, ordererAddr, isBuy, firstPrice, lastPrice, qty, quoteAmt)
	}
}
