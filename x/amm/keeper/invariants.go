package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "rewards-growth-global", RewardsGrowthGlobalInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rewards-growth-outside", RewardsGrowthOutsideInvariant(k))
	ir.RegisterRoute(types.ModuleName, "can-remove-liquidity", CanRemoveLiquidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "can-collect", CanCollectInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-current-liquidity", PoolCurrentLiquidityInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (res string, broken bool) {
		res, broken = RewardsGrowthGlobalInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = RewardsGrowthOutsideInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = CanRemoveLiquidityInvariant(k)(ctx)
		if broken {
			return
		}
		res, broken = CanCollectInvariant(k)(ctx)
		if broken {
			return
		}
		return PoolCurrentLiquidityInvariant(k)(ctx)
	}
}

func RewardsGrowthGlobalInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
			poolState := k.MustGetPoolState(ctx, pool.Id)
			if poolState.FeeGrowthGlobal.IsAnyNegative() {
				msg += fmt.Sprintf(
					"\tpool %d has negative fee growth global: %s\n",
					pool.Id, poolState.FeeGrowthGlobal)
				cnt++
			}
			if poolState.FarmingRewardsGrowthGlobal.IsAnyNegative() {
				msg += fmt.Sprintf(
					"\tpool %d has negative farming rewards growth global: %s\n",
					pool.Id, poolState.FarmingRewardsGrowthGlobal)
				cnt++
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "rewards growth global",
			fmt.Sprintf(
				"found %d pool(s) with wrong fee growth or farming rewards growth global\n%s",
				cnt, msg)), broken
	}
}

func RewardsGrowthOutsideInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		pools := k.GetAllPools(ctx)
		for _, pool := range pools {
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
		}
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "rewards growth outside",
			fmt.Sprintf(
				"found %d tick info(s) with wrong fee growth or farming rewards growth outside\n%s",
				cnt, msg)), broken
	}
}

func CanRemoveLiquidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		cnt := 0
		pools := k.GetAllPools(ctx)
		for _, pool := range pools {
			expectedBalances := sdk.Coins{}
			k.IteratePositionsByPool(ctx, pool.Id, func(position types.Position) (stop bool) {
				coin0, coin1, err := k.PositionAssets(ctx, position.Id)
				if err != nil {
					panic(err) // XXX
				}
				expectedBalances = expectedBalances.Add(coin0, coin1)
				return false
			})
			balances := k.bankKeeper.SpendableCoins(ctx, pool.MustGetReserveAddress())
			fmt.Println(balances, expectedBalances)
			if !balances.IsAllGTE(expectedBalances) {
				msg += fmt.Sprintf(
					"\tpool %d has %s in reserve which is smaller than expected %s\n",
					pool.Id, balances, expectedBalances)
				cnt++
			}
		}
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "can remove liquidity",
			fmt.Sprintf("found %d pool(s) with insufficient reserve balances\n%s", cnt, msg)), broken
	}
}

func CanCollectInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		ctx, _ = ctx.CacheContext()
		msg := ""
		cnt := 0
		k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
			defer func() {
				if r := recover(); r != nil {
					msg += fmt.Sprintf(
						"\tcannot collect rewards from position %d: %v\n", position.Id, r)
					cnt++
				}
			}()
			ownerAddr := position.MustGetOwnerAddress()
			fee, farmingRewards, err := k.CollectibleCoins(ctx, position.Id)
			if err != nil {
				msg += fmt.Sprintf(
					"\tcannot calculate collectible coins for position %d: %s\n", position.Id, err)
				cnt++
				return false
			}
			if err := k.Collect(ctx, ownerAddr, ownerAddr, position.Id, fee.Add(farmingRewards...)); err != nil {
				msg += fmt.Sprintf(
					"\tcannot collect rewards from position %d: %s\n", position.Id, err)
				cnt++
				return false
			}
			return false
		})
		broken := cnt != 0
		return sdk.FormatInvariant(
			types.ModuleName, "can collect",
			fmt.Sprintf("found %d invalid position state(s)\n%s", cnt, msg)), broken
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
					pool.Id, poolState.CurrentLiquidity, currentLiquidity)
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
