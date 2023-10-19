package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
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
		poolState.CurrentSqrtPrice.Power(2).Dec(), maxPriceRatio)

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

func (k Keeper) IteratePoolOrders(
	ctx sdk.Context, market exchangetypes.Market, pool types.Pool,
	isBuy bool, cb func(price sdk.Dec, qty sdk.Int) (stop bool)) {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).
		AmountOf(pool.DenomOut(isBuy))
	orderLiquidity := poolState.CurrentLiquidity
	currentSqrtPrice := poolState.CurrentSqrtPrice

	// lte = true if isBuy = true
	k.IterateTickInfos(ctx, pool.Id, poolState.CurrentTick, isBuy, func(tick int32, tickInfo types.TickInfo) (stop bool) {
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
					orderTick, valid = types.NextOrderTick(
						true, orderLiquidity, currentSqrtPrice, market.OrderQuantityLimits.Min, market.OrderQuoteLimits.Min, pool.TickSpacing)
					if !valid || orderTick < tick {
						break
					}
				} else {
					orderTick, valid = types.NextOrderTick(
						false, orderLiquidity, currentSqrtPrice, market.OrderQuantityLimits.Min, market.OrderQuoteLimits.Min, pool.TickSpacing)
					if !valid || orderTick > tick {
						break
					}
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := cremath.NewBigDecFromDec(orderPrice).SqrtMut()
				var qty sdk.Int
				if isBuy {
					qty = utils.MinInt(
						reserveBalance.ToDec().QuoTruncate(orderPrice).TruncateInt(),
						types.Amount0DeltaRounding(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false))
				} else {
					qty = utils.MinInt(
						reserveBalance,
						types.Amount0DeltaRounding(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false))
				}
				if qty.IsPositive() {
					if cb(orderPrice, qty) {
						return true
					}
				}
				reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, orderPrice, qty))
				currentSqrtPrice = orderSqrtPrice
				if orderTick == tick {
					break
				}
			}
		}
		currentSqrtPrice = types.SqrtPriceAtTick(tick)
		if isBuy {
			orderLiquidity = orderLiquidity.Sub(tickInfo.NetLiquidity)
		} else {
			orderLiquidity = orderLiquidity.Add(tickInfo.NetLiquidity)
		}
		return false
	})
}

