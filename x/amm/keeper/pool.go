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

func (k Keeper) IteratePoolOrders(ctx sdk.Context, pool types.Pool, isBuy bool, cb func(price, qty sdk.Dec) (stop bool)) {
	poolState := k.MustGetPoolState(ctx, pool.Id)
	reserveBalance := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress()).
		AmountOf(pool.DenomOut(isBuy)).ToDec()
	orderLiquidity := poolState.CurrentLiquidity
	currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)

	iterCb := func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if orderLiquidity.IsPositive() {
			prevTick := poolState.CurrentTick
			for {
				if !reserveBalance.IsPositive() {
					return true
				}

				var orderTick int32
				if isBuy {
					// L^2 + 4 * MinQty * L * sqrt(P_current)
					intermediate := orderLiquidity.ToDec().Power(2).Add(
						pool.MinOrderQuantity.MulInt(orderLiquidity).MulTruncate(currentSqrtPrice).MulInt64(4))
					orderSqrtPrice := utils.DecApproxSqrt(intermediate).Sub(orderLiquidity.ToDec()).
						QuoTruncate(pool.MinOrderQuantity.MulInt64(2))
					orderPrice := orderSqrtPrice.Power(2)
					orderTick = types.AdjustPriceToTickSpacing(orderPrice, pool.TickSpacing, false)
					if orderTick >= prevTick {
						orderTick = types.AdjustTickToTickSpacing(prevTick, pool.TickSpacing, false) - int32(pool.TickSpacing)
					}
					if orderTick < tick {
						orderTick = tick
					}
				} else {
					orderSqrtPrice := currentSqrtPrice.MulInt(orderLiquidity).
						QuoRoundUp(orderLiquidity.ToDec().Sub(pool.MinOrderQuantity.Mul(currentSqrtPrice)))
					orderPrice := orderSqrtPrice.Power(2)
					orderTick = types.AdjustPriceToTickSpacing(orderPrice, pool.TickSpacing, true)
					if orderTick <= prevTick {
						orderTick = types.AdjustTickToTickSpacing(prevTick, pool.TickSpacing, true) + int32(pool.TickSpacing)
					}
					if orderTick > tick {
						orderTick = tick
					}
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
				var qty sdk.Dec
				if isBuy {
					qty = sdk.MinDec(
						reserveBalance.QuoTruncate(orderPrice),
						types.Amount1DeltaDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity).QuoTruncate(orderPrice))
				} else {
					qty = sdk.MinDec(
						reserveBalance,
						types.Amount0DeltaRoundingDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false))
				}
				if qty.GTE(pool.MinOrderQuantity) || orderTick == tick {
					if cb(orderPrice, qty) {
						return true
					}
					reserveBalance = reserveBalance.Sub(exchangetypes.DepositAmount(isBuy, orderPrice, qty))
					currentSqrtPrice = orderSqrtPrice
				}

				if orderTick == tick {
					break
				}
				prevTick = orderTick
			}
		} else {
			currentSqrtPrice = types.SqrtPriceAtTick(tick)
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
