package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// RegisterInvariants registers all liquidstaking invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "btoken-supply-with-del-shares",
		BTokenSupplyWithTotalDelSharesInvariant(k))
}

// AllInvariants runs all invariants of the liquidstaking module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []func(Keeper) sdk.Invariant{
			BTokenSupplyWithTotalDelSharesInvariant,
		} {
			res, stop := inv(k)(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

// BTokenSupplyWithTotalDelSharesInvariant checks that the amount of btoken supply with total delegation shares of liquid staking.
func BTokenSupplyWithTotalDelSharesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		broken := false
		lvs := k.GetAllLiquidValidators(ctx)
		if lvs.Len() == 0 {
			return msg, broken
		}
		totalDelShares := lvs.TotalDelShares(ctx, k.stakingKeeper)
		bondedBondDenom := k.BondedBondDenom(ctx)
		bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, bondedBondDenom)
		if bTokenTotalSupply.IsPositive() && !totalDelShares.IsPositive() {
			msg = "found btoken supply with non-positive totalDelShares"
			broken = true
		}
		UBDs := k.GetLiquidUnbonding(ctx, types.LiquidStakingProxyAcc)
		if !bTokenTotalSupply.IsPositive() && totalDelShares.IsPositive() && len(UBDs) == 0 {
			msg = "found totalDelShares with non-positive btoken supply"
			broken = true
		}
		return sdk.FormatInvariant(
			types.ModuleName, "bonded token supply with total del shares invariant broken",
			msg,
		), broken
	}
}
