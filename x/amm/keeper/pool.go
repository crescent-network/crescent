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
	currentTick := poolState.CurrentTick
	placeCurrentTickOrder := isBuy && currentTick%int32(pool.TickSpacing) == 0 &&
		!poolState.CurrentPrice.Equal(exchangetypes.PriceAtTick(currentTick))
	orderLiquidity := poolState.CurrentLiquidity

	currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)

	iterCb := func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if orderLiquidity.IsPositive() {
			for placeCurrentTickOrder || currentTick != tick {
				if !reserveBalance.IsPositive() {
					return true
				}

				var orderTick int32
				if placeCurrentTickOrder {
					orderTick = currentTick
				} else {
					orderTick = types.NextTick(currentTick, pool.TickSpacing, !isBuy)
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
				// TODO: use binary search here to find positive qty order

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

				currentTick = orderTick
				placeCurrentTickOrder = false
			}
		} else {
			currentTick = tick
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
		k.IterateTickInfosBelow(ctx, pool.Id, currentTick, true, iterCb)
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, currentTick, iterCb)
	}
}
