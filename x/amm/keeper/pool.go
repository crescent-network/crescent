package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, senderAddr sdk.AccAddress, denom0, denom1 string, tickSpacing uint32) (types.Pool, error) {
	// TODO: charge pool creation fee from senderAddr
	poolId := k.GetNextPoolIdWithUpdate(ctx) // TODO: reject creating new pool with same parameters

	reserveAddr := types.DerivePoolReserveAddress(poolId)
	pool := types.NewPool(poolId, denom0, denom1, tickSpacing, reserveAddr)
	k.SetPool(ctx, pool)

	return pool, nil
}
