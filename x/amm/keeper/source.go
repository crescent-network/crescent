package keeper

import (
	"fmt"

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
	minPrice, maxPrice := exchangetypes.OrderPriceLimit(
		poolState.CurrentSqrtPrice.Power(2), maxPriceRatio)

	reserveAddr := pool.MustGetReserveAddress()
	accQty := utils.ZeroInt
	accQuote := utils.ZeroDec
	numPriceLevels := 0
	k.IteratePoolOrders(ctx, market, pool, opts.IsBuy, func(price sdk.Dec, qty sdk.Int) (stop bool) {
		if (opts.IsBuy && price.LT(minPrice)) ||
			(!opts.IsBuy && price.GT(maxPrice)) {
			return true
		}
		if opts.ReachedLimit(price, accQty, accQuote, numPriceLevels) {
			return true
		}
		createOrder(reserveAddr, price, qty)
		accQty = accQty.Add(qty)
		accQuote = accQuote.Add(price.MulInt(qty))
		numPriceLevels++
		return false
	})
	return nil
}

func (k OrderSource) AfterOrdersExecuted(ctx sdk.Context, _ exchangetypes.Market, ordererAddr sdk.AccAddress, results []*exchangetypes.MemOrder) error {
	pool := k.MustGetPoolByReserveAddress(ctx, ordererAddr)
	return k.AfterPoolOrdersExecuted(ctx, pool, results)
}

func (k Keeper) IteratePoolOrders(ctx sdk.Context, market exchangetypes.Market, pool types.Pool, isBuy bool, cb func(price sdk.Dec, qty sdk.Int) (stop bool)) {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).
		AmountOf(pool.DenomOut(isBuy))
	orderLiquidity := poolState.CurrentLiquidity
	currentSqrtPrice := poolState.CurrentSqrtPrice

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
						true, orderLiquidity, currentSqrtPrice, market.OrderQuantityLimits.Min, market.OrderQuoteLimits.Min, pool.TickSpacing)
					if !valid || orderTick < tick {
						orderTick = tick
					}
				} else {
					orderTick, valid = NextOrderTick(
						false, orderLiquidity, currentSqrtPrice, market.OrderQuantityLimits.Min, market.OrderQuoteLimits.Min, pool.TickSpacing)
					if !valid || orderTick > tick {
						orderTick = tick
					}
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
				var qty sdk.Int
				if isBuy {
					qty = utils.MinInt(
						reserveBalance.ToDec().QuoTruncate(orderPrice).TruncateInt(),
						types.Amount1DeltaDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity).QuoTruncate(orderPrice).TruncateInt())
				} else {
					qty = utils.MinInt(
						reserveBalance,
						types.Amount0DeltaRoundingDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false).TruncateInt())
				}
				if qty.IsPositive() && (orderTick == tick || (qty.GTE(market.OrderQuantityLimits.Min))) {
					if cb(orderPrice, qty) {
						return true
					}
					reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, orderPrice, qty))
					currentSqrtPrice = orderSqrtPrice
				} else { // No more possible order price
					break
				}
				if orderTick == tick {
					break
				}
			}
		} else {
			currentSqrtPrice = types.SqrtPriceAtTick(tick)
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

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, orders []*exchangetypes.MemOrder) error {
	isBuy := orders[0].IsBuy

	poolState := k.MustGetPoolState(ctx, pool.Id)
	rewards := sdk.Coins{}

	queue := newOrderResultQueue(matchResultsFromMemOrders(orders))

	for queue.remainingPaid.IsPositive() {
		nextTick, found := k.nextTick(ctx, pool.Id, poolState.CurrentTick, isBuy)
		if !found {
			break
		}
		targetPrice := exchangetypes.PriceAtTick(nextTick)
		targetSqrtPrice := utils.DecApproxSqrt(targetPrice)

		var amtOut sdk.Int
		if isBuy {
			amtOut = types.Amount1Delta(
				targetSqrtPrice, poolState.CurrentSqrtPrice, poolState.CurrentLiquidity)
		} else {
			amtOut = types.Amount0DeltaRounding(
				poolState.CurrentSqrtPrice, targetSqrtPrice, poolState.CurrentLiquidity, false)
		}

		var nextSqrtPrice sdk.Dec
		if queue.remainingPaid.GTE(amtOut) {
			nextSqrtPrice = targetSqrtPrice
		} else {
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, queue.remainingPaid, isBuy)
		}

		paid := utils.MinInt(amtOut, queue.remainingPaid)
		received, feeReceived := queue.pull(paid)

		// Calculate CPM adjustment fee.
		var expectedReceived sdk.Int
		if isBuy {
			expectedReceived = types.Amount0DeltaRounding(
				nextSqrtPrice, poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedReceived = types.Amount1Delta(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity)
		}

		if received.GT(expectedReceived) {
			extraReceived := received.Sub(expectedReceived)
			feeCoin := sdk.NewCoin(pool.DenomIn(isBuy), extraReceived)
			if !feeCoin.IsZero() {
				rewards = rewards.Add(feeCoin)
				poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
					sdk.NewDecCoinsFromCoins(feeCoin).
						MulDecTruncate(types.DecMulFactor).
						QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
			}
		} else if received.LT(expectedReceived) {
			// XXX
		}

		if feeReceived.IsPositive() {
			feeCoin := sdk.NewCoin(pool.DenomOut(isBuy), feeReceived)
			rewards = rewards.Add(feeCoin)
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
				sdk.NewDecCoinsFromCoins(feeCoin).
					MulDecTruncate(types.DecMulFactor).
					QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
		}

		poolState.CurrentSqrtPrice = nextSqrtPrice
		if poolState.CurrentSqrtPrice.Equal(targetSqrtPrice) {
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
			poolState.CurrentTick = exchangetypes.TickAtPrice(poolState.CurrentSqrtPrice.Power(2))
		}
	}

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
	isBuy bool, liquidity sdk.Int, currentSqrtPrice sdk.Dec, minOrderQty, minOrderQuote sdk.Int, tickSpacing uint32) (tick int32, valid bool) {
	liquidityDec := liquidity.ToDec()
	minOrderQtyDec := minOrderQty.ToDec()
	minOrderQuoteDec := minOrderQuote.ToDec()
	currentTick := exchangetypes.TickAtPrice(currentSqrtPrice.Power(2))
	if isBuy {
		// 1. Check min order qty
		// L^2 + 4 * MinQty * L * sqrt(P_current)
		intermediate := liquidityDec.Power(2).Add(
			minOrderQtyDec.Mul(liquidityDec).MulTruncate(currentSqrtPrice).MulInt64(4))
		orderSqrtPrice := utils.DecApproxSqrt(intermediate).Sub(liquidityDec).QuoTruncate(minOrderQtyDec.MulInt64(2))
		if !orderSqrtPrice.IsPositive() {
			return 0, false
		}
		// 2. Check min order quote
		orderSqrtPrice2 := currentSqrtPrice.Mul(liquidityDec).Sub(minOrderQuoteDec).QuoTruncate(liquidityDec)
		if !orderSqrtPrice2.IsPositive() {
			return 0, false
		}
		orderSqrtPrice = sdk.MinDec(orderSqrtPrice, orderSqrtPrice2)
		if orderSqrtPrice.GT(currentSqrtPrice) {
			return 0, false
		}
		tick = types.AdjustPriceToTickSpacing(orderSqrtPrice.Power(2), tickSpacing, false)
		if tick == currentTick {
			tick -= int32(tickSpacing)
		}
		return tick, true
	}
	// 1. Check min order qty
	orderSqrtPrice := currentSqrtPrice.Mul(liquidityDec).
		QuoRoundUp(liquidityDec.Sub(minOrderQtyDec.Mul(currentSqrtPrice)))
	if !orderSqrtPrice.IsPositive() {
		return 0, false
	}
	// 2. Check min order quote
	orderSqrtPrice2 := minOrderQuoteDec.Mul(currentSqrtPrice).QuoRoundUp(liquidityDec).Add(currentSqrtPrice)
	if !orderSqrtPrice2.IsPositive() {
		return 0, false
	}
	orderSqrtPrice = sdk.MaxDec(orderSqrtPrice, orderSqrtPrice2)
	if orderSqrtPrice.LT(currentSqrtPrice) {
		return 0, false
	}
	tick = types.AdjustPriceToTickSpacing(orderSqrtPrice.Power(2), tickSpacing, true)
	if tick == currentTick {
		tick += int32(tickSpacing)
	}
	return tick, true
}

