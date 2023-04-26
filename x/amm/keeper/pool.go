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
	state := types.NewPoolState(
		exchangetypes.TickAtPrice(price, TickPrecision), utils.DecApproxSqrt(price))
	k.SetPoolState(ctx, pool.Id, state)

	return pool, nil
}

func (k Keeper) iterateTicksBelowPoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, poolState types.PoolState, lowestTick int32, cb func(tick int32, liquidity sdk.Dec) (stop bool)) {
	q, _ := utils.DivMod(poolState.CurrentTick, int32(pool.TickSpacing))
	currentTick := q * int32(pool.TickSpacing)
	liquidity := poolState.CurrentLiquidity
	k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if liquidity.IsPositive() {
			for ; currentTick >= tick && currentTick >= lowestTick; currentTick -= int32(pool.TickSpacing) {
				if cb(currentTick, liquidity) {
					return true
				}
			}
		}
		if tick <= lowestTick {
			return true
		}
		liquidity = liquidity.Sub(tickInfo.NetLiquidity)
		return false
	})
}

func (k Keeper) iterateTicksAbovePoolPriceWithLiquidity(ctx sdk.Context, pool types.Pool, poolState types.PoolState, highestTick int32, cb func(tick int32, liquidity sdk.Dec) (stop bool)) {
	currentTick := (poolState.CurrentTick + int32(pool.TickSpacing)) / int32(pool.TickSpacing) * int32(pool.TickSpacing)
	liquidity := poolState.CurrentLiquidity
	// TODO: What if there's no tick infos above the current pool's tick but
	//       still there's liquidity below highestTick? Is this even possible?
	k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if liquidity.IsPositive() {
			for ; currentTick <= tick && currentTick <= highestTick; currentTick += int32(pool.TickSpacing) {
				if cb(currentTick, liquidity) {
					return true
				}
			}
		}
		if tick >= highestTick {
			return true
		}
		liquidity = liquidity.Add(tickInfo.NetLiquidity)
		return false
	})
}
