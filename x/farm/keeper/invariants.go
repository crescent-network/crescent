package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "reference-count", ReferenceCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "current-rewards", OutstandingRewardsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "can-withdraw", CanWithdrawInvariant(k))
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
		return CanWithdrawInvariant(k)(ctx)
	}
}

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
