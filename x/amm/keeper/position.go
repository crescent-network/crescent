package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) AddLiquidity(
	ctx sdk.Context, ownerAddr sdk.AccAddress, poolId uint64, lowerPrice, upperPrice sdk.Dec,
	desiredAmt0, desiredAmt1, minAmt0, minAmt1 sdk.Int) (position types.Position, liquidity sdk.Dec, amt0, amt1 sdk.Int, err error) {
	lowerTick, valid := exchangetypes.ValidateTickPrice(lowerPrice, TickPrecision)
	if !valid {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid lower tick")
		return
	}
	upperTick, valid := exchangetypes.ValidateTickPrice(upperPrice, TickPrecision)
	if !valid {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid upper tick")
		return
	}

	pool, found := k.GetPool(ctx, poolId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
		return
	}
	poolState := k.MustGetPoolState(ctx, poolId)

	sqrtPriceA := types.SqrtPriceAtTick(lowerTick, TickPrecision) // TODO: use tick prec param
	sqrtPriceB := types.SqrtPriceAtTick(upperTick, TickPrecision) // TODO: use tick prec param

	liquidity = types.LiquidityForAmounts(
		poolState.CurrentSqrtPrice, sqrtPriceA, sqrtPriceB, desiredAmt0, desiredAmt1)

	position, amt0, amt1 = k.modifyPosition(
		ctx, pool, ownerAddr, lowerTick, upperTick, liquidity)

	if amt0.LT(minAmt0) || amt1.LT(minAmt1) {
		// TODO: use more verbose error message
		err = sdkerrors.Wrapf(
			types.ErrConditionsNotMet, "(%s, %s) < (%s, %s)", amt0, amt1, minAmt0, minAmt1)
		return
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
	if err = k.bankKeeper.SendCoins(
		ctx, ownerAddr, sdk.MustAccAddressFromBech32(pool.ReserveAddress), depositCoins); err != nil {
		return
	}

	return
}

func (k Keeper) RemoveLiquidity(
	ctx sdk.Context, ownerAddr sdk.AccAddress, positionId uint64, liquidity sdk.Dec, minAmt0, minAmt1 sdk.Int) (position types.Position, amt0, amt1 sdk.Int, err error) {
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
		err = sdkerrors.Wrapf(
			types.ErrConditionsNotMet, "(%s, %s) < (%s, %s)", amt0, amt1, minAmt0, minAmt1)
		return
	}

	withdrawCoins := sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
	if err = k.bankKeeper.SendCoins(
		ctx, sdk.MustAccAddressFromBech32(pool.ReserveAddress), ownerAddr, withdrawCoins); err != nil {
		return
	}

	return
}

func (k Keeper) modifyPosition(
	ctx sdk.Context, pool types.Pool, ownerAddr sdk.AccAddress,
	lowerTick, upperTick int32, liquidityDelta sdk.Dec) (position types.Position, amt0, amt1 sdk.Int) {
	// TODO: validate ticks
	var found bool
	position, found = k.GetPositionByParams(ctx, pool.Id, ownerAddr, lowerTick, upperTick)
	if !found {
		positionId := k.GetNextPositionIdWithUpdate(ctx)
		position = types.NewPosition(positionId, pool.Id, ownerAddr, lowerTick, upperTick)
	}

	poolState := k.MustGetPoolState(ctx, pool.Id)
	flippedLower := k.updateTick(ctx, pool.Id, lowerTick, poolState.CurrentTick, liquidityDelta, false)
	flippedUpper := k.updateTick(ctx, pool.Id, upperTick, poolState.CurrentTick, liquidityDelta, true)

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
	sqrtPriceA := types.SqrtPriceAtTick(lowerTick, TickPrecision)
	sqrtPriceB := types.SqrtPriceAtTick(upperTick, TickPrecision)
	if poolState.CurrentTick < lowerTick {
		amt0 = types.Amount0Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
		amt1 = utils.ZeroInt
	} else if poolState.CurrentTick < upperTick {
		amt0 = types.Amount0Delta(poolState.CurrentSqrtPrice, sqrtPriceB, liquidityDelta)
		amt1 = types.Amount1Delta(sqrtPriceA, poolState.CurrentSqrtPrice, liquidityDelta)
		poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(liquidityDelta)
		k.SetPoolState(ctx, pool.Id, poolState)
	} else {
		amt0 = utils.ZeroInt
		amt1 = types.Amount1Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
	}
	return
}
