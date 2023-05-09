package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) updateTick(
	ctx sdk.Context, poolId uint64, tick, currentTick int32, liquidityDelta, feeGrowthGlobal0, feeGrowthGlobal1 sdk.Dec, upper bool) (flipped bool) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found {
		tickInfo = types.NewTickInfo()
	}

	grossLiquidityBefore := tickInfo.GrossLiquidity
	grossLiquidityAfter := tickInfo.GrossLiquidity.Add(liquidityDelta)
	flipped = grossLiquidityAfter.IsZero() != grossLiquidityBefore.IsZero()

	if grossLiquidityBefore.IsZero() {
		if tick <= currentTick {
			tickInfo.FeeGrowthOutside0 = feeGrowthGlobal0
			tickInfo.FeeGrowthOutside1 = feeGrowthGlobal1
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
	feeGrowthGlobal0, feeGrowthGlobal1 sdk.Dec) (feeGrowthInside0, feeGrowthInside1 sdk.Dec) {
	lower, found := k.GetTickInfo(ctx, poolId, lowerTick)
	if !found { // sanity check
		panic("lower tick info not found")
	}
	upper, found := k.GetTickInfo(ctx, poolId, upperTick)
	if !found { // sanity check
		panic("upper tick info not found")
	}

	var feeGrowthBelow0, feeGrowthBelow1 sdk.Dec
	if currentTick >= lowerTick {
		feeGrowthBelow0 = lower.FeeGrowthOutside0
		feeGrowthBelow1 = lower.FeeGrowthOutside1
	} else {
		feeGrowthBelow0 = feeGrowthGlobal0.Sub(lower.FeeGrowthOutside0)
		feeGrowthBelow1 = feeGrowthGlobal1.Sub(lower.FeeGrowthOutside1)
	}
	var feeGrowthAbove0, feeGrowthAbove1 sdk.Dec
	if currentTick < upperTick {
		feeGrowthAbove0 = upper.FeeGrowthOutside0
		feeGrowthAbove1 = upper.FeeGrowthOutside1
	} else {
		feeGrowthAbove0 = feeGrowthGlobal0.Sub(upper.FeeGrowthOutside0)
		feeGrowthAbove1 = feeGrowthGlobal1.Sub(upper.FeeGrowthOutside1)
	}

	feeGrowthInside0 = feeGrowthGlobal0.Sub(feeGrowthBelow0).Sub(feeGrowthAbove0)
	feeGrowthInside1 = feeGrowthGlobal1.Sub(feeGrowthBelow1).Sub(feeGrowthAbove1)
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

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tick int32, poolState types.PoolState) (netLiquidity sdk.Dec) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found { // sanity check
		panic("tick info not found")
	}
	tickInfo.FeeGrowthOutside0 = poolState.FeeGrowthGlobal0.Sub(tickInfo.FeeGrowthOutside0)
	tickInfo.FeeGrowthOutside1 = poolState.FeeGrowthGlobal1.Sub(tickInfo.FeeGrowthOutside1)
	tickInfo.FarmingRewardsGrowthOutside, _ = poolState.FarmingRewardsGrowthGlobal.SafeSub(tickInfo.FarmingRewardsGrowthOutside)
	k.SetTickInfo(ctx, poolId, tick, tickInfo)
	return tickInfo.NetLiquidity
}
