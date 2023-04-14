package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, creatorAddr sdk.AccAddress, denom0, denom1 string, tickSpacing uint32, price sdk.Dec) (pool types.Pool, err error) {
	// Charge pool creation fee to the module account
	creationFee := k.GetPoolCreationFee(ctx)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, creationFee); err != nil {
		return pool, err
	}

	// Create a new pool
	poolId := k.GetNextPoolIdWithUpdate(ctx) // TODO: reject creating new pool with same parameters
	reserveAddr := types.DerivePoolReserveAddress(poolId)
	pool = types.NewPool(poolId, denom0, denom1, tickSpacing, reserveAddr)
	k.SetPool(ctx, pool)
	k.SetPoolsByMarketIndex(ctx, pool)
	k.SetPoolByReserveAddressIndex(ctx, pool)

	// Set initial pool state
	state := types.NewPoolState(
		exchangetypes.TickAtPrice(price, TickPrecision), utils.DecApproxSqrt(price))
	k.SetPoolState(ctx, pool.Id, state)

	return pool, nil
}

func (k Keeper) UpdatePoolOrders(ctx sdk.Context, pool types.Pool, lowerTick, upperTick int32) {
	marketId := exchangetypes.DeriveMarketId(pool.Denom0, pool.Denom1)
	market, found := k.exchangeKeeper.GetSpotMarket(ctx, marketId)
	if found {
		k.updateSpotMarketOrders(ctx, market, pool, lowerTick, upperTick)
	}
}

func (k Keeper) UpdateSpotMarketOrders(ctx sdk.Context, market exchangetypes.SpotMarket, lowerTick, upperTick int32) {
	k.IteratePoolsByMarket(ctx, market.Id, func(pool types.Pool) (stop bool) {
		k.updateSpotMarketOrders(ctx, market, pool, lowerTick, upperTick)
		return false
	})
}

func (k Keeper) updateSpotMarketOrders(
	ctx sdk.Context, market exchangetypes.SpotMarket,
	pool types.Pool, lowerTick, upperTick int32) {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
	initialReserves := k.bankKeeper.SpendableCoins(ctx, reserveAddr)
	reserve0, reserve1 := initialReserves.AmountOf(pool.Denom0), initialReserves.AmountOf(pool.Denom1)
	// TODO: cancel previous orders
	k.iterateTicksBelowPoolPriceWithLiquidity(ctx, pool, poolState, lowerTick, func(tick int32, liquidity sdk.Dec) {
		prevOrderId, found := k.GetPoolOrder(ctx, pool.Id, market.Id, tick)
		if found {
			prevOrder, err := k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, market.Id, prevOrderId)
			if err != nil {
				panic(err)
			}
			k.DeletePoolOrder(ctx, pool.Id, market.Id, tick) // TODO: use cancel hook to delete pool order
			if prevOrder.IsBuy {
				reserve1 = reserve1.Add(prevOrder.RemainingDeposit)
			} else {
				reserve0 = reserve0.Add(prevOrder.RemainingDeposit)
			}
		}
		// TODO: check out of tick range
		sqrtPriceAbove := types.SqrtPriceAtTick(tick+int32(pool.TickSpacing), TickPrecision)
		sqrtPriceAbove = sdk.MinDec(poolState.CurrentSqrtPrice, sqrtPriceAbove)
		price := exchangetypes.PriceAtTick(tick, TickPrecision)
		sqrtPrice := utils.DecApproxSqrt(price)
		qty := utils.MinInt(
			reserve1.ToDec().QuoTruncate(price).TruncateInt(),
			types.Amount1DeltaRounding(sqrtPrice, sqrtPriceAbove, liquidity, false).ToDec().QuoTruncate(price).TruncateInt())
		if qty.IsPositive() {
			order, execQuote, err := k.exchangeKeeper.PlaceSpotLimitOrder(
				ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market.Id,
				true, price, qty)
			if err != nil {
				panic(err)
			}
			if !execQuote.IsZero() { // sanity check
				panic("pool order matched with another order")
			}
			k.SetPoolOrder(ctx, pool.Id, market.Id, tick, order.Id)
			reserve1 = reserve1.Sub(exchangetypes.DepositAmount(true, price, qty))
		}
	})
	k.iterateTicksAbovePoolPriceWithLiquidity(ctx, pool, poolState, upperTick, func(tick int32, liquidity sdk.Dec) {
		prevOrderId, found := k.GetPoolOrder(ctx, pool.Id, market.Id, tick)
		if found {
			prevOrder, err := k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, market.Id, prevOrderId)
			if err != nil {
				panic(err)
			}
			k.DeletePoolOrder(ctx, pool.Id, market.Id, tick)
			if prevOrder.IsBuy {
				reserve1 = reserve1.Add(prevOrder.RemainingDeposit)
			} else {
				reserve0 = reserve0.Add(prevOrder.RemainingDeposit)
			}
		}
		sqrtPriceBelow := types.SqrtPriceAtTick(tick-int32(pool.TickSpacing), TickPrecision)
		sqrtPriceBelow = sdk.MaxDec(poolState.CurrentSqrtPrice, sqrtPriceBelow)
		price := exchangetypes.PriceAtTick(tick, TickPrecision) // TODO: use tick prec param
		sqrtPrice := utils.DecApproxSqrt(price)
		qty := utils.MinInt(
			reserve0,
			types.Amount0DeltaRounding(sqrtPriceBelow, sqrtPrice, liquidity, false))
		if qty.IsPositive() {
			order, execQuote, err := k.exchangeKeeper.PlaceSpotLimitOrder(
				ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market.Id,
				false, price, qty)
			if err != nil {
				panic(err)
			}
			if !execQuote.IsZero() { // sanity check
				panic("pool order matched with another order")
			}
			k.SetPoolOrder(ctx, pool.Id, market.Id, tick, order.Id)
			reserve0 = reserve0.Sub(qty)
		}
	})
}

