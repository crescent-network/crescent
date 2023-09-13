package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) RunValidations(ctx sdk.Context) {
	k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
		if !position.Liquidity.IsPositive() {
			return false
		}
		pool := k.MustGetPool(ctx, position.PoolId)
		poolState := k.MustGetPoolState(ctx, position.PoolId)

		coin0, coin1, err := k.PositionAssets(ctx, position.Id)
		if err != nil {
			panic(err)
		}

		types.ValidatePositionState(pool, poolState, position, sdk.NewCoins(coin0, coin1))
		return false
	})
}
