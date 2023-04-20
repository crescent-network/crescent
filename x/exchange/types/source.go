package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SpotOrderSource interface {
	RequestTransientSpotOrders(ctx sdk.Context, market SpotMarket, isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int)
}
