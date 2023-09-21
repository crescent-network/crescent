package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.OrderSource = OrderSource{}

type OrderSource struct {
	Keeper
}

func NewOrderSource(k Keeper) OrderSource {
	return OrderSource{k}
}

func (k OrderSource) Name() string {
	return types.ModuleName
}

func (k OrderSource) ConstructMemOrderBookSide(
	ctx sdk.Context, market exchangetypes.Market,
	createOrder exchangetypes.CreateOrderFunc,
	opts exchangetypes.MemOrderBookSideOptions) error {
	pool, found := k.GetPoolByMarket(ctx, market.Id)
	if !found {
		return nil // no pool found
	}
	maxPriceRatio := k.exchangeKeeper.GetMaxOrderPriceRatio(ctx)
	poolState := k.MustGetPoolState(ctx, pool.Id)
	minPrice, maxPrice := exchangetypes.OrderPriceLimit(poolState.CurrentPrice, maxPriceRatio)

	reserveAddr := pool.MustGetReserveAddress()
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	numPriceLevels := 0
	k.IteratePoolOrders(ctx, pool, opts.IsBuy, func(price, qty, openQty sdk.Dec) (stop bool) {
		if (opts.IsBuy && price.LT(minPrice)) ||
			(!opts.IsBuy && price.GT(maxPrice)) {
			return true
		}
		if opts.ReachedLimit(price, accQty, accQuote, numPriceLevels) {
			return true
		}
		createOrder(reserveAddr, price, qty, openQty)
		accQty = accQty.Add(qty)
		accQuote = accQuote.Add(exchangetypes.QuoteAmount(!opts.IsBuy, price, qty))
		numPriceLevels++
		return false
	})
	return nil
}

func (k OrderSource) AfterOrdersExecuted(ctx sdk.Context, _ exchangetypes.Market, ordererAddr sdk.AccAddress, results []*exchangetypes.MemOrder) error {
	pool := k.MustGetPoolByReserveAddress(ctx, ordererAddr)
	return k.AfterPoolOrdersExecuted(ctx, pool, results)
}

func (k Keeper) IteratePoolOrders(ctx sdk.Context, pool types.Pool, isBuy bool, cb func(price, qty, openQty sdk.Dec) (stop bool)) {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).
		AmountOf(pool.DenomOut(isBuy)).ToDec()
	orderLiquidity := poolState.CurrentLiquidity
	currentPrice := poolState.CurrentPrice

	iterCb := func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if orderLiquidity.IsPositive() {
			for {
				if !reserveBalance.IsPositive() {
					return true
				}
				var (
					orderTick int32
					valid     bool
				)
				if isBuy {
					orderTick, valid = NextOrderTick(
						true, orderLiquidity, currentPrice, pool.MinOrderQuantity, pool.MinOrderQuote, pool.TickSpacing)
					if !valid || orderTick < tick {
						orderTick = tick
					}
				} else {
					orderTick, valid = NextOrderTick(
						false, orderLiquidity, currentPrice, pool.MinOrderQuantity, pool.MinOrderQuote, pool.TickSpacing)
					if !valid || orderTick > tick {
						orderTick = tick
					}
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
				currentSqrtPrice := utils.DecApproxSqrt(currentPrice)
				var qty, openQty sdk.Dec
				if isBuy {
					qty = types.Amount1DeltaDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity).QuoTruncate(orderPrice)
					openQty = sdk.MinDec(reserveBalance.QuoTruncate(orderPrice), qty)
				} else {
					qty = types.Amount0DeltaRoundingDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false)
					openQty = sdk.MinDec(reserveBalance, qty)
				}
				if openQty.IsPositive() && (orderTick == tick || (openQty.GTE(pool.MinOrderQuantity))) {
					if cb(orderPrice, qty, openQty) {
						return true
					}
					reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, orderPrice, qty))
					currentPrice = orderPrice
				} else { // No more possible order price
					break
				}
				if orderTick == tick {
					break
				}
			}
		} else {
			currentPrice = exchangetypes.PriceAtTick(tick)
		}
		if isBuy {
			orderLiquidity = orderLiquidity.Sub(tickInfo.NetLiquidity)
		} else {
			orderLiquidity = orderLiquidity.Add(tickInfo.NetLiquidity)
		}
		return false
	}
	if isBuy {
		k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, true, iterCb)
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, iterCb)
	}
}

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, results []*exchangetypes.MemOrder) error {
	isBuy := results[0].IsBuy()

	poolState := k.MustGetPoolState(ctx, pool.Id)
	currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
	rewards := sdk.Coins{}

	for len(results) > 0 {
		nextTick, found := k.nextTick(ctx, pool.Id, poolState.CurrentTick, isBuy)
		if !found {
			break
		}
		targetPrice := exchangetypes.PriceAtTick(nextTick)
		targetSqrtPrice := utils.DecApproxSqrt(targetPrice)

		amtOut := utils.ZeroDec
		amtIn := utils.ZeroDec
		fee := utils.ZeroDec
		for len(results) > 0 && ((isBuy && results[0].Price().GTE(targetPrice)) ||
			(!isBuy && results[0].Price().LTE(targetPrice))) {
			result := results[0]
			amtOut = amtOut.Add(result.PaidWithoutFee())
			amtIn = amtIn.Add(result.Received())
			if result.Fee().IsNegative() {
				fee = fee.Add(result.Fee().Neg())
			}
			results = results[1:]
		}

		var expectedAmtOutToTargetSqrtPrice sdk.Dec
		if isBuy {
			expectedAmtOutToTargetSqrtPrice = types.Amount1DeltaDec(
				targetSqrtPrice, currentSqrtPrice, poolState.CurrentLiquidity)
		} else {
			expectedAmtOutToTargetSqrtPrice = types.Amount0DeltaRoundingDec(
				currentSqrtPrice, targetSqrtPrice, poolState.CurrentLiquidity, false)
		}
		var nextSqrtPrice sdk.Dec
		if amtOut.GTE(expectedAmtOutToTargetSqrtPrice) {
			nextSqrtPrice = targetSqrtPrice
		} else {
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, amtOut, isBuy)
		}

		// Calculate expected amount in.
		var expectedAmtIn sdk.Dec
		if isBuy {
			expectedAmtIn = types.Amount0DeltaRoundingDec(
				nextSqrtPrice, currentSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedAmtIn = types.Amount1DeltaDec(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity)
		}
		if amtIn.GT(expectedAmtIn) {
			fee := amtIn.Sub(expectedAmtIn).TruncateInt()
			feeCoin := sdk.NewCoin(pool.DenomIn(isBuy), fee)
			if !feeCoin.IsZero() {
				rewards = rewards.Add(feeCoin)
				poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
					sdk.NewDecCoinsFromCoins(feeCoin).
						MulDecTruncate(types.DecMulFactor).
						QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
			}
		} else if amtIn.LT(expectedAmtIn) {
			panic(amtIn.Sub(expectedAmtIn))
		}

		if fee.IsPositive() {
			feeCoin := sdk.NewCoin(pool.DenomOut(isBuy), fee.TruncateInt())
			if !feeCoin.IsZero() {
				rewards = rewards.Add(feeCoin)
				poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
					sdk.NewDecCoinsFromCoins(feeCoin).
						MulDecTruncate(types.DecMulFactor).
						QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
			}
		}

		currentSqrtPrice = nextSqrtPrice

		if currentSqrtPrice.Equal(targetSqrtPrice) {
			netLiquidity := k.crossTick(ctx, pool.Id, nextTick, poolState)
			if isBuy {
				netLiquidity = netLiquidity.Neg()
			}
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
			if isBuy {
				poolState.CurrentTick = nextTick - 1
			} else {
				poolState.CurrentTick = nextTick
			}
		} else {
			poolState.CurrentTick = exchangetypes.TickAtPrice(currentSqrtPrice.Power(2))
		}
	}
	poolState.CurrentPrice = currentSqrtPrice.Power(2)
	k.SetPoolState(ctx, pool.Id, poolState)
	if rewards.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(
			ctx, pool.MustGetReserveAddress(), pool.MustGetRewardsPoolAddress(), rewards); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) nextTick(ctx sdk.Context, poolId uint64, currentTick int32, lte bool) (nextTick int32, found bool) {
	if lte {
		k.IterateTickInfosBelow(ctx, poolId, currentTick, true, func(tick int32, tickInfo types.TickInfo) (stop bool) {
			nextTick = tick
			found = true
			return true
		})
	} else {
		k.IterateTickInfosAbove(ctx, poolId, currentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
			nextTick = tick
			found = true
			return true
		})
	}
	return
}

