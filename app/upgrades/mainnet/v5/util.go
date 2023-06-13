package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// AdjustPriceToTickSpacing returns rounded tick price considering the tick
// spacing.
func AdjustPriceToTickSpacing(price sdk.Dec, tickSpacing uint32, roundUp bool) sdk.Dec {
	// We assume that the price is already a valid tick price, so here's
	// no check for that.
	tick := exchangetypes.TickAtPrice(price)
	ts := int32(tickSpacing)
	if roundUp {
		q, _ := utils.DivMod(tick+ts-1, ts)
		return exchangetypes.PriceAtTick(q * ts)
	}
	q, _ := utils.DivMod(tick, ts)
	return exchangetypes.PriceAtTick(q * ts)
}
