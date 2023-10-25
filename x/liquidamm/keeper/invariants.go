package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "share-supply", ShareSupplyInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		return ShareSupplyInvariant(k)(ctx)
	}
}

func ShareSupplyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		ctx, _ = ctx.CacheContext()
		msg := ""
		cnt := 0
		k.IterateAllPublicPositions(ctx, func(publicPosition types.PublicPosition) (stop bool) {
			shareDenom := types.ShareDenom(publicPosition.Id)
			shareSupply := k.bankKeeper.GetSupply(ctx, shareDenom)
			ammPosition, found := k.GetAMMPosition(ctx, publicPosition)
			if shareSupply.IsZero() {
				if found {
					if !ammPosition.Liquidity.IsZero() {
						msg += fmt.Sprintf(
							"\tpublic position %d should have no liquidity but has %s\n",
							publicPosition.Id, ammPosition.Liquidity)
						cnt++
					} // else ok
				} // else ok
			} else {
				if !found {
					msg += fmt.Sprintf(
						"\tpublic position %d doesn't have amm position but has share supply %s\n",
						publicPosition.Id, shareSupply)
					cnt++
				} else if ammPosition.Liquidity.LT(shareSupply.Amount) {
					msg += fmt.Sprintf(
						"\tpublic position %d should have more liquidity than share supply: %s < %s",
						publicPosition.Id, ammPosition.Liquidity, shareSupply.Amount)
					cnt++
				}
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "share supply",
			fmt.Sprintf(
				"found %d wrong public position state(s)\n%s", cnt, msg)), broken
	}
}