var twoBigDec = cremath.NewBigDec(2) // XXX

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, orders []*exchangetypes.MemOrder) error {
	isBuy := orders[0].IsBuy
	poolState := k.MustGetPoolState(ctx, pool.Id)
	rewards := sdk.Coins{}

	// Make a queue of MatchResult from MemOrders.
	results := make([]exchangetypes.MatchResult, 0, len(orders))
	qtyRemaining := cremath.ZeroBigDec()
	for _, order := range orders {
		result := order.Result()
		qtyRemaining = qtyRemaining.Add(cremath.NewBigDecFromInt(result.ExecutedQuantity))
		results = append(results, result)
	}

	fmt.Printf("AfterPoolOrdersExecuted\n")
	fmt.Printf("isBuy=%v currentSqrtPrice=%s qtyRemaining=%s\n",
		isBuy, poolState.CurrentSqrtPrice, qtyRemaining)
	for i, order := range orders {
		fmt.Printf("  | price=%s qty=%s executed=%s paid=%s(%s+%s) received=%s)\n",
			order.Price, order.Quantity, results[i].ExecutedQuantity,
			results[i].Paid.Add(results[i].FeeReceived), results[i].Paid,
			results[i].FeeReceived, results[i].Received)
	}

	// prevExecutedQty is the amount that were partially processed from
	// the foremost result in queue before.
	prevExecutedQty := cremath.ZeroBigDec()

	for qtyRemaining.IsPositive() {
		nextTick, found := k.nextTick(ctx, pool.Id, poolState.CurrentTick, isBuy)
		if !found { // sanity check
			// If the amount remaining is positive, then there must be the next tick.
			panic("next tick not found")
		}
		targetPrice := exchangetypes.PriceAtTick(nextTick)
		targetSqrtPrice := cremath.NewBigDecFromDec(targetPrice).SqrtMut()
		fmt.Printf("qtyRemaining=%s liquidity=%s currentSqrtPrice=%s targetSqrtPrice=%s\n",
			qtyRemaining, poolState.CurrentLiquidity, poolState.CurrentSqrtPrice, targetSqrtPrice)

		var (
			nextSqrtPrice cremath.BigDec
			qty           cremath.BigDec
		)
		if isBuy {
			expectedAmtIn := types.Amount0DeltaRoundingBigDec(
				poolState.CurrentSqrtPrice, targetSqrtPrice, poolState.CurrentLiquidity, true)
			if qtyRemaining.GTE(expectedAmtIn) {
				nextSqrtPrice = targetSqrtPrice
				qty = expectedAmtIn
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromInputBigDec(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, qtyRemaining, isBuy)
				qty = qtyRemaining
			}
			fmt.Printf("expectedAmtIn=%s nextSqrtPrice=%s\n", expectedAmtIn, nextSqrtPrice)
		} else {
			expectedAmtOut := types.Amount0DeltaRoundingBigDec(
				poolState.CurrentSqrtPrice, targetSqrtPrice, poolState.CurrentLiquidity, false)
			if qtyRemaining.GTE(expectedAmtOut) {
				nextSqrtPrice = targetSqrtPrice
				qty = expectedAmtOut
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromOutputBigDec(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, qtyRemaining, isBuy)
				qty = qtyRemaining
			}
			fmt.Printf("expectedAmtOut=%s nextSqrtPrice=%s\n", expectedAmtOut, nextSqrtPrice)
		}
		fmt.Printf("qty=%s\n", qty)

		// Calculate received and feeReceived based on amtOut
		executedQuote, feeReceived := cremath.ZeroBigDec(), cremath.ZeroBigDec()
		for remainingQty := qty; remainingQty.IsPositive(); {
			result := results[0] // The foremost result in the queue.

			remainingExecutedQty := cremath.NewBigDecFromInt(result.ExecutedQuantity).
				Sub(prevExecutedQty)
			executedQty := cremath.MinBigDec(remainingQty, remainingExecutedQty)

			if isBuy {
				executedQuote = executedQuote.Add(
					executedQty.MulInt(result.Paid.Add(result.FeeReceived)).
						QuoInt(result.ExecutedQuantity))
			} else {
				executedQuote = executedQuote.Add(executedQty.MulInt(result.Received).
					QuoInt(result.ExecutedQuantity))
			}
			feeReceived = feeReceived.Add(
				executedQty.MulInt(result.FeeReceived).QuoInt(result.ExecutedQuantity))

			if executedQty.Equal(remainingExecutedQty) {
				results = results[1:]
				prevExecutedQty = cremath.ZeroBigDec()
			} else {
				prevExecutedQty = prevExecutedQty.Add(executedQty)
			}
			remainingQty = remainingQty.Sub(executedQty)
		}
		fmt.Printf("executedQuote=%s feeReceived=%s\n", executedQuote, feeReceived)

		// Accrue CPM adjustment fee.
		var adjustmentFee cremath.BigDec
		if isBuy {
			expectedAmtOut := types.Amount1DeltaRoundingBigDec(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false).TruncateInt()
			adjustmentFee = cremath.NewBigDecFromInt(expectedAmtOut).Sub(executedQuote)
			fmt.Printf("[CPM adj] expectedAmtOut=%s adjustmentFee=%s\n",
				expectedAmtOut, adjustmentFee)
		} else {
			expectedAmtIn := types.Amount1DeltaRoundingBigDec(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true).Ceil().TruncateInt()
			adjustmentFee = executedQuote.Sub(cremath.NewBigDecFromInt(expectedAmtIn))
			fmt.Printf("[CPM adj] expectedAmtIn=%s adjustmentFee=%s\n",
				expectedAmtIn, adjustmentFee)
		}
		if adjustmentFee.GTE(twoBigDec) {
			feeCoin := sdk.NewCoin(pool.DenomIn(isBuy), adjustmentFee.Sub(twoBigDec).TruncateInt())
			rewards = rewards.Add(feeCoin)
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
				sdk.NewDecCoinsFromCoins(feeCoin).
					MulDecTruncate(types.DecMulFactor).
					QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
		} else {
			fmt.Printf("[CPM adj] diff=%s\n", adjustmentFee)
			if adjustmentFee.LT(cremath.NewBigDec(-1)) { // receivedDiff < -1
				panic(adjustmentFee)
			}
		}

		if feeReceived.GT(cremath.OneBigDec()) {
			feeCoin := sdk.NewCoin(pool.DenomOut(isBuy), feeReceived.TruncateInt())
			rewards = rewards.Add(feeCoin)
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(
				sdk.NewDecCoinsFromCoins(feeCoin).
					MulDecTruncate(types.DecMulFactor).
					QuoDecTruncate(poolState.CurrentLiquidity.ToDec())...)
		}

		qtyRemaining = qtyRemaining.Sub(qty)

		// Update current sqrt price and handle tick crossing.
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
			poolState.CurrentTick = exchangetypes.TickAtPrice(
				poolState.CurrentSqrtPrice.Power(2).Dec())
		}
	}
	if !qtyRemaining.IsZero() { // sanity check
		panic("qtyRemaining must be zero after matching")
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
	k.IterateTickInfos(ctx, poolId, currentTick, lte, func(tick int32, tickInfo types.TickInfo) (stop bool) {
		nextTick = tick
		found = true
		return true
	})
	return
}
