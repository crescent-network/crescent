package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) AddLiquidity(
	ctx sdk.Context, ownerAddr sdk.AccAddress, poolId uint64, lowerTick, upperTick int32,
	desiredAmt0, desiredAmt1, minAmt0, minAmt1 sdk.Int) (position types.Position, liquidity, amt0, amt1 sdk.Int, err error) {
	pool, found := k.GetPool(ctx, poolId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
		return
	}

	if !pool.Initialized {
		if desiredAmt0.IsZero() || desiredAmt1.IsZero() {
			err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must specify both coins for initial liquidity")
			return
		}

		currentPrice := desiredAmt1.ToDec().Quo(desiredAmt0.ToDec())
		var currentSqrtPrice sdk.Dec
		currentSqrtPrice, err = currentPrice.ApproxSqrt()
		if err != nil {
			return
		}

		pool.CurrentSqrtPrice = currentSqrtPrice
		pool.CurrentTick = exchangetypes.TickAtPrice(currentSqrtPrice, TickPrecision) // TODO: use tick prec param
		pool.Initialized = true
		k.SetPool(ctx, pool)
	}

	sqrtPriceA, err := types.SqrtPriceAtTick(lowerTick, TickPrecision) // TODO: use tick prec param
	if err != nil {
		return
	}
	sqrtPriceB, err := types.SqrtPriceAtTick(upperTick, TickPrecision) // TODO: use tick prec param
	if err != nil {
		return
	}
	liquidity = types.LiquidityForAmounts(
		pool.CurrentSqrtPrice, sqrtPriceA, sqrtPriceB, desiredAmt0, desiredAmt1)

	position, amt0, amt1 = k.modifyPosition(
		ctx, pool, ownerAddr, lowerTick, upperTick, liquidity)

	if amt0.LT(minAmt0) || amt1.LT(minAmt1) {
		// TODO: use more verbose error message
		err = types.ErrConditionsNotMet
		return
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
	if err = k.bankKeeper.SendCoins(
		ctx, ownerAddr, sdk.MustAccAddressFromBech32(pool.ReserveAddress), depositCoins); err != nil {
		return
	}

	k.UpdatePoolOrders(ctx, pool, lowerTick, upperTick)

	return
}

func (k Keeper) RemoveLiquidity(
	ctx sdk.Context, ownerAddr sdk.AccAddress, positionId uint64, liquidity, minAmt0, minAmt1 sdk.Int) (position types.Position, amt0, amt1 sdk.Int, err error) {
	var found bool
	position, found = k.GetPosition(ctx, positionId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
		return
	}

	if ownerAddr.String() != position.Owner {
		err = sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "position is not owned by the user")
		return
	}

	if position.Liquidity.LT(liquidity) {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "position liquidity is smaller than the liquidity specified")
		return
	}

	pool, found := k.GetPool(ctx, position.PoolId)
	if !found { // sanity check
		panic("pool not found")
	}

	position, amt0, amt1 = k.modifyPosition(
		ctx, pool, ownerAddr, position.LowerTick, position.UpperTick, liquidity.Neg())
	amt0, amt1 = amt0.Neg(), amt1.Neg()

	if amt0.LT(minAmt0) || amt1.LT(minAmt1) {
		// TODO: use more verbose error message
		err = types.ErrConditionsNotMet
		return
	}

	withdrawCoins := sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
	if err = k.bankKeeper.SendCoins(
		ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), ownerAddr, withdrawCoins); err != nil {
		return
	}

	k.UpdatePoolOrders(ctx, pool, position.LowerTick, position.UpperTick)

	return
}

func (k Keeper) modifyPosition(
	ctx sdk.Context, pool types.Pool, ownerAddr sdk.AccAddress,
	lowerTick, upperTick int32, liquidityDelta sdk.Int) (position types.Position, amt0, amt1 sdk.Int) {
	// TODO: validate ticks
	var found bool
	position, found = k.GetPositionByParams(ctx, pool.Id, ownerAddr, lowerTick, upperTick)
	if !found {
		positionId := k.GetNextPositionIdWithUpdate(ctx)
		position = types.NewPosition(positionId, pool.Id, ownerAddr, lowerTick, upperTick)
	}

	flippedLower := k.updateTick(ctx, pool.Id, lowerTick, pool.CurrentTick, liquidityDelta, false)
	flippedUpper := k.updateTick(ctx, pool.Id, upperTick, pool.CurrentTick, liquidityDelta, true)

	// TODO: get fee growth inside

	// TODO: set position with new fee growth inside
	position.Liquidity = position.Liquidity.Add(liquidityDelta)
	k.SetPosition(ctx, position)

	if liquidityDelta.IsNegative() {
		if flippedLower {
			k.DeleteTickInfo(ctx, pool.Id, lowerTick)
		}
		if flippedUpper {
			k.DeleteTickInfo(ctx, pool.Id, upperTick)
		}
	}

	// TODO: handle prec param and error correctly
	sqrtPriceA, err := types.SqrtPriceAtTick(lowerTick, TickPrecision)
	if err != nil {
		panic(err)
	}
	sqrtPriceB, err := types.SqrtPriceAtTick(upperTick, TickPrecision)
	if err != nil {
		panic(err)
	}
	if pool.CurrentTick < lowerTick {
		amt0 = types.Amount0Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
		amt1 = sdk.ZeroInt()
	} else if pool.CurrentTick < upperTick {
		amt0 = types.Amount0Delta(pool.CurrentSqrtPrice, sqrtPriceB, liquidityDelta)
		amt1 = types.Amount1Delta(sqrtPriceA, pool.CurrentSqrtPrice, liquidityDelta)
		pool.CurrentLiquidity = pool.CurrentLiquidity.Add(liquidityDelta)
		k.SetPool(ctx, pool)
	} else {
		amt0 = sdk.ZeroInt()
		amt1 = types.Amount1Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
	}
	return
}
