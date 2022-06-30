package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// RegisterInvariants registers all farming invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "positive-staking-amount",
		PositiveStakingAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "positive-queued-staking-amount",
		PositiveQueuedStakingAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "staking-reserved-amount",
		StakingReservedAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "remaining-rewards-amount",
		RemainingRewardsAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "non-negative-outstanding-rewards",
		NonNegativeOutstandingRewardsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "outstanding-rewards-amount",
		OutstandingRewardsAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "unharvested-rewards-amount",
		UnharvestedRewardsAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "non-negative-historical-rewards",
		NonNegativeHistoricalRewardsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "positive-total-stakings-amount",
		PositiveTotalStakingsAmountInvariant(k))
}

// AllInvariants runs all invariants of the farming module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []func(Keeper) sdk.Invariant{
			PositiveStakingAmountInvariant,
			StakingReservedAmountInvariant,
			RemainingRewardsAmountInvariant,
			NonNegativeOutstandingRewardsInvariant,
			OutstandingRewardsAmountInvariant,
			UnharvestedRewardsAmountInvariant,
			NonNegativeHistoricalRewardsInvariant,
			PositiveTotalStakingsAmountInvariant,
		} {
			res, stop := inv(k)(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

// PositiveStakingAmountInvariant checks that the amount of staking coins is positive.
func PositiveStakingAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		k.IterateStakings(ctx, func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool) {
			if !staking.Amount.IsPositive() {
				msg += fmt.Sprintf("\t%v has non-positive staking amount: %v\n",
					farmerAcc, sdk.Coin{Denom: stakingCoinDenom, Amount: staking.Amount})
				count++
			}
			return false
		})
		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "positive staking amount",
			fmt.Sprintf("found %d staking coins with non-positive amount\n%s", count, msg),
		), broken
	}
}

// PositiveQueuedStakingAmountInvariant checks that the amount of queued staking coins is positive.
func PositiveQueuedStakingAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		k.IterateQueuedStakings(ctx, func(_ time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
			if !queuedStaking.Amount.IsPositive() {
				msg += fmt.Sprintf("\t%v has non-positive queued staking amount: %v\n",
					farmerAcc, sdk.Coin{Denom: stakingCoinDenom, Amount: queuedStaking.Amount})
				count++
			}
			return false
		})
		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "positive queued staking amount",
			fmt.Sprintf("found %d queued staking coins with non-positive amount\n%s", count, msg),
		), broken
	}
}

// StakingReservedAmountInvariant checks that the balance of StakingReserveAcc greater than the amount of staked, Queued coins in all staking objects.
func StakingReservedAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateStakingReservedAmount(ctx)
		broken := err != nil
		return sdk.FormatInvariant(types.ModuleName, "staking reserved amount",
			"the balance of StakingReserveAcc less than the amount of staked, queued coins in all staking objects",
		), broken
	}
}

// RemainingRewardsAmountInvariant checks that the balance of the RewardsReserveAcc of all plans greater than the total amount of unwithdrawn reward coins in all reward objects
func RemainingRewardsAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.ValidateRemainingRewardsAmount(ctx)
		broken := err != nil
		return sdk.FormatInvariant(types.ModuleName, "remaining rewards amount",
			"the balance of the RewardsReserveAcc of all plans less than the total amount of unwithdrawn reward coins in all reward objects",
		), broken
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
		balances := k.bankKeeper.SpendableCoins(ctx, types.RewardsReserveAcc)
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

// UnharvestedRewardsAmountInvariant checks that UnharvestedRewards are
// consistent with rewards that can be withdrawn.
func UnharvestedRewardsAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		totalRewards := sdk.Coins{}
		k.IterateAllUnharvestedRewards(ctx, func(_ sdk.AccAddress, _ string, rewards types.UnharvestedRewards) (stop bool) {
			totalRewards = totalRewards.Add(rewards.Rewards...)
			return false
		})
		balances := k.bankKeeper.SpendableCoins(ctx, types.UnharvestedRewardsReserveAcc)
		_, hasNeg := balances.SafeSub(totalRewards)
		broken := hasNeg
		return sdk.FormatInvariant(
			types.ModuleName, "wrong unharvested rewards amount",
			fmt.Sprintf("balances of unharvested rewards reserve account is less than total unharvested rewards\n"+
				"\texpected minimum balances: %s\n"+
				"\tactual balances: %s", totalRewards, balances,
			),
		), broken
	}
}

// NonNegativeHistoricalRewardsInvariant checks that all HistoricalRewards are
// non-negative.
func NonNegativeHistoricalRewardsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		k.IterateHistoricalRewards(ctx, func(stakingCoinDenom string, epoch uint64, rewards types.HistoricalRewards) (stop bool) {
			if rewards.CumulativeUnitRewards.IsAnyNegative() {
				msg += fmt.Sprintf("\t%v has negative historical rewards at epoch %d: %v\n",
					stakingCoinDenom, epoch, rewards.CumulativeUnitRewards)
				count++
			}
			return false
		})
		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "non-negative historical rewards",
			fmt.Sprintf("found %d staking coin with negative historical rewards\n%s", count, msg),
		), broken
	}
}

// PositiveTotalStakingsAmountInvariant checks that all TotalStakings
// have positive amount.
func PositiveTotalStakingsAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		count := 0
		k.IterateTotalStakings(ctx, func(stakingCoinDenom string, totalStakings types.TotalStakings) (stop bool) {
			if !totalStakings.Amount.IsPositive() {
				msg += fmt.Sprintf("\t%v has non-positive total staking amount: %v\n",
					stakingCoinDenom, totalStakings.Amount)
				count++
			}
			return false
		})
		broken := count != 0
		return sdk.FormatInvariant(
			types.ModuleName, "positive total stakings amount",
			fmt.Sprintf("found %d total stakings with non-positive amount\n%s", count, msg),
		), broken
	}
}
