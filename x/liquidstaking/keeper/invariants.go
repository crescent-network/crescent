package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// RegisterInvariants registers all liquidstaking invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "net-amount",
		NetAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "total-liquid-tokens",
		TotalLiquidTokensInvariant(k))
	ir.RegisterRoute(types.ModuleName, "liquid-delegation",
		LiquidDelegationInvariant(k))
}

// AllInvariants runs all invariants of the liquidstaking module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []func(Keeper) sdk.Invariant{
			NetAmountInvariant,
			TotalLiquidTokensInvariant,
			LiquidDelegationInvariant,
		} {
			res, stop := inv(k)(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

// NetAmountInvariant checks that the amount of btoken supply with NetAmount.
func NetAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		broken := false
		lvs := k.GetAllLiquidValidators(ctx)
		if lvs.Len() == 0 {
			return msg, broken
		}
		nas := k.GetNetAmountState(ctx)
		balance := k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc).Amount
		NetAmountExceptBalance := nas.NetAmount.Sub(balance.ToDec())
		liquidBondDenom := k.LiquidBondDenom(ctx)
		bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom)
		if bTokenTotalSupply.IsPositive() && !NetAmountExceptBalance.IsPositive() {
			msg = "found positive btoken supply with non-positive net amount"
			broken = true
		}
		if !bTokenTotalSupply.IsPositive() && NetAmountExceptBalance.IsPositive() {
			msg = "found positive net amount with non-positive btoken supply"
			broken = true
		}
		return sdk.FormatInvariant(
			types.ModuleName, "btoken supply with net amount invariant broken",
			msg,
		), broken
	}
}

// TotalLiquidTokensInvariant checks equal total liquid tokens of proxy account with total liquid tokens of liquid validators.
func TotalLiquidTokensInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		lvs := k.GetAllLiquidValidators(ctx)
		if lvs.Len() == 0 {
			return "", false
		}

		_, _, totalDelegationTokensOfProxyAcc := k.CheckDelegationStates(ctx, types.LiquidStakingProxyAcc)
		totalLiquidTokensOfLiquidValidators, _ := lvs.TotalLiquidTokens(ctx, k.stakingKeeper, false)

		broken := !totalDelegationTokensOfProxyAcc.Equal(totalLiquidTokensOfLiquidValidators)
		return sdk.FormatInvariant(
			types.ModuleName, "total liquid tokens invariant broken",
			fmt.Sprintf("found unmatched total delegation tokens of proxy account %s with total liquid tokens of all liquid validators %s\n",
				totalDelegationTokensOfProxyAcc.String(), totalLiquidTokensOfLiquidValidators.String()),
		), broken
	}
}

// LiquidDelegationInvariant checks all delegation of proxy account involved liquid validators.
func LiquidDelegationInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		liquidValidatorMap := k.GetAllLiquidValidators(ctx).Map()

		// remove delegation condition -> Unbond(slash, undelegate, redelegate)
		// remove validator condition -> Unbond(slash, undelegate, redelegate), UnbondAllMatureValidators(BlockValidatorUpdates on staking endblock)
		k.stakingKeeper.IterateDelegations(
			ctx, types.LiquidStakingProxyAcc,
			func(_ int64, del stakingtypes.DelegationI) (stop bool) {
				delAddr := del.GetValidatorAddr().String()
				if _, ok := liquidValidatorMap[delAddr]; !ok {
					msg += fmt.Sprintf("\t%s has delegation but not liquid validator\n", delAddr)
					count++
				}
				return false
			},
		)

		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "total liquid tokens invariant broken",
			fmt.Sprintf("found %d delegation of proxy account for not liquid validators\n%s", count, msg),
		), broken
	}
}