type orderResultQueue struct {
	results                    []exchangetypes.MatchResult
	remainingPaid, partialPaid sdk.Int
}

func matchResultsFromMemOrders(orders []*exchangetypes.MemOrder) []exchangetypes.MatchResult {
	results := make([]exchangetypes.MatchResult, 0, len(orders))
	for _, order := range orders {
		results = append(results, order.Result())
	}
	return results
}

func newOrderResultQueue(results []exchangetypes.MatchResult) *orderResultQueue {
	remainingPaid := utils.ZeroInt
	for _, result := range results {
		remainingPaid = remainingPaid.Add(result.Paid)
	}
	return &orderResultQueue{
		results:       results,
		remainingPaid: remainingPaid,
		partialPaid:   utils.ZeroInt,
	}
}

func (queue *orderResultQueue) pull(paid sdk.Int) (received, feeReceived sdk.Int) {
	if paid.GT(queue.remainingPaid) { // sanity check
		panic(fmt.Sprintf("paid %s > remaining %s", paid, queue.remainingPaid))
	}
	// We don't need to include `len(results) > 0` in the loop condition
	// since we already checked paid <= remainingPaid.
	received = utils.ZeroInt
	feeReceived = utils.ZeroInt
	for paid.IsPositive() {
		result := queue.results[0]
		resultRemainingPaid := result.Paid.Sub(queue.partialPaid)
		p := sdk.MinInt(paid, resultRemainingPaid)

		ratio := p.ToDec().QuoTruncate(result.Paid.ToDec())
		received = received.Add(ratio.MulInt(result.Received).TruncateInt())
		feeReceived = feeReceived.Add(ratio.MulInt(result.FeeReceived).TruncateInt())

		if p.Equal(resultRemainingPaid) {
			queue.results = queue.results[1:]
			queue.partialPaid = utils.ZeroInt
		} else {
			queue.partialPaid = queue.partialPaid.Add(p)
		}
		paid = paid.Sub(p)
		queue.remainingPaid = queue.remainingPaid.Sub(p)
	}
	return
}