func NextOrderTick(
	isBuy bool, liquidity sdk.Int, currentPrice, minOrderQty, minOrderQuote sdk.Dec, tickSpacing uint32) (tick int32, valid bool) {
	currentSqrtPrice := utils.DecApproxSqrt(currentPrice)
	liquidityDec := liquidity.ToDec()
	if isBuy {
		// 1. Check min order qty
		// L^2 + 4 * MinQty * L * sqrt(P_current)
		intermediate := liquidityDec.Power(2).Add(
			minOrderQty.Mul(liquidityDec).MulTruncate(currentSqrtPrice).MulInt64(4))
		orderSqrtPrice := utils.DecApproxSqrt(intermediate).Sub(liquidityDec).QuoTruncate(minOrderQty.MulInt64(2))
		if !orderSqrtPrice.IsPositive() {
			return 0, false
		}
		// 2. Check min order quote
		orderSqrtPrice2 := currentSqrtPrice.Mul(liquidityDec).Sub(minOrderQuote).QuoTruncate(liquidityDec)
		if !orderSqrtPrice2.IsPositive() {
			return 0, false
		}
		orderPrice := sdk.MinDec(orderSqrtPrice, orderSqrtPrice2).Power(2)
		if orderPrice.GT(currentPrice) {
			return 0, false
		}
		tick = types.AdjustPriceToTickSpacing(orderPrice, tickSpacing, false)
		if orderPrice.Equal(currentPrice) { // This implies PriceAtTick(tick) == currentPrice
			tick -= int32(tickSpacing)
		}
		return tick, true
	}
	// 1. Check min order qty
	orderSqrtPrice := currentSqrtPrice.Mul(liquidityDec).
		QuoRoundUp(liquidityDec.Sub(minOrderQty.Mul(currentSqrtPrice)))
	if !orderSqrtPrice.IsPositive() {
		return 0, false
	}
	// 2. Check min order quote
	orderSqrtPrice2 := minOrderQuote.Mul(currentSqrtPrice).QuoRoundUp(liquidityDec).Add(currentSqrtPrice)
	if !orderSqrtPrice2.IsPositive() {
		return 0, false
	}
	orderPrice := sdk.MaxDec(orderSqrtPrice, orderSqrtPrice2).Power(2)
	if orderPrice.LT(currentPrice) {
		return 0, false
	}
	tick = types.AdjustPriceToTickSpacing(orderPrice, tickSpacing, true)
	if orderPrice.Equal(currentPrice) { // This implies PriceAtTick(tick) == currentPrice
		tick += int32(tickSpacing)
	}
	return tick, true
}