func (k Keeper) iterateTicksBelowPoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, poolState types.PoolState, lowestTick int32, cb func(tick int32, liquidity sdk.Dec)) {
	q, _ := utils.DivMod(poolState.CurrentTick, int32(pool.TickSpacing))
	currentTick := q * int32(pool.TickSpacing)
	liquidity := poolState.CurrentLiquidity
	k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if liquidity.IsPositive() {
			for ; currentTick >= tick && currentTick >= lowestTick; currentTick -= int32(pool.TickSpacing) {
				cb(currentTick, liquidity)
			}
		}
		if tick <= lowestTick {
			return true
		}
		liquidity = liquidity.Add(tickInfo.NetLiquidity)
		return false
	})
}

func (k Keeper) iterateTicksAbovePoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, poolState types.PoolState, highestTick int32, cb func(tick int32, liquidity sdk.Dec)) {
	currentTick := (poolState.CurrentTick + int32(pool.TickSpacing)) / int32(pool.TickSpacing) * int32(pool.TickSpacing)
	liquidity := poolState.CurrentLiquidity
	// TODO: What if there's no tick infos above the current pool's tick but
	//       still there's liquidity below highestTick? Is this even possible?
	k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if liquidity.IsPositive() {
			for ; currentTick <= tick && currentTick <= highestTick; currentTick += int32(pool.TickSpacing) {
				cb(currentTick, liquidity)
			}
		}
		if tick >= highestTick {
			return true
		}
		liquidity = liquidity.Add(tickInfo.NetLiquidity)
		return false
	})
}

func (k Keeper) PlacePoolOrder(
	ctx sdk.Context, pool types.Pool, poolState types.PoolState, marketId string,
	isBuy bool, tick int32) error {
	price := exchangetypes.PriceAtTick(tick, TickPrecision)
	sqrtPrice := utils.DecApproxSqrt(price)
	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)

	// Cancel previous order
	prevOrderId, found := k.GetPoolOrder(ctx, pool.Id, marketId, tick)
	if found {
		if _, err := k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, marketId, prevOrderId); err != nil {
			return err
		}
		k.DeletePoolOrder(ctx, pool.Id, marketId, tick) // TODO: use cancel hook to delete pool order
	}

	// TODO: is it ok to use pool's current liquidity?
	var qty sdk.Int
	if isBuy {
		quote := utils.MinInt(
			k.bankKeeper.GetBalance(ctx, reserveAddr, pool.Denom1).Amount,
			types.Amount1DeltaRounding(sqrtPrice, poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, false))
		qty = quote.ToDec().QuoTruncate(price).TruncateInt()
	} else {
		qty = utils.MinInt(
			k.bankKeeper.GetBalance(ctx, reserveAddr, pool.Denom0).Amount,
			types.Amount0DeltaRounding(sqrtPrice, poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, false))
	}
	if qty.IsPositive() {
		order, execQuote, err := k.exchangeKeeper.PlaceSpotLimitOrder(
			ctx, reserveAddr, marketId, isBuy, price, qty)
		if err != nil {
			return err
		}
		if !execQuote.IsZero() { // sanity check
			panic("pool order matched with another order")
		}
		k.SetPoolOrder(ctx, pool.Id, order.MarketId, tick, order.Id)
	}
	return nil
}
