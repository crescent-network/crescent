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
	var tickLimit int32
	if priceLimit == nil {
		minTick, maxTick := exchangetypes.MinMaxTick(TickPrecision)
		if isBuy {
			tickLimit = minTick
		} else {
			tickLimit = maxTick
		}
	} else {
		tickLimit = exchangetypes.TickAtPrice(*priceLimit, TickPrecision)
	}
	k.IteratePoolsByMarket(ctx, market.Id, func(pool types.Pool) (stop bool) {
		accQty := utils.ZeroInt
		accQuote := utils.ZeroInt
		poolState := k.MustGetPoolState(ctx, pool.Id)
		reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
		reserve := k.bankKeeper.SpendableCoins(ctx, reserveAddr).AmountOf(pool.DenomOut(isBuy))
		var tickDelta int32
		cb := func(tick int32, liquidity sdk.Dec) (stop bool) {
			if qtyLimit != nil && !qtyLimit.Sub(accQty).IsPositive() {
				return true
			}
			if quoteLimit != nil && !quoteLimit.Sub(accQuote).IsPositive() {
				return true
			}
			if !reserve.IsPositive() {
				return true
			}
			// TODO: check out of tick range
			price := exchangetypes.PriceAtTick(tick, TickPrecision)
			sqrtPrice := utils.DecApproxSqrt(price)
			sqrtPriceBefore := types.SqrtPriceAtTick(tick+tickDelta, TickPrecision)
			var qty sdk.Int
			if isBuy {
				sqrtPriceBefore = sdk.MinDec(utils.DecApproxSqrt(poolState.CurrentPrice), sqrtPriceBefore)
				qty = utils.MinInt(
					reserve.ToDec().QuoTruncate(price).TruncateInt(),
					types.Amount1DeltaRounding(sqrtPriceBefore, sqrtPrice, liquidity, false).ToDec().QuoTruncate(price).TruncateInt())
			} else {
				sqrtPriceBefore = sdk.MaxDec(utils.DecApproxSqrt(poolState.CurrentPrice), sqrtPriceBefore)
				qty = utils.MinInt(
					reserve,
					types.Amount0DeltaRounding(sqrtPriceBefore, sqrtPrice, liquidity, false))
			}
			if qty.IsPositive() {
				if err := k.exchangeKeeper.CreateTransientOrder(
					ctx, market, reserveAddr, isBuy, price, qty, true); err != nil {
					panic(err)
				}
				reserve = reserve.Sub(exchangetypes.DepositAmount(isBuy, price, qty))
				accQty = accQty.Add(qty)
				accQuote = accQuote.Add(exchangetypes.QuoteAmount(!isBuy, price, qty))
			}
			return false
		}
		if isBuy {
			tickDelta = int32(pool.TickSpacing)
			k.iterateTicksBelowPoolPriceWithLiquidity(ctx, pool, poolState, tickLimit, cb)
		} else {
			tickDelta = -int32(pool.TickSpacing)
			k.iterateTicksAbovePoolPriceWithLiquidity(ctx, pool, poolState, tickLimit, cb)
		}
		return true // Only one pool can participate in matching
	})
}
