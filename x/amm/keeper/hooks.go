package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.ExchangeHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotLimitOrder, qty sdk.Int) {
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	// Update pool orders from the market's last price to the new last price
}
