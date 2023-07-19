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

func (k OrderSource) GenerateOrders(
	ctx sdk.Context, market exchangetypes.Market,
	createOrder exchangetypes.CreateOrderFunc,
	opts exchangetypes.GenerateOrdersOptions) {
	pool, found := k.GetPoolByMarket(ctx, market.Id)
	if !found {
		return // no pool found
	}

	reserveAddr := pool.MustGetReserveAddress()
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	numPriceLevels := 0
	k.IteratePoolOrders(ctx, pool, opts.IsBuy, func(price, qty sdk.Dec) (stop bool) {
		if opts.PriceLimit != nil &&
			((opts.IsBuy && price.LT(*opts.PriceLimit)) ||
				(!opts.IsBuy && price.GT(*opts.PriceLimit))) {
			return true
		}
		if opts.QuantityLimit != nil && !opts.QuantityLimit.Sub(accQty).IsPositive() {
			return true
		}
		if opts.QuoteLimit != nil && !opts.QuoteLimit.Sub(accQuote).IsPositive() {
			return true
		}
		if err := createOrder(reserveAddr, price, qty); err != nil {
			panic(err)
		}
		accQty = accQty.Add(qty)
		accQuote = accQuote.Add(exchangetypes.QuoteAmount(!opts.IsBuy, price, qty))
		if opts.MaxNumPriceLevels > 0 {
			numPriceLevels++
			if numPriceLevels >= opts.MaxNumPriceLevels {
				return true
			}
		}
		return false
	})
}

func (k OrderSource) AfterOrdersExecuted(ctx sdk.Context, _ exchangetypes.Market, results []exchangetypes.TempOrder) {
	orderers, m := exchangetypes.GroupTempOrderResultsByOrderer(results)
	for _, orderer := range orderers {
		ordererAddr := sdk.MustAccAddressFromBech32(orderer)
		pool := k.MustGetPoolByReserveAddress(ctx, ordererAddr)
		k.AfterPoolOrdersExecuted(ctx, pool, m[orderer])
	}
}

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, results []exchangetypes.TempOrder) {
	reserveAddr := pool.MustGetReserveAddress()
	poolState := k.MustGetPoolState(ctx, pool.Id)
	accruedRewards := sdk.NewCoins()

	// TODO: check if results are sorted?
	isBuy := results[0].Order.IsBuy
	firstOrderTick := exchangetypes.TickAtPrice(results[0].Order.Price)
	var targetTick int32
	foundTargetTick := false
	if isBuy {
		k.IterateTickInfosBelowInclusive(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
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
	for _, result := range results {
		orderTick := exchangetypes.TickAtPrice(result.Order.Price)

		if isBuy && max && poolState.CurrentTick == targetTick {
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
			foundTargetTick = false
			k.IterateTickInfosBelow(ctx, pool.Id, targetTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick <= orderTick {
					targetTick = tick
					foundTargetTick = true
					return true
				}
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

		currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
		var nextSqrtPrice, nextPrice sdk.Dec
		max = false
		// TODO: compare open quantity with zero directly (need changes in x/exchange)
		if result.Order.ExecutableQuantity(result.Order.Price).TruncateInt().IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(result.Order.Price)
			nextPrice = result.Order.Price
			max = true
		} else { // Partially executed
			// TODO: fix nextSqrtPrice?
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, result.Paid.Amount, result.Order.IsBuy)
			nextPrice = nextSqrtPrice.Power(2)
		}

		var expectedAmtIn sdk.Dec
		if result.Order.IsBuy {
			expectedAmtIn = types.Amount0DeltaRoundingDec(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedAmtIn = types.Amount1DeltaDec(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity)
		}
		denomIn := pool.DenomIn(isBuy)
		amtInDiff := result.Received.AmountOf(denomIn).Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			fee, _ := sdk.NewDecCoinFromDec(denomIn, amtInDiff).TruncateDecimal()
			accruedRewards = accruedRewards.Add(fee)
			feeGrowth := sdk.NewDecCoinFromDec(
				fee.Denom, fee.Amount.ToDec().
					MulTruncate(types.DecMulFactor).
					QuoTruncate(poolState.CurrentLiquidity.ToDec()))
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(feeGrowth)
		} else if amtInDiff.IsNegative() { // sanity check
			panic(amtInDiff)
		}

		// TODO: simplify code
		if len(result.Received) > 1 { // extra fees
			denomOut := pool.DenomOut(isBuy)
			fee, _ := sdk.NewDecCoinFromDec(denomOut, result.Received.AmountOf(denomOut)).TruncateDecimal()
			accruedRewards = accruedRewards.Add(fee)
			feeGrowth := sdk.NewDecCoinFromDec(
				fee.Denom, fee.Amount.ToDec().
					MulTruncate(types.DecMulFactor).
					QuoTruncate(poolState.CurrentLiquidity.ToDec()))
			poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.Add(feeGrowth)
		}

		nextTick := exchangetypes.TickAtPrice(nextPrice)
		if !isBuy && max && nextTick == targetTick {
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
		}
		poolState.CurrentPrice = nextPrice
		poolState.CurrentTick = nextTick
	}
	k.SetPoolState(ctx, pool.Id, poolState)

	// TODO: use separate addresses for different pools
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, reserveAddr, types.ModuleName, accruedRewards); err != nil {
		panic(err)
	}
}
