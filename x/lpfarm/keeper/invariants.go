package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "reference-count", ReferenceCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "current-rewards", OutstandingRewardsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "can-withdraw", CanWithdrawInvariant(k))
	ir.RegisterRoute(types.ModuleName, "total-farming-amount", TotalFarmingAmountInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		res, broken = ReferenceCountInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = OutstandingRewardsInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = CanWithdrawInvariant(k)(ctx)
		if broken {
			return
		}
		return TotalFarmingAmountInvariant(k)(ctx)
	}
}

// ReferenceCountInvariant checks that the all historical rewards object has
// consistent reference counts with all farms and positions.
func ReferenceCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		numFarms := uint64(0)
		k.IterateAllFarms(ctx, func(denom string, farm types.Farm) (stop bool) {
			numFarms++
			return false
		})
		numPositions := uint64(0)
		k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
			numPositions++
			return false
		})
		expected := numFarms + numPositions
		got := uint64(0)
		k.IterateAllHistoricalRewards(ctx, func(denom string, period uint64, hist types.HistoricalRewards) (stop bool) {
			got += uint64(hist.ReferenceCount)
			return false
		})
		broken := got != expected
		return sdk.FormatInvariant(
			types.ModuleName, "reference count",
			fmt.Sprintf(
				"total reference count %d != expected %d(=%d farms+ %d positions)",
				got, expected, numFarms, numPositions,
			),
		), broken
	}
}

// OutstandingRewardsInvariant checks that the outstanding rewards of all
// farms are not smaller than the farm's current accrued rewards, and the reward
// pool has sufficient balances for those rewards.
func OutstandingRewardsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		outstanding := sdk.DecCoins{}
		msg := ""
		cnt := 0
		k.IterateAllFarms(ctx, func(denom string, farm types.Farm) (stop bool) {
			_, hasNeg := farm.OutstandingRewards.SafeSub(farm.CurrentRewards)
			if hasNeg {
				msg += fmt.Sprintf(
					"\tfarm %s has smaller outstanding rewards than current rewards: %s < %s\n",
					denom, farm.OutstandingRewards, farm.CurrentRewards)
				cnt++
			}
			outstanding = outstanding.Add(farm.CurrentRewards...)
			return false
		})
		if cnt != 0 {
			return sdk.FormatInvariant(
				types.ModuleName, "outstanding rewards",
				fmt.Sprintf(
					"found %d farm(s) with smaller outstanding rewards than current rewards\n%s",
					cnt, msg,
				),
			), true
		}
		balances := k.bankKeeper.SpendableCoins(ctx, types.RewardsPoolAddress)
		_, broken := sdk.NewDecCoinsFromCoins(balances...).SafeSub(outstanding)
		return sdk.FormatInvariant(
			types.ModuleName, "outstanding rewards",
			fmt.Sprintf(
				"rewards pool balances %s is smaller than expected %s",
				balances, outstanding,
			),
		), broken
	}
}

// CanWithdrawInvariant checks that all farmers can withdraw their rewards.
func CanWithdrawInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		defer func() {
			if r := recover(); r != nil {
				broken = true
				res = sdk.FormatInvariant(
					types.ModuleName, "can withdraw",
					"cannot withdraw due to negative outstanding rewards",
				)
			}
		}()

		ctx, _ = ctx.CacheContext()
		k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
			farmerAddr, _ := sdk.AccAddressFromBech32(position.Farmer)
			if _, err := k.Harvest(ctx, farmerAddr, position.Denom); err != nil {
				panic(err)
			}
			return false
		})
		return
	}
}

// TotalFarmingAmountInvariant checks that all farm's total farming amount are
// equal to the sum of all the positions' farming amount which belong to the farm.
func TotalFarmingAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		totalFarmingAmtByDenom := map[string]sdk.Int{}
		farmingAmtSumByDenom := map[string]sdk.Int{}
		k.IterateAllFarms(ctx, func(denom string, farm types.Farm) (stop bool) {
			totalFarmingAmtByDenom[denom] = farm.TotalFarmingAmount
			farmingAmtSumByDenom[denom] = sdk.ZeroInt()
			return false
		})
		k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
			farmingAmtSumByDenom[position.Denom] =
				farmingAmtSumByDenom[position.Denom].Add(position.FarmingAmount)
			return false
		})
		msg := ""
		cnt := 0
		for denom := range totalFarmingAmtByDenom {
			if !totalFarmingAmtByDenom[denom].Equal(farmingAmtSumByDenom[denom]) {
				msg += fmt.Sprintf(
					"\tfarm %s total farming amount %s != sum %s\n",
					denom, totalFarmingAmtByDenom[denom], farmingAmtSumByDenom[denom],
				)
				cnt++
			}
		}
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "total farming amount",
			fmt.Sprintf(
				"found %d farm(s) with wrong total farming amount\n%s",
				cnt, msg,
			),
		), broken
	}
}
