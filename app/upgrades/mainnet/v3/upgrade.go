package v3

import (
	"time"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	farmkeeper "github.com/crescent-network/crescent/v3/x/farm/keeper"
	farmtypes "github.com/crescent-network/crescent/v3/x/farm/types"
	farmingkeeper "github.com/crescent-network/crescent/v3/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v3/x/farming/types"
	liquidfarmingtypes "github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditykeeper "github.com/crescent-network/crescent/v3/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	marketmakerkeeper "github.com/crescent-network/crescent/v3/x/marketmaker/keeper"
	marketmakertypes "github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

const UpgradeName = "v3"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, marketmakerKeeper marketmakerkeeper.Keeper,
	liquidityKeeper liquiditykeeper.Keeper, farmKeeper farmkeeper.Keeper, farmingKeeper farmingkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVM, err
		}

		// Set newly added liquidity param
		liquidityKeeper.SetMaxNumMarketMakingOrderTicks(ctx, liquiditytypes.DefaultMaxNumMarketMakingOrderTicks)

		// Set param for new market maker module
		marketmakerParams := marketmakertypes.DefaultParams()
		marketmakerParams.DepositAmount = sdk.NewCoins(sdk.NewCoin("ucre", sdk.NewInt(1000000000)))
		marketmakerKeeper.SetParams(ctx, marketmakerParams)

		// Unstake all staked coins from x/farming and start farming on x/farm.
		stakedCoinsByFarmer := map[string]sdk.Coins{}
		var farmerAddrs []sdk.AccAddress
		farmingKeeper.IterateStakings(
			ctx, func(denom string, farmerAddr sdk.AccAddress, staking farmingtypes.Staking) (stop bool) {
				farmer := farmerAddr.String()
				if _, ok := stakedCoinsByFarmer[farmer]; !ok {
					farmerAddrs = append(farmerAddrs, farmerAddr)
				}
				stakedCoinsByFarmer[farmer] = stakedCoinsByFarmer[farmer].
					Add(sdk.NewCoin(denom, staking.Amount))
				farmingKeeper.IterateQueuedStakingsByFarmer(
					ctx, farmerAddr, func(_ string, _ time.Time, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
						stakedCoinsByFarmer[farmer] = stakedCoinsByFarmer[farmer].
							Add(sdk.NewCoin(denom, queuedStaking.Amount))
						return false
					})
				return false
			})
		for _, farmerAddr := range farmerAddrs {
			if err := farmingKeeper.Unstake(ctx, farmerAddr, stakedCoinsByFarmer[farmerAddr.String()]); err != nil {
				return nil, err
			}
			for _, stakedCoin := range stakedCoinsByFarmer[farmerAddr.String()] {
				if _, err := farmKeeper.Farm(ctx, farmerAddr, stakedCoin); err != nil {
					return nil, err
				}
			}
		}

		var lastPlanId, numPrivatePlans uint64
		farmingKeeper.IteratePlans(ctx, func(plan farmingtypes.PlanI) (stop bool) {
			epochAmt := sdk.NewDecCoinsFromCoins(plan.(*farmingtypes.FixedAmountPlan).EpochAmount...)
			var rewardAllocs []farmtypes.RewardAllocation
			for _, weight := range plan.GetStakingCoinWeights() {
				rewardsPerDay, _ := epochAmt.MulDecTruncate(weight.Amount).TruncateDecimal()
				rewardAllocs = append(
					rewardAllocs, farmtypes.NewDenomRewardAllocation(
						weight.Denom, rewardsPerDay))
			}
			farmKeeper.SetPlan(ctx, farmtypes.Plan{
				Id:                 plan.GetId(),
				Description:        plan.GetName(),
				FarmingPoolAddress: plan.GetFarmingPoolAddress().String(),
				TerminationAddress: plan.GetTerminationAddress().String(),
				RewardAllocations:  rewardAllocs,
				StartTime:          plan.GetStartTime(),
				EndTime:            plan.GetEndTime(),
				IsPrivate:          plan.GetType() == farmingtypes.PlanTypePrivate,
				IsTerminated:       plan.IsTerminated(),
			})
			farmingKeeper.DeletePlan(ctx, plan)
			lastPlanId = plan.GetId()
			if plan.GetType() == farmingtypes.PlanTypePrivate && !plan.IsTerminated() {
				numPrivatePlans++
			}
			return false
		})
		farmKeeper.SetLastPlanId(ctx, lastPlanId)
		farmKeeper.SetNumPrivatePlans(ctx, numPrivatePlans)
		farmingKeeper.DeleteGlobalPlanId(ctx)
		farmingKeeper.DeleteLastEpochTime(ctx)
		farmingKeeper.DeleteCurrentEpochDays(ctx)
		farmingKeeper.IterateCurrentEpochs(ctx, func(stakingCoinDenom string, _ uint64) (stop bool) {
			farmingKeeper.DeleteCurrentEpoch(ctx, stakingCoinDenom)
			return false
		})

		return newVM, err

	}
}

// Add new modules
var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		marketmakertypes.ModuleName,
		farmtypes.ModuleName,
		liquidfarmingtypes.ModuleName,
	},
}
