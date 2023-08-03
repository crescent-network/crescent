package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) RunValidations(ctx sdk.Context) {
	k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
		pool := k.MustGetPool(ctx, position.PoolId)
		poolState := k.MustGetPoolState(ctx, position.PoolId)

		cacheCtx, _ := ctx.CacheContext()
		ownerAddr := position.MustGetOwnerAddress()
		_, amt, err := k.RemoveLiquidity(cacheCtx, ownerAddr, ownerAddr, position.Id, position.Liquidity)
		if err != nil {
			panic(err)
		}

		types.ValidatePositionState(pool, poolState, position, amt)
		return false
	})
}
