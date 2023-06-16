package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) updateTick(
	ctx sdk.Context, poolId uint64, tick, currentTick int32, liquidityDelta sdk.Int,
	poolState types.PoolState, upper bool) (flipped bool) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found {
		tickInfo = types.NewTickInfo(utils.ZeroInt, utils.ZeroInt)
	}

	grossLiquidityBefore := tickInfo.GrossLiquidity
	grossLiquidityAfter := tickInfo.GrossLiquidity.Add(liquidityDelta)
	flipped = grossLiquidityAfter.IsZero() != grossLiquidityBefore.IsZero()

	if grossLiquidityBefore.IsZero() {
		if tick <= currentTick {
			tickInfo.FeeGrowthOutside = poolState.FeeGrowthGlobal
			tickInfo.FarmingRewardsGrowthOutside = poolState.FarmingRewardsGrowthGlobal
		}
	}

	tickInfo.GrossLiquidity = grossLiquidityAfter
	if upper {
		tickInfo.NetLiquidity = tickInfo.NetLiquidity.Sub(liquidityDelta)
	} else {
		tickInfo.NetLiquidity = tickInfo.NetLiquidity.Add(liquidityDelta)
	}

	k.SetTickInfo(ctx, poolId, tick, tickInfo)
	return
}

func (k Keeper) feeGrowthInside(
	ctx sdk.Context, poolId uint64, lowerTick, upperTick, currentTick int32,
	feeGrowthGlobal sdk.DecCoins) (feeGrowthInside sdk.DecCoins) {
	lower, found := k.GetTickInfo(ctx, poolId, lowerTick)
	if !found { // sanity check
		panic("lower tick info not found")
	}
	upper, found := k.GetTickInfo(ctx, poolId, upperTick)
	if !found { // sanity check
		panic("upper tick info not found")
	}

	var feeGrowthBelow sdk.DecCoins
	if currentTick >= lowerTick {
		feeGrowthBelow = lower.FeeGrowthOutside
	} else {
		feeGrowthBelow = feeGrowthGlobal.Sub(lower.FeeGrowthOutside)
	}
	var feeGrowthAbove sdk.DecCoins
	if currentTick < upperTick {
		feeGrowthAbove = upper.FeeGrowthOutside
	} else {
		feeGrowthAbove = feeGrowthGlobal.Sub(upper.FeeGrowthOutside)
	}

	feeGrowthInside, _ = feeGrowthGlobal.SafeSub(feeGrowthBelow)
	feeGrowthInside, _ = feeGrowthInside.SafeSub(feeGrowthAbove)
	return
}

func (k Keeper) farmingRewardsGrowthInside(
	ctx sdk.Context, poolId uint64, lowerTick, upperTick, currentTick int32,
	farmingRewardsGrowthGlobal sdk.DecCoins) sdk.DecCoins {
	lower, found := k.GetTickInfo(ctx, poolId, lowerTick)
	if !found { // sanity check
		panic("lower tick info not found")
	}
	upper, found := k.GetTickInfo(ctx, poolId, upperTick)
	if !found { // sanity check
		panic("upper tick info not found")
	}

	var rewardsGrowthBelow sdk.DecCoins
	if currentTick >= lowerTick {
		rewardsGrowthBelow = lower.FarmingRewardsGrowthOutside
	} else {
		rewardsGrowthBelow = farmingRewardsGrowthGlobal.Sub(lower.FarmingRewardsGrowthOutside)
	}
	var rewardsGrowthAbove sdk.DecCoins
	if currentTick < upperTick {
		rewardsGrowthAbove = upper.FarmingRewardsGrowthOutside
	} else {
		rewardsGrowthAbove = farmingRewardsGrowthGlobal.Sub(upper.FarmingRewardsGrowthOutside)
	}
	return farmingRewardsGrowthGlobal.Sub(rewardsGrowthBelow).Sub(rewardsGrowthAbove)
}

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tick int32, poolState types.PoolState) (netLiquidity sdk.Int) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found { // sanity check
		panic("tick info not found")
	}
	tickInfo.FeeGrowthOutside, _ = poolState.FeeGrowthGlobal.SafeSub(tickInfo.FeeGrowthOutside)
	tickInfo.FarmingRewardsGrowthOutside, _ = poolState.FarmingRewardsGrowthGlobal.SafeSub(tickInfo.FarmingRewardsGrowthOutside)
	k.SetTickInfo(ctx, poolId, tick, tickInfo)
	return tickInfo.NetLiquidity
}
