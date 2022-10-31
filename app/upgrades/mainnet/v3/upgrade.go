package v3

import (
	"time"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	farmingkeeper "github.com/crescent-network/crescent/v3/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v3/x/farming/types"
	liquidfarmingtypes "github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditykeeper "github.com/crescent-network/crescent/v3/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	lpfarmkeeper "github.com/crescent-network/crescent/v3/x/lpfarm/keeper"
	lpfarmtypes "github.com/crescent-network/crescent/v3/x/lpfarm/types"
	marketmakerkeeper "github.com/crescent-network/crescent/v3/x/marketmaker/keeper"
	marketmakertypes "github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

const UpgradeName = "v3"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, marketmakerKeeper marketmakerkeeper.Keeper,
	liquidityKeeper liquiditykeeper.Keeper, lpfarmKeeper lpfarmkeeper.Keeper, farmingKeeper farmingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper) upgradetypes.UpgradeHandler {
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

		lpfarmKeeper.SetPrivatePlanCreationFee(
			ctx, sdk.NewCoins(sdk.NewInt64Coin("ucre", 100_000000)))
		// Move fees collected in the farming module's fee collector.
		farmingFeeCollector, _ := sdk.AccAddressFromBech32(farmingKeeper.GetParams(ctx).FarmingFeeCollector)
		farmingFees := bankKeeper.SpendableCoins(ctx, farmingFeeCollector)
		if farmingFees.IsAllPositive() {
			farmFeeCollector, _ := sdk.AccAddressFromBech32(lpfarmKeeper.GetFeeCollector(ctx))
			if err := bankKeeper.SendCoins(
				ctx, farmingFeeCollector, farmFeeCollector, farmingFees); err != nil {
				return nil, err
			}
		}

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
				return false
			})
		farmingKeeper.IterateQueuedStakings(
			ctx, func(_ time.Time, denom string, farmerAddr sdk.AccAddress, queuedStaking farmingtypes.QueuedStaking) (stop bool) {
				farmer := farmerAddr.String()
				if _, ok := stakedCoinsByFarmer[farmer]; !ok {
					farmerAddrs = append(farmerAddrs, farmerAddr)
				}
				stakedCoinsByFarmer[farmer] = stakedCoinsByFarmer[farmer].
					Add(sdk.NewCoin(denom, queuedStaking.Amount))
				return false
			})
		for _, farmerAddr := range farmerAddrs {
			if err := farmingKeeper.Unstake(ctx, farmerAddr, stakedCoinsByFarmer[farmerAddr.String()]); err != nil {
				return nil, err
			}
			for _, stakedCoin := range stakedCoinsByFarmer[farmerAddr.String()] {
				if _, err := lpfarmKeeper.Farm(ctx, farmerAddr, stakedCoin); err != nil {
					return nil, err
				}
			}
		}

		var lastPlanId, numPrivatePlans uint64
		farmingKeeper.IteratePlans(ctx, func(plan farmingtypes.PlanI) (stop bool) {
			epochAmt := sdk.NewDecCoinsFromCoins(plan.(*farmingtypes.FixedAmountPlan).EpochAmount...)
			var rewardAllocs []lpfarmtypes.RewardAllocation
			for _, weight := range plan.GetStakingCoinWeights() {
				rewardsPerDay, _ := epochAmt.MulDecTruncate(weight.Amount).TruncateDecimal()
				rewardAllocs = append(
					rewardAllocs, lpfarmtypes.NewDenomRewardAllocation(
						weight.Denom, rewardsPerDay))
			}
			lpfarmKeeper.SetPlan(ctx, lpfarmtypes.Plan{
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
		lpfarmKeeper.SetLastPlanId(ctx, lastPlanId)
		lpfarmKeeper.SetNumPrivatePlans(ctx, numPrivatePlans)
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
		lpfarmtypes.ModuleName,
		liquidfarmingtypes.ModuleName,
	},
}
