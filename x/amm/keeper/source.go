package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.OrderSource = OrderSource{}
var threshold = sdk.NewDecWithPrec(1, 16) // XXX

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
	accQty := utils.ZeroInt
	accQuote := utils.ZeroDec
	numPriceLevels := 0
	k.IteratePoolOrders(ctx, pool, opts.IsBuy, func(price sdk.Dec, qty sdk.Int) (stop bool) {
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

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, results []*exchangetypes.MemOrder) error {
	reserveAddr := pool.MustGetReserveAddress()
	poolState := k.MustGetPoolState(ctx, pool.Id)
	accruedRewards := sdk.NewCoins()

	// TODO: check if results are sorted?
	isBuy := results[0].IsBuy
	firstOrderTick := exchangetypes.TickAtPrice(results[0].Price)
	var targetTick int32
	foundTargetTick := false
	if isBuy {
		k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, true, func(tick int32, tickInfo types.TickInfo) (stop bool) {
			if tick <= firstOrderTick {
				targetTick = tick
				foundTargetTick = true
				return true
			}
			netLiquidity := k.crossTick(ctx, pool.Id, tick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
			poolState.CurrentTick = tick
			poolState.CurrentPrice = exchangetypes.PriceAtTick(tick)
			return false
		})
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
			if tick >= firstOrderTick {
				targetTick = tick
				foundTargetTick = true
				return true
			}
			netLiquidity := k.crossTick(ctx, pool.Id, tick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
			poolState.CurrentTick = tick
			poolState.CurrentPrice = exchangetypes.PriceAtTick(tick)
			return false
		})
	}
	if !foundTargetTick { // sanity check
		panic("target tick not found")
	}

	max := false
	extraAmt0, extraAmt1 := utils.ZeroDec, utils.ZeroDec

	accrueFees := func() {
		extraAmt0Int := extraAmt0.TruncateInt()
		extraAmt1Int := extraAmt1.TruncateInt()
		fees := sdk.Coins{}
		if extraAmt0Int.IsPositive() {
			fees = fees.Add(sdk.NewCoin(pool.Denom0, extraAmt0Int))
		}
		if extraAmt1Int.IsPositive() {
			fees = fees.Add(sdk.NewCoin(pool.Denom1, extraAmt1Int))
		}
		if poolState.CurrentLiquidity.IsPositive() && !fees.IsZero() {
			accruedRewards = accruedRewards.Add(fees...)
			feeGrowth := sdk.NewDecCoinsFromCoins(fees...).
				MulDecTruncate(types.DecMulFactor).
				QuoDecTruncate(poolState.CurrentLiquidity.ToDec())
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(feeGrowth...)
		}
		extraAmt0 = utils.ZeroDec
		extraAmt1 = utils.ZeroDec
	}
	for i, order := range results {
		orderTick := exchangetypes.TickAtPrice(order.Price)

		if isBuy && max && poolState.CurrentTick == targetTick {
			accrueFees()
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
			foundTargetTick = false
			k.IterateTickInfosBelow(ctx, pool.Id, targetTick, false, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick <= orderTick {
					targetTick = tick
					foundTargetTick = true
					return true
				}
				accrueFees()
				netLiquidity = k.crossTick(ctx, pool.Id, tick, poolState)
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
				poolState.CurrentTick = tick
				poolState.CurrentPrice = exchangetypes.PriceAtTick(tick)
				return false
			})
			if !foundTargetTick { // sanity check
				panic("target tick not found")
			}
		} else if !isBuy && max && poolState.CurrentPrice.Equal(exchangetypes.PriceAtTick(targetTick)) {
			foundTargetTick = false
			k.IterateTickInfosAbove(ctx, pool.Id, targetTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick >= orderTick {
					targetTick = tick
					foundTargetTick = true
					return true
				}
				accrueFees()
				netLiquidity := k.crossTick(ctx, pool.Id, tick, poolState)
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
				poolState.CurrentTick = tick
				poolState.CurrentPrice = exchangetypes.PriceAtTick(tick)
				return false
			})
			if !foundTargetTick { // sanity check
				panic("target tick not found")
			}
		}

		res := order.Result()

		currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
		var nextSqrtPrice, nextPrice sdk.Dec
		max = false
		if i < len(results)-1 {
			nextSqrtPrice = utils.DecApproxSqrt(order.Price)
			nextPrice = order.Price
			max = true
		} else { // Partially executed
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, res.Paid, isBuy)
			nextPrice = nextSqrtPrice.Power(2)
		}

		var expectedAmtIn sdk.Dec
		if isBuy {
			expectedAmtIn = types.Amount0DeltaRoundingDec(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedAmtIn = types.Amount1DeltaDec(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity)
		}
		amtInDiff := res.Received.ToDec().Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			if isBuy {
				extraAmt0 = extraAmt0.Add(amtInDiff)
			} else {
				extraAmt1 = extraAmt1.Add(amtInDiff)
			}
		} else if amtInDiff.IsNegative() { // sanity check
			panic(fmt.Sprintf("amtInDiff is negative: %s", amtInDiff))
		}

		if res.FeeReceived.IsPositive() { // extra fees
			fee := res.FeeReceived.ToDec()
			if isBuy {
				extraAmt1 = extraAmt1.Add(fee)
			} else {
				extraAmt0 = extraAmt0.Add(fee)
			}
		}

		nextTick := exchangetypes.TickAtPrice(nextPrice)
		if !isBuy && max && nextTick == targetTick {
			accrueFees()
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
		}
		poolState.CurrentPrice = nextPrice
		poolState.CurrentTick = nextTick
	}
	accrueFees()
	k.SetPoolState(ctx, pool.Id, poolState)

	if accruedRewards.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(
			ctx, reserveAddr, pool.MustGetRewardsPoolAddress(), accruedRewards); err != nil {
			return err
		}
	}
	return nil
}
