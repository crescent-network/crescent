package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) updateTick(ctx sdk.Context, poolId uint64, tick, currentTick int32, liquidityDelta sdk.Int, upper bool) {
	tickInfo, found := k.GetTickInfo(ctx, poolId, tick)
	if !found {
		tickInfo = types.NewTickInfo()
	}

	if tickInfo.GrossLiquidity.IsZero() {
		if tick <= currentTick {
			// TODO: set fee growth outside
		}
	}

	tickInfo.GrossLiquidity = tickInfo.GrossLiquidity.Add(liquidityDelta)
	if upper {
		tickInfo.NetLiquidity = tickInfo.NetLiquidity.Sub(liquidityDelta)
	} else {
		tickInfo.NetLiquidity = tickInfo.NetLiquidity.Add(liquidityDelta)
	}

	fmt.Printf("DEBUG: SetTickInfo(%d, %d, %+v)\n", poolId, tick, tickInfo)
	k.SetTickInfo(ctx, poolId, tick, tickInfo)
}
