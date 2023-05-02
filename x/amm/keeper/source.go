package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.OrderSource = Keeper{}

func (k Keeper) RequestTransientOrders(
	ctx sdk.Context, market exchangetypes.Market, isBuy bool,
	priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int) {
	k.IteratePoolsByMarket(ctx, market.Id, func(pool types.Pool) (stop bool) {
		reserveAddr := pool.MustGetReserveAddress()
		accQty := utils.ZeroInt
		accQuote := utils.ZeroInt
		k.IteratePoolOrders(ctx, pool, isBuy, func(price sdk.Dec, qty sdk.Int) (stop bool) {
			if priceLimit != nil &&
				((isBuy && price.LT(*priceLimit)) ||
					(!isBuy && price.GT(*priceLimit))) {
				return true
			}
			if qtyLimit != nil && !qtyLimit.Sub(accQty).IsPositive() {
				return true
			}
			if quoteLimit != nil && !quoteLimit.Sub(accQuote).IsPositive() {
				return true
			}
			if err := k.exchangeKeeper.CreateTransientOrder(
				ctx, market, reserveAddr, isBuy, price, qty, true); err != nil {
				panic(err)
			}
			accQty = accQty.Add(qty)
			accQuote = accQuote.Add(exchangetypes.QuoteAmount(!isBuy, price, qty))
			return false
		})
		return true // Only one pool can participate in matching
	})
}
