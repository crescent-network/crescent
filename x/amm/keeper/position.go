package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) AddLiquidity(
	ctx sdk.Context, ownerAddr, fromAddr sdk.AccAddress, poolId uint64,
	lowerPrice, upperPrice sdk.Dec, desiredAmt sdk.Coins) (position types.Position, liquidity sdk.Int, amt sdk.Coins, err error) {
	// We did this checks in the msg's ValidateBasic, but check it again for
	// the cases when other modules(e.g. liquidamm) call AddLiquidity directly.
	if err = types.ValidatePriceRange(lowerPrice, upperPrice); err != nil {
		return
	}

	pool, found := k.GetPool(ctx, poolId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
		return
	}

	lowerTick := exchangetypes.TickAtPrice(lowerPrice)
	upperTick := exchangetypes.TickAtPrice(upperPrice)
	if lowerTick%int32(pool.TickSpacing) != 0 {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "lower tick %d must be multiple of tick spacing %d",
			lowerTick, pool.TickSpacing)
		return
	}
	if upperTick%int32(pool.TickSpacing) != 0 {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "upper tick %d must be multiple of tick spacing %d",
			upperTick, pool.TickSpacing)
		return
	}

	for _, coin := range desiredAmt {
		if coin.Denom != pool.Denom0 && coin.Denom != pool.Denom1 {
			err = sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "pool doesn't have denom %s", coin.Denom)
			return
		}
	}
	desiredAmt0, desiredAmt1 := desiredAmt.AmountOf(pool.Denom0), desiredAmt.AmountOf(pool.Denom1)
	poolState := k.MustGetPoolState(ctx, poolId)

	lowerSqrtPrice := types.SqrtPriceAtTick(lowerTick)
	upperSqrtPrice := types.SqrtPriceAtTick(upperTick)
	liquidity = types.LiquidityForAmounts(
		poolState.CurrentSqrtPrice, lowerSqrtPrice, upperSqrtPrice, desiredAmt0, desiredAmt1)
	if liquidity.IsZero() {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minted liquidity is zero")
		return
	}

	var amt0, amt1 sdk.Int
	position, amt0, amt1 = k.modifyPosition(
		ctx, pool, ownerAddr, lowerTick, upperTick, liquidity)

	amt = sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
	if !amt.IsAllPositive() { // sanity check
		panic("amt is not positive")
	}
	if err = k.bankKeeper.SendCoins(
		ctx, fromAddr, pool.MustGetReserveAddress(), amt); err != nil {
		return
	}

	if err = ctx.EventManager().EmitTypedEvent(&types.EventAddLiquidity{
		Owner:      ownerAddr.String(),
		PoolId:     poolId,
		LowerPrice: lowerPrice,
		UpperPrice: upperPrice,
		PositionId: position.Id,
		Liquidity:  liquidity,
		Amount:     amt,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) RemoveLiquidity(
	ctx sdk.Context, ownerAddr, toAddr sdk.AccAddress, positionId uint64, liquidity sdk.Int) (position types.Position, amt sdk.Coins, err error) {
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

	pool := k.MustGetPool(ctx, position.PoolId)

	var amt0, amt1 sdk.Int
	position, amt0, amt1 = k.modifyPosition(
		ctx, pool, ownerAddr, position.LowerTick, position.UpperTick, liquidity.Neg())
	amt0, amt1 = amt0.Neg(), amt1.Neg()

	reserveAddr := pool.MustGetReserveAddress()
	reserveBalances := k.bankKeeper.SpendableCoins(ctx, reserveAddr)
	poolState := k.MustGetPoolState(ctx, pool.Id)
	if poolState.TotalLiquidity.IsZero() { // the last liquidity removal from the pool
		amt = reserveBalances
	} else {
		// TODO: do not use sdk.Coins.Min here?
		amt = reserveBalances.Min(
			sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1)))
	}
	if amt.IsAllPositive() {
		if err = k.bankKeeper.SendCoins(ctx, reserveAddr, toAddr, amt); err != nil {
			return
		}
	}
	// Collect owed coins when removing all the liquidity from the position.
	if position.Liquidity.IsZero() {
		var fee, farmingRewards sdk.Coins
		fee, farmingRewards, err = k.CollectibleCoins(ctx, position.Id)
		if err != nil {
			return
		}
		if rewards := fee.Add(farmingRewards...); rewards.IsAllPositive() {
			if err = k.Collect(ctx, ownerAddr, toAddr, position.Id, rewards); err != nil {
				return
			}
		}
	}
	return
}

