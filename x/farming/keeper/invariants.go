package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

// RegisterInvariants registers all farming invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "staking-reserved",
		StakingReservedAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "remaining-rewards",
		RemainingRewardsAmountInvariant(k))
}

// AllInvariants runs all invariants of the farming module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := StakingReservedAmountInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		return RemainingRewardsAmountInvariant(k)(ctx)
	}
}

// StakingReservedAmountInvariant checks that the balance of StakingReserveAcc greater than the amount of staked, Queued coins in all staking objects.
func StakingReservedAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateStakingReservedAmount(ctx)
		broken := err != nil
		return sdk.FormatInvariant(types.ModuleName, "staking reserved amount invariant broken",
			"the balance of StakingReserveAcc less than the amount of staked, Queued coins in all staking objects"), broken
	}
}

// RemainingRewardsAmountInvariant checks that the balance of the RewardPoolAddresses of all plans greater than the total amount of unwithdrawn reward coins in all reward objects
func RemainingRewardsAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateRemainingRewardsAmount(ctx)
		broken := err != nil
		return sdk.FormatInvariant(types.ModuleName, "remaining rewards amount invariant broken",
			"the balance of the RewardPoolAddresses of all plans less than the total amount of unwithdrawn reward coins in all reward objects"), broken
	}
}
