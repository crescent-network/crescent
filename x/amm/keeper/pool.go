package keeper

import (
	"sort"

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
	q, _ := utils.DivMod(poolState.CurrentTick, int32(pool.TickSpacing))
	var startTick int32
	if isBuy {
		startTick = q * int32(pool.TickSpacing)
	} else {
		startTick = (q + 1) * int32(pool.TickSpacing)
	}

	iterCb := func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if orderLiquidity.IsPositive() {
			for {
				if !reserveBalance.IsPositive() {
					return true
				}

				var orderTick int32
				var qty sdk.Dec
				if isBuy {
					i := sort.Search(int((startTick-tick)/int32(pool.TickSpacing)), func(i int) bool {
						orderTick = startTick - int32(i)*int32(pool.TickSpacing)
						orderPrice := exchangetypes.PriceAtTick(orderTick)
						orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
						qty = sdk.MinDec(
							reserveBalance.QuoTruncate(orderPrice),
							types.Amount1DeltaDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity).QuoTruncate(orderPrice))
						return qty.GTE(pool.MinOrderQuantity)
					})
					orderTick = startTick - int32(i)*int32(pool.TickSpacing)
				} else {
					i := sort.Search(int((tick-startTick)/int32(pool.TickSpacing)), func(i int) bool {
						orderTick = startTick + int32(i)*int32(pool.TickSpacing)
						orderPrice := exchangetypes.PriceAtTick(orderTick)
						orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
						qty = sdk.MinDec(
							reserveBalance,
							types.Amount0DeltaRoundingDec(currentSqrtPrice, orderSqrtPrice, orderLiquidity, false))
						return qty.GTE(pool.MinOrderQuantity)
					})
					orderTick = startTick + int32(i)*int32(pool.TickSpacing)
				}
				orderPrice := exchangetypes.PriceAtTick(orderTick)
				orderSqrtPrice := utils.DecApproxSqrt(orderPrice)
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

				startTick = orderTick
				if startTick == tick {
					break
				}
			}
		} else {
			startTick = tick
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
