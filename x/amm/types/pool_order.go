package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func NextOrderTick(
	isBuy bool, liquidity sdk.Int, currentSqrtPrice cremath.BigDec, minOrderQty, minOrderQuote sdk.Int, tickSpacing uint32) (tick int32, valid bool) {
	// Q = minimum order qty
	// q = minimum order quote
	// P_c = current price
	// P_o = order price
	// L = liquidity
	liquidityBigDec := cremath.NewBigDecFromInt(liquidity)
	minOrderQtyBigDec := cremath.NewBigDecFromInt(minOrderQty)
	minOrderQuoteBigDec := cremath.NewBigDecFromInt(minOrderQuote)
	currentTick := exchangetypes.TickAtPrice(currentSqrtPrice.Power(2).Dec())
	if isBuy {
		// 1. Check min order qty
		// L*(sqrt(P_c) - sqrt(P_o)) / (sqrt(P_c)*sqrt(P_o)) >= Q
		// -> sqrt(P_o) <= L*sqrt(P_c) / (L + Q*sqrt(P_c))
		orderSqrtPrice := currentSqrtPrice.Mul(liquidityBigDec).
			QuoTruncateMut(liquidityBigDec.Add(minOrderQtyBigDec.Mul(currentSqrtPrice)))
		if !orderSqrtPrice.IsPositive() {
			return 0, false
		}
		// 2. Check min order quote
		// L*(sqrt(P_c) - sqrt(P_o)) / (sqrt(P_c)*sqrt(P_o)) * P_o >= q
		// -> L*P_o - L*sqrt(P_c)*sqrt(P_o) + q*sqrt(P_c) <= 0
		// -> sqrt(P_o) <= (L*sqrt(P_c) + sqrt(L^2*P_c - 4qL*sqrt(P_c))) / 2L
		// NOTE: if intermediate is negative, it indicates that there's no solution.
		intermediate := liquidityBigDec.Power(2).MulMut(currentSqrtPrice.Power(2)).
			SubMut(minOrderQuoteBigDec.Mul(liquidityBigDec).MulMut(currentSqrtPrice).MulInt64Mut(4))
		if intermediate.IsNegative() {
			return 0, false
		}
		orderSqrtPrice2 := liquidityBigDec.Mul(currentSqrtPrice).AddMut(intermediate.SqrtMut()).
			QuoTruncateMut(liquidityBigDec.MulInt64(2))
		if !orderSqrtPrice2.IsPositive() {
			return 0, false
		}
		orderSqrtPrice = cremath.MinBigDec(orderSqrtPrice, orderSqrtPrice2)
		if orderSqrtPrice.GT(currentSqrtPrice) {
			return 0, false
		}
		tick = AdjustPriceToTickSpacing(orderSqrtPrice.Power(2).Dec(), tickSpacing, false)
		if tick == currentTick {
			tick -= int32(tickSpacing)
		}
		return tick, true
	}
	// 1. Check min order qty
	// L*(sqrt(P_o) - sqrt(P_c)) / (sqrt(P_o)*sqrt(P_c)) >= Q
	// -> sqrt(P_o) >= L*sqrt(P_c) / (L - Q*sqrt(P_c))
	// NOTE: if denominator is not positive, it indicates that there's no solution.
	denominator := liquidityBigDec.Sub(minOrderQtyBigDec.Mul(currentSqrtPrice))
	if !denominator.IsPositive() {
		return 0, false
	}
	orderSqrtPrice := currentSqrtPrice.Mul(liquidityBigDec).QuoRoundUpMut(denominator)
	if !orderSqrtPrice.IsPositive() {
		return 0, false
	}
	// 2. Check min order quote
	// L*(sqrt(P_o) - sqrt(P_c)) / (sqrt(P_o)*sqrt(P_c)) * P_o >= q
	// -> L*P_o - L*sqrt(P_c)*sqrt(P_o) - q*sqrt(P_c) <= 0
	// -> sqrt(P_o) >= (L*sqrt(P_c) + sqrt(L^2*P_c + 4qL*sqrt(P_c))) / 2L
	intermediate := liquidityBigDec.Power(2).MulMut(currentSqrtPrice.Power(2)).
		AddMut(minOrderQuoteBigDec.Mul(liquidityBigDec).MulMut(currentSqrtPrice).MulInt64Mut(4))
	orderSqrtPrice2 := liquidityBigDec.Mul(currentSqrtPrice).AddMut(intermediate.SqrtMut()).
		QuoRoundUpMut(liquidityBigDec.MulInt64(2))
	if !orderSqrtPrice2.IsPositive() {
		return 0, false
	}
	orderSqrtPrice = cremath.MaxBigDec(orderSqrtPrice, orderSqrtPrice2)
	if orderSqrtPrice.LT(currentSqrtPrice) {
		return 0, false
	}
	tick = AdjustPriceToTickSpacing(orderSqrtPrice.Power(2).Dec(), tickSpacing, true)
	if tick == currentTick {
		tick += int32(tickSpacing)
	}
	return tick, true
}
