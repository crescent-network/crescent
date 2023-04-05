package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, creatorAddr sdk.AccAddress, denom0, denom1 string, tickSpacing uint32) (types.Pool, error) {
	// TODO: charge pool creation fee from senderAddr
	poolId := k.GetNextPoolIdWithUpdate(ctx) // TODO: reject creating new pool with same parameters

	reserveAddr := types.DerivePoolReserveAddress(poolId)
	pool := types.NewPool(poolId, denom0, denom1, tickSpacing, reserveAddr)
	k.SetPool(ctx, pool)
	k.SetPoolIndex(ctx, pool)

	return pool, nil
}

func (k Keeper) UpdateOrders(ctx sdk.Context, marketId string, lowerTick, upperTick int32) error {
	market, found := k.exchangeKeeper.GetSpotMarket(ctx, marketId)
	if !found { // sanity check
		panic("market not found")
	}

	k.IteratePoolsByMarketId(ctx, market.Id, func(pool types.Pool) (stop bool) {
		k.updateOrders(ctx, market, pool, lowerTick, upperTick)
		return false
	})

	return nil
}

func (k Keeper) updateOrders(
	ctx sdk.Context, market exchangetypes.SpotMarket,
	pool types.Pool, lowerTick, upperTick int32) {
	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
	initialReserves := k.bankKeeper.SpendableCoins(ctx, reserveAddr)
	reserve0, reserve1 := initialReserves.AmountOf(pool.Denom0), initialReserves.AmountOf(pool.Denom1)
	k.IterateTicksBelowPoolPriceWithLiquidity(ctx, pool, lowerTick, func(tick int32, liquidity sdk.Int) {
		price := exchangetypes.PriceAtTick(tick, 4) // TODO: use tick prec param
		sqrtPrice, err := price.ApproxSqrt()
		if err != nil { // TODO: return error
			panic(err)
		}
		sqrtPriceAbove, err := types.SqrtPriceAtTick(tick+int32(pool.TickSpacing), 4)
		if err != nil {
			panic(err)
		}
		sqrtPriceAbove = sdk.MinDec(pool.CurrentSqrtPrice, sqrtPriceAbove)
		qty := sdk.MinInt(
			reserve1.ToDec().QuoTruncate(price).TruncateInt(),
			sqrtPriceAbove.Sub(sqrtPrice).MulInt(liquidity).QuoTruncate(price).TruncateInt())
		reserve1 = reserve1.Sub(market.OfferCoin(true, price, qty).Amount)
		_, _, err = k.exchangeKeeper.PlaceSpotOrder(
			ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market.Id,
			true, &price, qty)
		if err != nil {
			panic(err)
		}
	})
	k.IterateTicksAbovePoolPriceWithLiquidity(ctx, pool, upperTick, func(tick int32, liquidity sdk.Int) {
		price := exchangetypes.PriceAtTick(tick, 4) // TODO: use tick prec param
		sqrtPrice, err := price.ApproxSqrt()
		if err != nil { // TODO: return error
			panic(err)
		}
		sqrtPriceBelow, err := types.SqrtPriceAtTick(tick-int32(pool.TickSpacing), 4)
		if err != nil {
			panic(err)
		}
		sqrtPriceBelow = sdk.MaxDec(pool.CurrentSqrtPrice, sqrtPriceBelow)
		qty := sdk.MinInt(
			reserve0,
			sdk.OneDec().QuoTruncate(sqrtPriceBelow).Sub(sdk.OneDec().QuoTruncate(sqrtPrice)).MulInt(liquidity).TruncateInt())
		reserve0 = reserve0.Sub(qty)
		_, _, err = k.exchangeKeeper.PlaceSpotOrder(
			ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), market.Id,
			false, &price, qty)
		if err != nil {
			panic(err)
		}
	})
}

func (k Keeper) IterateTicksBelowPoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, lowestTick int32, cb func(tick int32, liquidity sdk.Int)) {
	currentTick := pool.CurrentTick
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
	currentTick := pool.CurrentTick
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
