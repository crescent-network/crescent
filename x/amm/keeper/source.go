package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.TemporaryOrderSource = TemporaryOrderSource{}

type TemporaryOrderSource struct {
	Keeper
}

func NewTemporaryOrderSource(k Keeper) TemporaryOrderSource {
	return TemporaryOrderSource{k}
}

func (k TemporaryOrderSource) Name() string {
	return types.ModuleName
}

func (k TemporaryOrderSource) GenerateOrders(
	ctx sdk.Context, market exchangetypes.Market,
	cb exchangetypes.TemporaryOrderCallback,
	opts exchangetypes.TemporaryOrderOptions) {
	// Select the first pool since there will be only one pool per market
	// TODO: use GetPoolByMarket instead of IteratePoolsByMarket
	var pool types.Pool
	k.IteratePoolsByMarket(ctx, market.Id, func(p types.Pool) (stop bool) {
		pool = p
		return true
	})
	if pool == (types.Pool{}) { //  TODO: use flag
		return // no pool found
	}

	reserveAddr := pool.MustGetReserveAddress()
	accQty := utils.ZeroInt
	accQuote := utils.ZeroInt
	k.IteratePoolOrders(ctx, pool, opts.IsBuy, func(price sdk.Dec, qty sdk.Int, liquidity sdk.Dec) (stop bool) {
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
		if err := cb(reserveAddr, price, qty); err != nil {
			panic(err)
		}
		accQty = accQty.Add(qty)
		accQuote = accQuote.Add(exchangetypes.QuoteAmount(!opts.IsBuy, price, qty))
		return false
	})
}

func (k TemporaryOrderSource) AfterOrdersExecuted(ctx sdk.Context, _ exchangetypes.Market, results []exchangetypes.TemporaryOrderResult) {
	// TODO: group results by orderer
	orderers, m := exchangetypes.GroupTemporaryOrderResultsByOrderer(results)
	for _, orderer := range orderers {
		ordererAddr := sdk.MustAccAddressFromBech32(orderer)
		pool, found := k.GetPoolByReserveAddress(ctx, ordererAddr)
		if !found {
			panic("pool not found")
		}
		k.AfterPoolOrdersExecuted(ctx, pool, m[orderer])
	}
}

func (k Keeper) AfterPoolOrdersExecuted(ctx sdk.Context, pool types.Pool, results []exchangetypes.TemporaryOrderResult) {
	// TODO: check if results are sorted?
	isBuy := results[0].Order.IsBuy

	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
	poolState := k.MustGetPoolState(ctx, pool.Id)
	accruedRewards := sdk.NewCoins()

	for _, result := range results {
		orderTick := exchangetypes.TickAtPrice(result.Order.Price, TickPrecision)
		if isBuy {
			k.IterateTickInfosBelowInclusive(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick <= orderTick {
					return true
				}
				netLiquidity := k.crossTick(ctx, pool.Id, tick, poolState.FeeGrowthGlobal0, poolState.FeeGrowthGlobal1)
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(netLiquidity)
				poolState.CurrentTick = tick
				poolState.CurrentPrice = exchangetypes.PriceAtTick(tick, TickPrecision)
				return false
			})
		} else {
			k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick >= orderTick {
					return true
				}
				netLiquidity := k.crossTick(ctx, pool.Id, tick, poolState.FeeGrowthGlobal0, poolState.FeeGrowthGlobal1)
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
				poolState.CurrentTick = tick
				poolState.CurrentPrice = exchangetypes.PriceAtTick(tick, TickPrecision)
				return false
			})
		}

		var targetTick int32
		if isBuy {
			k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				targetTick = tick
				return true
			})
		} else {
			k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				targetTick = tick
				return true
			})
		}
		targetPrice := exchangetypes.PriceAtTick(targetTick, TickPrecision)

		currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
		var nextSqrtPrice, nextPrice sdk.Dec
		max := false
		if result.Order.OpenQuantity.IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(result.Order.Price)
			nextPrice = result.Order.Price
			max = true
		} else { // Partially executed
			// TODO: fix nextSqrtPrice?
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, result.Paid.Amount, result.Order.IsBuy)
			nextPrice = nextSqrtPrice.Power(2)
		}

		var expectedAmtIn sdk.Int
		if result.Order.IsBuy {
			expectedAmtIn = types.Amount0DeltaRounding(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedAmtIn = types.Amount1DeltaRounding(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		}
		amtInDiff := result.Received.Amount.Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			accruedRewards = accruedRewards.Add(sdk.NewCoin(result.Received.Denom, amtInDiff))
			feeGrowth := amtInDiff.ToDec().QuoTruncate(poolState.CurrentLiquidity)
			if result.Order.IsBuy {
				poolState.FeeGrowthGlobal0 = poolState.FeeGrowthGlobal0.Add(feeGrowth)
			} else {
				poolState.FeeGrowthGlobal1 = poolState.FeeGrowthGlobal1.Add(feeGrowth)
			}
		} else if amtInDiff.IsNegative() { // sanity check
			//panic(amtInDiff)
		}

		if result.Fee.IsNegative() {
			accruedRewards = accruedRewards.Add(sdk.NewCoin(result.Fee.Denom, result.Fee.Amount.Neg()))
			feeGrowth := result.Fee.Amount.Neg().ToDec().QuoTruncate(poolState.CurrentLiquidity)
			if result.Fee.Denom == pool.Denom0 {
				poolState.FeeGrowthGlobal0 = poolState.FeeGrowthGlobal0.Add(feeGrowth)
			} else {
				poolState.FeeGrowthGlobal1 = poolState.FeeGrowthGlobal1.Add(feeGrowth)
			}
		}

		if !isBuy && max && nextPrice.Equal(targetPrice) {
			netLiquidity := k.crossTick(ctx, pool.Id, targetTick, poolState.FeeGrowthGlobal0, poolState.FeeGrowthGlobal1)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
		}
		poolState.CurrentPrice = nextPrice
		poolState.CurrentTick = exchangetypes.TickAtPrice(nextPrice, TickPrecision)
	}
	k.SetPoolState(ctx, pool.Id, poolState)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, reserveAddr, types.ModuleName, accruedRewards); err != nil {
		panic(err)
	}
}
