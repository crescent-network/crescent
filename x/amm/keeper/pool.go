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
	poolExists := false
	k.IteratePoolsByMarket(ctx, marketId, func(pool types.Pool) (stop bool) {
		poolExists = true
		return true
	})
	if poolExists {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot create more than one pool per  market")
		return
	}

	creationFee := k.GetPoolCreationFee(ctx)
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, creationFee); err != nil {
		return
	}

	// Create a new pool
	poolId := k.GetNextPoolIdWithUpdate(ctx) // TODO: reject creating new pool with same parameters
	defaultTickSpacing := k.GetDefaultTickSpacing(ctx)
	pool = types.NewPool(poolId, marketId, market.BaseDenom, market.QuoteDenom, defaultTickSpacing)
	k.SetPool(ctx, pool)
	k.SetPoolsByMarketIndex(ctx, pool)
	k.SetPoolByReserveAddressIndex(ctx, pool)

	// Set initial pool state
	state := types.NewPoolState(exchangetypes.TickAtPrice(price, TickPrecision), price)
	k.SetPoolState(ctx, pool.Id, state)

	return pool, nil
}

func (k Keeper) IteratePoolOrders(ctx sdk.Context, pool types.Pool, isBuy bool, cb func(price sdk.Dec, qty sdk.Int) (stop bool)) {
	ts := int32(pool.TickSpacing)
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).AmountOf(pool.DenomOut(isBuy))
	k.IteratePoolOrderTicks(ctx, pool, poolState, isBuy, func(tick int32, liquidity sdk.Dec) (stop bool) {
		if !reserveBalance.IsPositive() {
			return true
		}
		// TODO: check out of tick range
		price := exchangetypes.PriceAtTick(tick, TickPrecision)
		sqrtPrice := utils.DecApproxSqrt(price)
		var qty sdk.Int
		if isBuy {
			prevSqrtPrice := sdk.MinDec(
				types.SqrtPriceAtTick(tick+ts, TickPrecision),
				utils.DecApproxSqrt(poolState.CurrentPrice))
			qty = utils.MinInt(
				reserveBalance.ToDec().QuoTruncate(price).TruncateInt(),
				types.Amount1DeltaRounding(prevSqrtPrice, sqrtPrice, liquidity, false).ToDec().QuoTruncate(price).TruncateInt())
		} else {
			prevSqrtPrice := sdk.MaxDec(
				types.SqrtPriceAtTick(tick-ts, TickPrecision),
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

func (k Keeper) IteratePoolOrderTicks(ctx sdk.Context, pool types.Pool, poolState types.PoolState, isBuy bool, cb func(tick int32, liquidity sdk.Dec) (stop bool)) {
	ts := int32(pool.TickSpacing)
	currentTick := poolState.CurrentTick / ts * ts
	currentLiquidity := poolState.CurrentLiquidity
	iterCb := func(tick int32, tickInfo types.TickInfo) bool {
		if currentLiquidity.IsPositive() {
			for currentTick != tick {
				var nextTick int32
				if isBuy {
					nextTick = currentTick - ts
				} else {
					nextTick = currentTick + ts
				}
				if cb(nextTick, currentLiquidity) {
					return true
				}
				currentTick = nextTick
			}
		} else {
			currentTick = tick
		}
		if isBuy {
			currentLiquidity = currentLiquidity.Sub(tickInfo.NetLiquidity)
		} else {
			currentLiquidity = currentLiquidity.Add(tickInfo.NetLiquidity)
		}
		return false
	}
	if isBuy {
		k.IterateTickInfosBelow(ctx, pool.Id, currentTick, iterCb)
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, currentTick, iterCb)
	}
}
