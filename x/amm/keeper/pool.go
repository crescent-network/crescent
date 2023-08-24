package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, creatorAddr sdk.AccAddress, marketId uint64, price sdk.Dec) (pool types.Pool, err error) {
	market, found := k.exchangeKeeper.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}
	if found := k.LookupPoolByMarket(ctx, market.Id); found {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot create more than one pool per market")
		return
	}

	creationFee := k.GetPoolCreationFee(ctx)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, creationFee); err != nil {
		err = sdkerrors.Wrap(err, "insufficient pool creation fee")
		return
	}

	// Create a new pool
	poolId := k.GetNextPoolIdWithUpdate(ctx)
	defaultTickSpacing := k.GetDefaultTickSpacing(ctx)
	pool = types.NewPool(poolId, marketId, market.BaseDenom, market.QuoteDenom, defaultTickSpacing)
	k.SetPool(ctx, pool)
	k.SetPoolByMarketIndex(ctx, pool)
	k.SetPoolByReserveAddressIndex(ctx, pool)

	// Set initial pool state
	state := types.NewPoolState(exchangetypes.TickAtPrice(price), price)
	k.SetPoolState(ctx, pool.Id, state)

	if err = ctx.EventManager().EmitTypedEvent(&types.EventCreatePool{
		Creator:  creatorAddr.String(),
		MarketId: marketId,
		Price:    price,
		PoolId:   poolId,
	}); err != nil {
		return
	}

	return pool, nil
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
				if openQty.GTE(pool.MinOrderQuantity) || orderTick == tick {
					if cb(orderPrice, qty, openQty) {
						return true
					}
					reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, orderPrice, qty))
					currentPrice = orderPrice
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
		if !orderSqrtPrice.IsPositive() || orderSqrtPrice.GTE(currentSqrtPrice) {
			return 0, false
		}
		// 2. Check min order quote
		orderSqrtPrice2 := currentSqrtPrice.Mul(liquidityDec).Sub(minOrderQuote).QuoTruncate(liquidityDec)
		if !orderSqrtPrice2.IsPositive() || orderSqrtPrice2.GTE(currentSqrtPrice) {
			return 0, false
		}
		orderPrice := sdk.MinDec(orderSqrtPrice, orderSqrtPrice2).Power(2)
		tick = types.AdjustPriceToTickSpacing(orderPrice, tickSpacing, false)
		return tick, true
	}
	// 1. Check min order qty
	orderSqrtPrice := currentSqrtPrice.Mul(liquidityDec).
		QuoRoundUp(liquidityDec.Sub(minOrderQty.Mul(currentSqrtPrice)))
	if !orderSqrtPrice.IsPositive() || orderSqrtPrice.LTE(currentSqrtPrice) {
		return 0, false
	}
	// 2. Check min order quote
	orderSqrtPrice2 := minOrderQuote.Mul(currentSqrtPrice).QuoRoundUp(liquidityDec).Add(currentSqrtPrice)
	if !orderSqrtPrice2.IsPositive() || orderSqrtPrice2.LTE(currentSqrtPrice) {
		return 0, false
	}
	orderPrice := sdk.MaxDec(orderSqrtPrice, orderSqrtPrice2).Power(2)
	tick = types.AdjustPriceToTickSpacing(orderPrice, tickSpacing, true)
	return tick, true
}
