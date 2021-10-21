package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

// RegisterInvariants registers all farming invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "staking-reserved-amount",
		StakingReservedAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "remaining-rewards-amount",
		RemainingRewardsAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "non-negative-outstanding-rewards",
		NonNegativeOutstandingRewardsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "outstanding-rewards-amount",
		OutstandingRewardsAmountInvariant(k))
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

// NonNegativeOutstandingRewardsInvariant checks that all OutstandingRewards are
// non-negative.
func NonNegativeOutstandingRewardsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		k.IterateOutstandingRewards(ctx, func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool) {
			if rewards.Rewards.IsAnyNegative() {
				msg += fmt.Sprintf("\t%v has negative outstanding rewards: %v\n", stakingCoinDenom, rewards.Rewards)
				count++
			}
			return false
		})
		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "non-negative outstanding rewards",
			fmt.Sprintf("found %d staking coin with negative outstanding rewards\n%s", count, msg),
		), broken
	}
}

// OutstandingRewardsAmountInvariant checks that OutstandingRewards are
// consistent with rewards that can be withdrawn.
func OutstandingRewardsAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		totalRewards := sdk.DecCoins{}
		k.IterateOutstandingRewards(ctx, func(stakingCoinDenom string, rewards types.OutstandingRewards) (stop bool) {
			totalRewards = totalRewards.Add(rewards.Rewards...)
			return false
		})
		balances := k.bankKeeper.GetAllBalances(ctx, k.GetRewardsReservePoolAcc(ctx))
		_, hasNeg := sdk.NewDecCoinsFromCoins(balances...).SafeSub(totalRewards)
		broken := hasNeg
		return sdk.FormatInvariant(
			types.ModuleName, "wrong outstanding rewards",
			fmt.Sprintf("balance of rewards reserve pool is less than outstanding rewards\n"+
				"\texpected minimum amount of balance: %s\n"+
				"\tbalance: %s", totalRewards, balances,
			),
		), broken
	}
}
