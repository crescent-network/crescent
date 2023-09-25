package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "can-cancel-order", ShareLiquidityInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		return ShareLiquidityInvariant(k)(ctx)
	}
}

// HERE! Check if share = 0 & liquidity != 0 or share != 0 & liquidity = 0
func ShareLiquidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return "", false
	}
}
