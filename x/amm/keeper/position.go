package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) AddLiquidity(
	ctx sdk.Context, senderAddr sdk.AccAddress, poolId uint64, lowerTick, upperTick int32,
	desiredAmt0, desiredAmt1, minAmt0, minAmt1 sdk.Int) (position types.Position, liquidity, amt0, amt1 sdk.Int, err error) {
	pool, found := k.GetPool(ctx, poolId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
		return
	}

	// TODO: if this is the first liquidity, initialize pool price

	sqrtPriceA, err := types.SqrtPriceFromTick(lowerTick, 4) // TODO: use tick prec param
	if err != nil {
		return
	}
	sqrtPriceB, err := types.SqrtPriceFromTick(upperTick, 4) // TODO: use tick prec param
	if err != nil {
		return
	}
	liquidity = types.LiquidityForAmounts(pool.CurrentSqrtPrice, sqrtPriceA, sqrtPriceB, desiredAmt0, desiredAmt1)
	fmt.Printf("DEBUG: liquidity=%s\n", liquidity)

	position, amt0, amt1 = k.modifyPosition(ctx, pool, senderAddr, lowerTick, upperTick, liquidity)

	return
}

func (k Keeper) RemoveLiquidity(
	ctx sdk.Context, senderAddr sdk.AccAddress, positionId uint64, liquidity, minAmt0, minAmt1 sdk.Int) (position types.Position, amt0, amt1 sdk.Int, err error) {
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

	k.updateTick(ctx, pool.Id, lowerTick, pool.CurrentTick, liquidityDelta, false)
	k.updateTick(ctx, pool.Id, upperTick, pool.CurrentTick, liquidityDelta, true)

	// TODO: get fee growth inside

	// TODO: set position with new fee growth inside
	position.Liquidity = position.Liquidity.Add(liquidityDelta)
	k.SetPosition(ctx, position)

	// calculate amt0, amt1
	return
}