func (k Keeper) Collect(
	ctx sdk.Context, ownerAddr, toAddr sdk.AccAddress, positionId uint64, amt sdk.Coins) error {
	position, found := k.GetPosition(ctx, positionId)
	if !found {
		return sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	if ownerAddr.String() != position.Owner {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "position is not owned by the user")
	}
	pool := k.MustGetPool(ctx, position.PoolId)

	if position.Liquidity.IsPositive() {
		// Burn zero liquidity to calculate OwedFee, OwedFarmingRewards and
		// update LastFeeGrowthInside, LastFarmingRewardsInside.
		var err error
		position, _, err = k.RemoveLiquidity(ctx, ownerAddr, toAddr, positionId, utils.ZeroInt)
		if err != nil {
			return err
		}
	}
	collectible := position.OwedFee.Add(position.OwedFarmingRewards...)
	if !collectible.IsAllGTE(amt) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "collectible %s is smaller than %s", collectible, amt)
	}
	// Collect fee from the pool's rewards pool first.
	fee := amt.Min(position.OwedFee)
	position.OwedFee = position.OwedFee.Sub(fee)
	if fee.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(ctx, pool.MustGetRewardsPoolAddress(), toAddr, fee); err != nil {
			return err
		}
	}
	// Then collect farming rewards from the farming rewards pool.
	farmingRewards := amt.Sub(fee)
	position.OwedFarmingRewards = position.OwedFarmingRewards.Sub(farmingRewards)
	if farmingRewards.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(ctx, types.FarmingRewardsPoolAddress, toAddr, farmingRewards); err != nil {
			return err
		}
	}
	k.SetPosition(ctx, position)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCollect{
		Owner:      ownerAddr.String(),
		PositionId: positionId,
		Amount:     amt,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) PositionAssets(ctx sdk.Context, positionId uint64) (coin0, coin1 sdk.Coin, err error) {
	position, found := k.GetPosition(ctx, positionId)
	if !found {
		return coin0, coin1, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	pool := k.MustGetPool(ctx, position.PoolId)
	if position.Liquidity.IsZero() {
		coin0 = sdk.NewInt64Coin(pool.Denom0, 0)
		coin1 = sdk.NewInt64Coin(pool.Denom1, 0)
		return
	}
	ctx, _ = ctx.CacheContext()
	_, amt0, amt1 := k.modifyPosition(
		ctx, pool, position.MustGetOwnerAddress(), position.LowerTick, position.UpperTick, position.Liquidity.Neg())
	amt0, amt1 = amt0.Neg(), amt1.Neg()
	coin0 = sdk.NewCoin(pool.Denom0, amt0)
	coin1 = sdk.NewCoin(pool.Denom1, amt1)
	return
}

func (k Keeper) CollectibleCoins(ctx sdk.Context, positionId uint64) (fee, farmingRewards sdk.Coins, err error) {
	position, found := k.GetPosition(ctx, positionId)
	if !found {
		return nil, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	ctx, _ = ctx.CacheContext()
	ownerAddr := position.MustGetOwnerAddress()
	if position.Liquidity.IsPositive() {
		position, _, err = k.RemoveLiquidity(ctx, ownerAddr, ownerAddr, positionId, utils.ZeroInt)
		if err != nil {
			return nil, nil, err
		}
	}
	return position.OwedFee, position.OwedFarmingRewards, nil
}

func (k Keeper) modifyPosition(
	ctx sdk.Context, pool types.Pool, ownerAddr sdk.AccAddress,
	lowerTick, upperTick int32, liquidityDelta sdk.Int) (position types.Position, amt0, amt1 sdk.Int) {
	// TODO: validate ticks
	var found bool
	position, found = k.GetPositionByParams(ctx, ownerAddr, pool.Id, lowerTick, upperTick)
	if !found {
		positionId := k.GetNextPositionIdWithUpdate(ctx)
		position = types.NewPosition(positionId, pool.Id, ownerAddr, lowerTick, upperTick)
		k.SetPositionByParamsIndex(ctx, position)
		k.SetPositionsByPoolIndex(ctx, position)
	}

	if liquidityDelta.IsZero() && !position.Liquidity.IsPositive() { // sanity check
		panic("cannot poke zero liquidity positions")
	}

	// begin _updatePosition()
	poolState := k.MustGetPoolState(ctx, pool.Id)
	var flippedLower, flippedUpper bool
	if !liquidityDelta.IsZero() {
		flippedLower = k.updateTick(
			ctx, pool.Id, lowerTick, poolState.CurrentTick, liquidityDelta, poolState, false)
		flippedUpper = k.updateTick(
			ctx, pool.Id, upperTick, poolState.CurrentTick, liquidityDelta, poolState, true)
	}

	feeGrowthInside, farmingRewardsGrowthInside := k.rewardsGrowthInside(
		ctx, pool.Id, lowerTick, upperTick, poolState)

	var owedFee, owedFarmingRewards sdk.Coins
	if position.Liquidity.IsPositive() {
		feeGrowthDiff := feeGrowthInside.Sub(position.LastFeeGrowthInside)
		owedFee, _ = feeGrowthDiff.
			MulDecTruncate(position.Liquidity.ToDec()).
			QuoDecTruncate(types.DecMulFactor).
			TruncateDecimal()
		farmingRewardsDiff := farmingRewardsGrowthInside.Sub(position.LastFarmingRewardsGrowthInside)
		owedFarmingRewards, _ = farmingRewardsDiff.
			MulDecTruncate(position.Liquidity.ToDec()).
			QuoDecTruncate(types.DecMulFactor).
			TruncateDecimal()
	}

	position.Liquidity = position.Liquidity.Add(liquidityDelta)
	position.LastFeeGrowthInside = feeGrowthInside
	position.OwedFee = position.OwedFee.Add(owedFee...)
	position.LastFarmingRewardsGrowthInside = farmingRewardsGrowthInside
	position.OwedFarmingRewards = position.OwedFarmingRewards.Add(owedFarmingRewards...)
	k.SetPosition(ctx, position)

	if liquidityDelta.IsNegative() {
		if flippedLower {
			k.DeleteTickInfo(ctx, pool.Id, lowerTick)
		}
		if flippedUpper {
			k.DeleteTickInfo(ctx, pool.Id, upperTick)
		}
	}
	// end _updatePosition()

	amt0 = utils.ZeroInt
	amt1 = utils.ZeroInt
	if !liquidityDelta.IsZero() {
		sqrtPriceA := types.SqrtPriceAtTick(lowerTick)
		sqrtPriceB := types.SqrtPriceAtTick(upperTick)
		if poolState.CurrentTick < lowerTick {
			amt0 = types.Amount0Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
		} else if poolState.CurrentTick < upperTick {
			amt0 = types.Amount0Delta(poolState.CurrentSqrtPrice, sqrtPriceB, liquidityDelta)
			amt1 = types.Amount1Delta(sqrtPriceA, poolState.CurrentSqrtPrice, liquidityDelta)
			poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(liquidityDelta)
		} else {
			amt1 = types.Amount1Delta(sqrtPriceA, sqrtPriceB, liquidityDelta)
		}
		poolState.TotalLiquidity = poolState.TotalLiquidity.Add(liquidityDelta)
		k.SetPoolState(ctx, pool.Id, poolState)
	}
	return
}
