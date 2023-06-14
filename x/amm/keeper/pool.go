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
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot create more than one pool per  market")
		return
	}

	creationFee := k.GetPoolCreationFee(ctx)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, creationFee); err != nil {
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

func (k Keeper) IteratePoolOrders(ctx sdk.Context, pool types.Pool, isBuy bool, cb func(price sdk.Dec, qty sdk.Int) (stop bool)) {
	ts := int32(pool.TickSpacing)
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).AmountOf(pool.DenomOut(isBuy))
	k.IteratePoolOrderTicks(ctx, pool, poolState, isBuy, func(tick int32, liquidity sdk.Int) (stop bool) {
		if !reserveBalance.IsPositive() {
			return true
		}
		// TODO: check out of tick range
		price := exchangetypes.PriceAtTick(tick)
		sqrtPrice := utils.DecApproxSqrt(price)
		var qty sdk.Int
		if isBuy {
			prevSqrtPrice := sdk.MinDec(
				types.SqrtPriceAtTick(tick+ts),
				utils.DecApproxSqrt(poolState.CurrentPrice))
			qty = utils.MinInt(
				reserveBalance.ToDec().QuoTruncate(price).TruncateInt(),
				types.Amount1DeltaRounding(prevSqrtPrice, sqrtPrice, liquidity, false).ToDec().QuoTruncate(price).TruncateInt())
		} else {
			prevSqrtPrice := sdk.MaxDec(
				types.SqrtPriceAtTick(tick-ts),
				utils.DecApproxSqrt(poolState.CurrentPrice))
			qty = utils.MinInt(
				reserveBalance,
				types.Amount0DeltaRounding(prevSqrtPrice, sqrtPrice, liquidity, false))
		}
		if qty.IsPositive() {
			if cb(price, qty) {
				return true
			}
			reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, price, qty))
		}
		return false
	})
}

func (k Keeper) IteratePoolOrderTicks(ctx sdk.Context, pool types.Pool, poolState types.PoolState, isBuy bool, cb func(tick int32, liquidity sdk.Int) (stop bool)) {
	ts := int32(pool.TickSpacing)
	currentTick := poolState.CurrentTick
	orderLiquidity := poolState.CurrentLiquidity
	if isBuy && currentTick%ts == 0 {
		if cb(currentTick, orderLiquidity) {
			return
		}
	}
	iterCb := func(tick int32, tickInfo types.TickInfo) bool {
		if orderLiquidity.IsPositive() {
			for currentTick != tick {
				var nextTick int32
				if isBuy {
					nextTick = (currentTick+ts-1)/ts*ts - ts
				} else {
					nextTick = currentTick/ts*ts + ts
				}
				if cb(nextTick, orderLiquidity) {
					return true
				}
				currentTick = nextTick
			}
		} else {
			currentTick = tick
		}
		if isBuy {
			orderLiquidity = orderLiquidity.Sub(tickInfo.NetLiquidity)
		} else {
			orderLiquidity = orderLiquidity.Add(tickInfo.NetLiquidity)
		}
		return false
	}
	if isBuy {
		k.IterateTickInfosBelowInclusive(ctx, pool.Id, currentTick, iterCb)
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, currentTick, iterCb)
	}
}
