package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/cremath"
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
	if creationFee.IsAllPositive() {
		if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, creationFee); err != nil {
			err = sdkerrors.Wrap(err, "insufficient pool creation fee")
			return
		}
	}

	// Create a new pool
	poolId := k.GetNextPoolIdWithUpdate(ctx)
	defaultTickSpacing := k.GetDefaultTickSpacing(ctx)
	defaultMinOrderQty := k.GetDefaultMinOrderQuantity(ctx)
	defaultMinOrderQuote := k.GetDefaultMinOrderQuote(ctx)
	pool = types.NewPool(
		poolId, marketId, market.BaseDenom, market.QuoteDenom, defaultTickSpacing,
		defaultMinOrderQty, defaultMinOrderQuote)
	k.SetPool(ctx, pool)
	k.SetPoolByMarketIndex(ctx, pool)
	k.SetPoolByReserveAddressIndex(ctx, pool)

	// Set initial pool state
	state := types.NewPoolState(
		exchangetypes.TickAtPrice(price),
		cremath.NewBigDecFromDec(price))
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
	liquidity := poolState.CurrentLiquidity
	currentPrice := poolState.CurrentPrice

	iterCb := func(tick int32, tickInfo types.TickInfo) (stop bool) {
		if liquidity.IsPositive() {
			currentPrice, reserveBalance = types.PoolOrders(
				isBuy, currentPrice, liquidity, reserveBalance,
				tick, pool.TickSpacing, pool.MinOrderQuantity, pool.MinOrderQuote, cb)
		} else {
			currentPrice = cremath.NewBigDecFromDec(exchangetypes.PriceAtTick(tick))
		}
		if isBuy {
			liquidity = liquidity.Sub(tickInfo.NetLiquidity)
		} else {
			liquidity = liquidity.Add(tickInfo.NetLiquidity)
		}
		return false
	}
	if isBuy {
		k.IterateTickInfosBelow(ctx, pool.Id, poolState.CurrentTick, true, iterCb)
	} else {
		k.IterateTickInfosAbove(ctx, pool.Id, poolState.CurrentTick, iterCb)
	}
}
