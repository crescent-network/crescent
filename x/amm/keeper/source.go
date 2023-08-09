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
	opts exchangetypes.MemOrderBookSideOptions) {
	pool, found := k.GetPoolByMarket(ctx, market.Id)
	if !found {
		return // no pool found
	}

	reserveAddr := pool.MustGetReserveAddress()
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	numPriceLevels := 0
	k.IteratePoolOrders(ctx, pool, opts.IsBuy, func(price, qty sdk.Dec) (stop bool) {
		if opts.ReachedLimit(price, accQty, accQuote, numPriceLevels) {
			return true
		}
		createOrder(reserveAddr, price, qty)
		accQty = accQty.Add(qty)
		accQuote = accQuote.Add(exchangetypes.QuoteAmount(!opts.IsBuy, price, qty))
		numPriceLevels++
		return false
	})
}

func (k OrderSource) AfterOrdersExecuted(ctx sdk.Context, _ exchangetypes.Market, ordererAddr sdk.AccAddress, results []*exchangetypes.MemOrder) {
	pool := k.MustGetPoolByReserveAddress(ctx, ordererAddr)
	k.AfterPoolOrdersExecuted(ctx, pool, results)
}

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, results []*exchangetypes.MemOrder) {
	reserveAddr := pool.MustGetReserveAddress()
	poolState := k.MustGetPoolState(ctx, pool.Id)
	accruedRewards := sdk.NewCoins()

	// TODO: check if results are sorted?
	isBuy := results[0].IsBuy()
	firstOrderTick := exchangetypes.TickAtPrice(results[0].Price())
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
	for i, result := range results {
		orderTick := exchangetypes.TickAtPrice(result.Price())

		if isBuy && max && poolState.CurrentTick == targetTick {
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
			foundTargetTick = false
			k.IterateTickInfosBelow(ctx, pool.Id, targetTick, false, func(tick int32, tickInfo types.TickInfo) (stop bool) {
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
		if i < len(results)-1 || result.ExecutableQuantity().LTE(utils.SmallestDec) {
			nextSqrtPrice = utils.DecApproxSqrt(result.Price())
			nextPrice = result.Price()
			max = true
		} else { // Partially executed
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, result.PaidWithoutFee(), isBuy)
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
		amtInDiff := result.Received().Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			denomIn := pool.DenomIn(isBuy)
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
		if result.Fee().IsNegative() { // extra fees
			denomOut := pool.DenomOut(isBuy)
			fee := sdk.NewCoin(denomOut, result.Fee().Neg().TruncateInt())
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

	if err := k.bankKeeper.SendCoins(
		ctx, reserveAddr, pool.MustGetRewardsPoolAddress(), accruedRewards); err != nil {
		panic(err)
	}
}
