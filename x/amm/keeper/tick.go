package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) updateTick(ctx sdk.Context, poolId uint64, tick, currentTick int32, liquidityDelta sdk.Dec, upper bool) (flipped bool) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found {
		tickInfo = types.NewTickInfo()
	}

	grossLiquidityBefore := tickInfo.GrossLiquidity
	grossLiquidityAfter := tickInfo.GrossLiquidity.Add(liquidityDelta)
	flipped = grossLiquidityAfter.IsZero() != grossLiquidityBefore.IsZero()

	if grossLiquidityBefore.IsZero() {
		if tick <= currentTick {
			// TODO: set fee growth outside
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
