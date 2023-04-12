package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, creatorAddr sdk.AccAddress, denom0, denom1 string, tickSpacing uint32, price sdk.Dec) (pool types.Pool, err error) {
	// TODO: charge pool creation fee from senderAddr
	poolId := k.GetNextPoolIdWithUpdate(ctx) // TODO: reject creating new pool with same parameters

	var sqrtPrice sdk.Dec
	sqrtPrice, err = price.ApproxSqrt()
	if err != nil {
		return
	}
	reserveAddr := types.DerivePoolReserveAddress(poolId)
	pool = types.NewPool(
		poolId, denom0, denom1, tickSpacing, reserveAddr,
		exchangetypes.TickAtPrice(sqrtPrice, TickPrecision), sqrtPrice)
	k.SetPool(ctx, pool)
	k.SetPoolsByMarketIndex(ctx, pool)
	k.SetPoolByReserveAddressIndex(ctx, pool)

	return pool, nil
}

func (k Keeper) UpdatePoolOrders(ctx sdk.Context, pool types.Pool, lowerTick, upperTick int32) {
	marketId := exchangetypes.DeriveMarketId(pool.Denom0, pool.Denom1)
	if _, found := k.exchangeKeeper.GetSpotMarket(ctx, marketId); found {
		k.UpdateSpotMarketOrders(ctx, marketId, lowerTick, upperTick)
	}
}

func (k Keeper) UpdateSpotMarketOrders(ctx sdk.Context, marketId string, lowerTick, upperTick int32) {
	k.IteratePoolsByMarket(ctx, marketId, func(pool types.Pool) (stop bool) {
		k.updateSpotMarketOrders(ctx, marketId, pool, lowerTick, upperTick)
		return false
	})
}

func (k Keeper) updateSpotMarketOrders(
	ctx sdk.Context, marketId string,
	pool types.Pool, lowerTick, upperTick int32) {
	market, found := k.exchangeKeeper.GetSpotMarket(ctx, marketId)
	if !found {
		panic("market not found")
	}
	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
	initialReserves := k.bankKeeper.SpendableCoins(ctx, reserveAddr)
	reserve0, reserve1 := initialReserves.AmountOf(pool.Denom0), initialReserves.AmountOf(pool.Denom1)
	// TODO: cancel previous orders
	k.IterateTicksBelowPoolPriceWithLiquidity(ctx, pool, lowerTick, func(tick int32, liquidity sdk.Int) {
		prevOrderId, found := k.GetPoolOrder(ctx, pool.Id, marketId, tick)
		if found {
			prevOrder, err := k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, marketId, prevOrderId)
			if err != nil {
				panic(err)
			}
			k.DeletePoolOrder(ctx, pool.Id, marketId, tick) // TODO: use cancel hook to delete pool order
			if prevOrder.IsBuy {
				reserve1 = reserve1.Add(prevOrder.RemainingDeposit)
			} else {
				reserve0 = reserve0.Add(prevOrder.RemainingDeposit)
			}
		}
		// TODO: check out of tick range
		sqrtPriceAbove, err := types.SqrtPriceAtTick(tick+int32(pool.TickSpacing), TickPrecision)
		if err != nil {
			panic(err)
		}
		sqrtPriceAbove = sdk.MinDec(pool.CurrentSqrtPrice, sqrtPriceAbove)
		price := exchangetypes.PriceAtTick(tick, TickPrecision)
		sqrtPrice, err := price.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		qty := sdk.MinInt(
			reserve1.ToDec().QuoTruncate(price).TruncateInt(),
			sqrtPriceAbove.Sub(sqrtPrice).MulInt(liquidity).QuoTruncate(price).TruncateInt())
		if qty.IsPositive() {
			order, execQuote, err := k.exchangeKeeper.PlaceSpotLimitOrder(
				ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market,
				true, price, qty)
			if err != nil {
				panic(err)
			}
			if !execQuote.IsZero() { // sanity check
				panic("pool order matched with another order")
			}
			k.SetPoolOrder(ctx, pool.Id, marketId, tick, order.Id)
			reserve1 = reserve1.Sub(exchangetypes.DepositAmount(true, price, qty))
		}
	})
	k.IterateTicksAbovePoolPriceWithLiquidity(ctx, pool, upperTick, func(tick int32, liquidity sdk.Int) {
		prevOrderId, found := k.GetPoolOrder(ctx, pool.Id, marketId, tick)
		if found {
			prevOrder, err := k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, marketId, prevOrderId)
			if err != nil {
				panic(err)
			}
			k.DeletePoolOrder(ctx, pool.Id, marketId, tick)
			if prevOrder.IsBuy {
				reserve1 = reserve1.Add(prevOrder.RemainingDeposit)
			} else {
				reserve0 = reserve0.Add(prevOrder.RemainingDeposit)
			}
		}
		sqrtPriceBelow, err := types.SqrtPriceAtTick(tick-int32(pool.TickSpacing), TickPrecision)
		if err != nil {
			panic(err)
		}
		sqrtPriceBelow = sdk.MaxDec(pool.CurrentSqrtPrice, sqrtPriceBelow)
		price := exchangetypes.PriceAtTick(tick, TickPrecision) // TODO: use tick prec param
		sqrtPrice, err := price.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		qty := sdk.MinInt(
			reserve0,
			utils.OneDec.QuoTruncate(sqrtPriceBelow).Sub(utils.OneDec.QuoRoundUp(sqrtPrice)).MulInt(liquidity).TruncateInt())
		if qty.IsPositive() {
			order, execQuote, err := k.exchangeKeeper.PlaceSpotLimitOrder(
				ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market,
				false, price, qty)
			if err != nil {
				panic(err)
			}
			if !execQuote.IsZero() { // sanity check
				panic("pool order matched with another order")
			}
			k.SetPoolOrder(ctx, pool.Id, marketId, tick, order.Id)
			reserve0 = reserve0.Sub(qty)
		}
	})
}

func (k Keeper) IterateTicksBelowPoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, lowestTick int32, cb func(tick int32, liquidity sdk.Int)) {
	q, _ := utils.DivMod(pool.CurrentTick, int32(pool.TickSpacing))
	currentTick := q * int32(pool.TickSpacing)
	liquidity := pool.CurrentLiquidity
	k.IterateTickInfosBelow(ctx, pool.Id, pool.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
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

func (k Keeper) IterateTicksAbovePoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, highestTick int32, cb func(tick int32, liquidity sdk.Int)) {
	currentTick := pool.CurrentTick / int32(pool.TickSpacing) * int32(pool.TickSpacing) // TODO: check division
	liquidity := pool.CurrentLiquidity
	// TODO: What if there's no tick infos above the current pool's tick but
	//       still there's liquidity below highestTick? Is this even possible?
	k.IterateTickInfosAbove(ctx, pool.Id, pool.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
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
