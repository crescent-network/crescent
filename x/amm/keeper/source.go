package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) RequestTransientSpotOrders(ctx sdk.Context, market exchangetypes.SpotMarket, isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) {
	k.IteratePoolsByMarket(ctx, market.Id, func(pool types.Pool) (stop bool) {
		remainingQty := qty
		poolState := k.MustGetPoolState(ctx, pool.Id)
		reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
		reserve := k.bankKeeper.SpendableCoins(ctx, reserveAddr).AmountOf(pool.DenomOut(isBuy))
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
		var tickDelta int32
		cb := func(tick int32, liquidity sdk.Dec) (stop bool) {
			// TODO: check out of tick range
			price := exchangetypes.PriceAtTick(tick, TickPrecision)
			sqrtPrice := utils.DecApproxSqrt(price)
			sqrtPriceBefore := types.SqrtPriceAtTick(tick+tickDelta, TickPrecision)
			var qty sdk.Int
			if isBuy {
				sqrtPriceBefore = sdk.MinDec(poolState.CurrentSqrtPrice, sqrtPriceBefore)
				qty = utils.MinInt(
					reserve.ToDec().QuoTruncate(price).TruncateInt(),
					types.Amount1DeltaRounding(sqrtPriceBefore, sqrtPrice, liquidity, false).ToDec().QuoTruncate(price).TruncateInt())
			} else {
				sqrtPriceBefore = sdk.MaxDec(poolState.CurrentSqrtPrice, sqrtPriceBefore)
				qty = utils.MinInt(
					reserve,
					types.Amount0DeltaRounding(sqrtPriceBefore, sqrtPrice, liquidity, false))
			}
			if qty.IsPositive() {
				if err := k.exchangeKeeper.CreateTransientSpotOrder(
					ctx, market, reserveAddr, isBuy, price, qty, true); err != nil {
					panic(err)
				}
				reserve = reserve.Sub(exchangetypes.DepositAmount(isBuy, price, qty))
				remainingQty = remainingQty.Sub(qty)
			}
			return !(remainingQty.IsPositive() && reserve.IsPositive())
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
