package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "rewards-growth", RewardsGrowthInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-current-liquidity", PoolCurrentLiquidityInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		res, broken = RewardsGrowthInvariant(k)(ctx)
		if broken {
			return
		}
		return PoolCurrentLiquidityInvariant(k)(ctx)
	}
}

func RewardsGrowthInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
			poolState := k.MustGetPoolState(ctx, pool.Id)
			k.IterateTickInfosByPool(ctx, pool.Id, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if _, hasNeg := poolState.FeeGrowthGlobal.SafeSub(tickInfo.FeeGrowthOutside); hasNeg {
					msg += fmt.Sprintf(
						"\tpool %d has tick info with wrong fee growth outside at %d: %s > %s\n",
						pool.Id, tick, tickInfo.FeeGrowthOutside, poolState.FeeGrowthGlobal)
					cnt++
				}
				if _, hasNeg := poolState.FarmingRewardsGrowthGlobal.SafeSub(tickInfo.FarmingRewardsGrowthOutside); hasNeg {
					msg += fmt.Sprintf(
						"\tpool %d has tick info with bad wrongfarming rewards growth outside at %d: %s > %s\n",
						pool.Id, tick, tickInfo.FarmingRewardsGrowthOutside, poolState.FarmingRewardsGrowthGlobal)
					cnt++
				}
				return false
			})
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "rewards growth",
			fmt.Sprintf(
				"found %d tick info(s) with wrong fee growth or farming rewards growth\n%s",
				cnt, msg)), broken
	}
}

func PoolCurrentLiquidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		var pools []types.Pool
		k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
			pools = append(pools, pool)
			return false
		})
		for _, pool := range pools {
			poolState := k.MustGetPoolState(ctx, pool.Id)
			currentLiquidity := utils.ZeroInt
			k.IterateTickInfosByPool(ctx, pool.Id, func(tick int32, tickInfo types.TickInfo) (stop bool) {
				if tick > poolState.CurrentTick {
					return true
				}
				currentLiquidity = currentLiquidity.Add(tickInfo.NetLiquidity)
				return false
			})
			if !poolState.CurrentLiquidity.Equal(currentLiquidity) {
				msg += fmt.Sprintf(
					"\tpool %d has wrong current liquidity: %s != %s\n",
					pool.Id, poolState.CurrentPrice, currentLiquidity)
				cnt++
			}
		}
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "pool current liquidity",
			fmt.Sprintf(
				"found %d pool(s) with wrong current liquidity\n%s",
				cnt, msg)), broken
	}
}
