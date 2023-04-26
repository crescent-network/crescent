package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OrderSource interface {
	RequestTransientOrders(ctx sdk.Context, market Market, isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int)
}
