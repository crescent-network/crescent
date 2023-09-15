package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/cremath"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func PoolOrders(
	isBuy bool, currentPrice cremath.BigDec, liquidity sdk.Int, reserveBalance sdk.Dec,
	nextTick int32, tickSpacing uint32,
	minOrderQty, minOrderQuote sdk.Dec, cb func(price, qty, openQty sdk.Dec) (stop bool)) (newPrice cremath.BigDec, newReserveBalance sdk.Dec) {
	for reserveBalance.IsPositive() {
		orderTick, valid := NextOrderTick(
			isBuy, liquidity, currentPrice, minOrderQty, minOrderQuote, tickSpacing)
		if !valid || ((isBuy && orderTick < nextTick) || (!isBuy && orderTick > nextTick)) {
			orderTick = nextTick
		}

		orderPrice := exchangetypes.PriceAtTick(orderTick)
		orderSqrtPrice := cremath.NewBigDecFromDec(orderPrice).SqrtMut()
		currentSqrtPrice := currentPrice.Sqrt()

		var qty, openQty sdk.Dec
		if isBuy {
			qty = Amount1DeltaRoundingDec(
				currentSqrtPrice, orderSqrtPrice, liquidity, false).QuoTruncate(orderPrice)
			openQty = sdk.MinDec(reserveBalance.QuoTruncate(orderPrice), qty)
		} else {
			qty = Amount0DeltaRoundingDec(
				currentSqrtPrice, orderSqrtPrice, liquidity, false)
			openQty = sdk.MinDec(reserveBalance, qty)
		}
		if openQty.IsPositive() && (orderTick == nextTick || (openQty.GTE(minOrderQty))) {
			if cb(orderPrice, qty, openQty) {
				break
			}
			reserveBalance = reserveBalance.Sub(
				exchangetypes.DepositAmount(isBuy, orderPrice, openQty))
			currentPrice = cremath.NewBigDecFromDec(orderPrice)
		} else { // No more possible order price
			break
		}
		if orderTick == nextTick {
			break
		}
	}
	return currentPrice, reserveBalance
}

func NextOrderTick(
	isBuy bool, liquidity sdk.Int, currentPrice cremath.BigDec, minOrderQty, minOrderQuote sdk.Dec, tickSpacing uint32) (tick int32, valid bool) {
	currentSqrtPrice := currentPrice.Sqrt()
	liquidityBigDec := cremath.NewBigDecFromInt(liquidity)
	minOrderQtyBigDec := cremath.NewBigDecFromDec(minOrderQty)
	minOrderQuoteBigDec := cremath.NewBigDecFromDec(minOrderQuote)
	if isBuy {
		// 1. Check min order qty
		// L^2 + 4 * MinQty * L * sqrt(P_current)
		intermediate := liquidityBigDec.Power(2).AddMut(
			minOrderQtyBigDec.Mul(liquidityBigDec).MulTruncateMut(currentSqrtPrice).MulInt64Mut(4))
		orderSqrtPrice := intermediate.SqrtMut().SubMut(liquidityBigDec).QuoTruncateMut(minOrderQtyBigDec.MulInt64(2))
		if !orderSqrtPrice.IsPositive() {
			return 0, false
		}
		// 2. Check min order quote
		orderSqrtPrice2 := currentSqrtPrice.Mul(liquidityBigDec).SubMut(minOrderQuoteBigDec).QuoTruncateMut(liquidityBigDec)
		if !orderSqrtPrice2.IsPositive() {
			return 0, false
		}
		orderPrice := cremath.MinBigDec(orderSqrtPrice, orderSqrtPrice2).PowerMut(2)
		if orderPrice.GT(currentPrice) {
			return 0, false
		}
		orderPriceDec := orderPrice.Dec()
		tick = AdjustPriceToTickSpacing(orderPriceDec, tickSpacing, false)
		if orderPrice.Equal(currentPrice) { // This implies PriceAtTick(tick) == currentPrice
			tick -= int32(tickSpacing)
		}
		return tick, true
	}
	// 1. Check min order qty
	orderSqrtPrice := currentSqrtPrice.Mul(liquidityBigDec).
		QuoRoundUpMut(liquidityBigDec.Sub(minOrderQtyBigDec.Mul(currentSqrtPrice)))
	if !orderSqrtPrice.IsPositive() {
		return 0, false
	}
	// 2. Check min order quote
	orderSqrtPrice2 := minOrderQuoteBigDec.Mul(currentSqrtPrice).QuoRoundUpMut(liquidityBigDec).AddMut(currentSqrtPrice)
	if !orderSqrtPrice2.IsPositive() {
		return 0, false
	}
	orderPrice := cremath.MaxBigDec(orderSqrtPrice, orderSqrtPrice2).PowerMut(2)
	if orderPrice.LT(currentPrice) {
		return 0, false
	}
	orderPriceDec := orderPrice.DecRoundUp()
	tick = AdjustPriceToTickSpacing(orderPriceDec, tickSpacing, true)
	if orderPrice.Equal(currentPrice) { // This implies PriceAtTick(tick) == currentPrice
		tick += int32(tickSpacing)
	}
	return tick, true
}
