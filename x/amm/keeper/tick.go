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

func (k Keeper) rewardsGrowthInside(
	ctx sdk.Context, poolId uint64, lowerTick, upperTick int32,
	poolState types.PoolState) (feeGrowthInside, farmingRewardsGrowthInside sdk.DecCoins) {
	lower := k.MustGetTickInfo(ctx, poolId, lowerTick)
	upper := k.MustGetTickInfo(ctx, poolId, upperTick)

	var feeGrowthBelow, farmingRewardsGrowthBelow sdk.DecCoins
	if poolState.CurrentTick >= lowerTick {
		feeGrowthBelow = lower.FeeGrowthOutside
		farmingRewardsGrowthBelow = lower.FarmingRewardsGrowthOutside
	} else {
		feeGrowthBelow = poolState.FeeGrowthGlobal.Sub(lower.FeeGrowthOutside)
		farmingRewardsGrowthBelow = poolState.FarmingRewardsGrowthGlobal.
			Sub(lower.FarmingRewardsGrowthOutside)
	}
	var feeGrowthAbove, farmingRewardsGrowthAbove sdk.DecCoins
	if poolState.CurrentTick < upperTick {
		feeGrowthAbove = upper.FeeGrowthOutside
		farmingRewardsGrowthAbove = upper.FarmingRewardsGrowthOutside
	} else {
		feeGrowthAbove = poolState.FeeGrowthGlobal.Sub(upper.FeeGrowthOutside)
		farmingRewardsGrowthAbove = poolState.FarmingRewardsGrowthGlobal.
			Sub(upper.FarmingRewardsGrowthOutside)
	}

	feeGrowthInside, _ = poolState.FeeGrowthGlobal.SafeSub(feeGrowthBelow)
	feeGrowthInside, _ = feeGrowthInside.SafeSub(feeGrowthAbove)
	farmingRewardsGrowthInside, _ = poolState.FarmingRewardsGrowthGlobal.
		SafeSub(farmingRewardsGrowthBelow)
	farmingRewardsGrowthInside, _ = farmingRewardsGrowthInside.
		SafeSub(farmingRewardsGrowthAbove)
	return feeGrowthInside, farmingRewardsGrowthInside
}

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tick int32, poolState types.PoolState) (netLiquidity sdk.Int) {
	tickInfo := k.MustGetTickInfo(ctx, poolId, tick)
	tickInfo.FeeGrowthOutside, _ = poolState.FeeGrowthGlobal.SafeSub(tickInfo.FeeGrowthOutside)
	tickInfo.FarmingRewardsGrowthOutside, _ = poolState.FarmingRewardsGrowthGlobal.SafeSub(tickInfo.FarmingRewardsGrowthOutside)
	k.SetTickInfo(ctx, poolId, tick, tickInfo)
	return tickInfo.NetLiquidity
}
