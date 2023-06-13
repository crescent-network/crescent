package v5

import (
	"errors"
	"strings"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	utils "github.com/crescent-network/crescent/v5/types"
	ammkeeper "github.com/crescent-network/crescent/v5/x/amm/keeper"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	liquiditykeeper "github.com/crescent-network/crescent/v5/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmkeeper "github.com/crescent-network/crescent/v5/x/lpfarm/keeper"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
	markerkeeper "github.com/crescent-network/crescent/v5/x/marker/keeper"
	markertypes "github.com/crescent-network/crescent/v5/x/marker/types"
)

const UpgradeName = "v5"

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, bankKeeper bankkeeper.Keeper, distrKeeper distrkeeper.Keeper,
	liquidityKeeper liquiditykeeper.Keeper, lpFarmKeeper lpfarmkeeper.Keeper, exchangeKeeper exchangekeeper.Keeper,
	ammKeeper ammkeeper.Keeper, markerKeeper markerkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set new module parameters.
		markerKeeper.SetParams(ctx, markertypes.DefaultParams())
		exchangeParams := exchangetypes.DefaultParams()
		exchangeKeeper.SetParams(ctx, exchangeParams)
		ammParams := ammtypes.DefaultParams()
		ammKeeper.SetParams(ctx, ammParams)

		// Migrate farming plans and staked coins to the new amm module.
		// TODO: disable event emissions?

		// Unfarm all coins from all farmers.
		// This will remove all farmers' positions and withdraw their rewards
		// from the rewards pool.
		lpFarmKeeper.IterateAllPositions(ctx, func(position lpfarmtypes.Position) (stop bool) {
			farmerAddr := sdk.MustAccAddressFromBech32(position.Farmer)
			coin := sdk.NewCoin(position.Denom, position.FarmingAmount)
			if _, err = lpFarmKeeper.Unfarm(ctx, farmerAddr, coin); err != nil {
				return true
			}
			return false
		})
		if err != nil {
			return nil, err
		}
		// Send remaining rewards(due to decimal truncation) in the rewards
		// pool to the community pool.
		remainingRewards := bankKeeper.SpendableCoins(ctx, lpfarmtypes.RewardsPoolAddress)
		if remainingRewards.IsAllPositive() {
			if err := distrKeeper.FundCommunityPool(ctx, remainingRewards, lpfarmtypes.RewardsPoolAddress); err != nil {
				return nil, err
			}
		}
		// TODO: send coins in the fee collector?

		// Migrate pairs to markets.
		// To avoid confusions, we use the same id as the pair id for markets.
		// Note that the new pool ids differ from the old ones.
		defaultMakerFeeRate := exchangeParams.Fees.DefaultMakerFeeRate
		defaultTakerFeeRate := exchangeParams.Fees.DefaultTakerFeeRate
		pairs := map[uint64]liquiditytypes.Pair{}
		var pairIds []uint64 // For ordered map access
		_ = liquidityKeeper.IterateAllPairs(ctx, func(pair liquiditytypes.Pair) (stop bool, err error) {
			// If there's a coin inside a pair which has no supply in bank module,
			// skip the pair.
			// Note, however, if the pool exists then we can safely assume that
			// a market with the same id with the pool's pair id exists.
			if !bankKeeper.HasSupply(ctx, pair.BaseCoinDenom) || !bankKeeper.HasSupply(ctx, pair.QuoteCoinDenom) {
				return false, nil
			}

			// Save a new market with the same id as the pair's id and set
			// corresponding indexes, too.
			market := exchangetypes.NewMarket(
				pair.Id, pair.BaseCoinDenom, pair.QuoteCoinDenom,
				defaultMakerFeeRate, defaultTakerFeeRate)
			exchangeKeeper.SetMarket(ctx, market)
			exchangeKeeper.SetMarketByDenomsIndex(ctx, market)
			exchangeKeeper.SetMarketState(ctx, market.Id, exchangetypes.NewMarketState(pair.LastPrice))

			// Cache pairs for later iteration.
			pairs[pair.Id] = pair
			pairIds = append(pairIds, pair.Id)
			return false, nil
		})

		// Create a new pool for each market if the market's corresponding pair
		// had at least one active pool.
		defaultTickSpacing := ammParams.DefaultTickSpacing
		newPoolIdByPairId := map[uint64]uint64{}
		pairIdByOldPoolId := map[uint64]uint64{}
		oldPools := map[uint64]liquiditytypes.Pool{}
		for _, pairId := range pairIds {
			// These variables are used to calculate the new pool's initial price.
			// If there's the last price in the pair, then the new pool uses
			// the pair's last price.
			// If there's no last price, then the new pool uses the average
			// price of all the old pools in the pair.
			sumActivePoolPrices := utils.ZeroDec
			numActivePools := 0

			pair := pairs[pairId]
			_ = liquidityKeeper.IteratePoolsByPair(ctx, pairId, func(pool liquiditytypes.Pool) (stop bool, err error) {
				if pool.Disabled { // Skip disabled pools.
					return false, nil
				}
				pairIdByOldPoolId[pool.Id] = pairId
				// Sum pool prices only when the pair has no last price.
				if pair.LastPrice == nil {
					rx, ry := liquidityKeeper.GetPoolBalances(ctx, pool)
					// It is safe to pass an empty sdk.Int to Price() method.
					sumActivePoolPrices = sumActivePoolPrices.Add(
						pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{}).Price())
				}
				numActivePools++
				oldPools[pool.Id] = pool
				return false, nil
			})
			if numActivePools > 0 {
				var newPoolPrice sdk.Dec
				if pair.LastPrice == nil {
					newPoolPrice = sumActivePoolPrices.QuoInt64(int64(numActivePools))
				} else {
					newPoolPrice = *pair.LastPrice
				}

				// Save the new pool. We generate pool ids starting from 1.
				newPoolId := ammKeeper.GetNextPoolIdWithUpdate(ctx)
				marketId := pairId // We used the same ids as pairs for markets.
				newPool := ammtypes.NewPool(
					newPoolId, marketId, pair.BaseCoinDenom, pair.QuoteCoinDenom, defaultTickSpacing)
				ammKeeper.SetPool(ctx, newPool)
				// Set corresponding indexes.
				ammKeeper.SetPoolByMarketIndex(ctx, newPool)
				ammKeeper.SetPoolByReserveAddressIndex(ctx, newPool)
				// Set initial pool state with the pool price we've calculated.
				newPoolState := ammtypes.NewPoolState(exchangetypes.TickAtPrice(newPoolPrice), newPoolPrice)
				ammKeeper.SetPoolState(ctx, newPoolId, newPoolState)
				newPoolIdByPairId[pairId] = newPoolId
			}
		}

		accsBalances := bankKeeper.GetAccountsBalances(ctx)
		// Migrate liquidity provisions. We do this by withdrawing all liquidity
		// from old pools and add the liquidity to the newly created pools.
		for _, accBalances := range accsBalances {
			for _, coin := range accBalances.Coins {
				var oldPoolId uint64
				oldPoolId, err = liquiditytypes.ParsePoolCoinDenom(coin.Denom)
				if err != nil { // Skip
					continue
				}

				// Withdraw from the old pool.
				// This is done by creating a new withdrawal request and
				// execute it immediately.
				// Within the same block, x/liquidity module's BeginBlocker will
				// delete all the requests stored in
				addr := accBalances.GetAddress()
				var req liquiditytypes.WithdrawRequest
				req, err = liquidityKeeper.Withdraw(
					ctx, liquiditytypes.NewMsgWithdraw(addr, oldPoolId, coin))
				if err != nil {
					return nil, err
				}
				if err = liquidityKeeper.ExecuteWithdrawRequest(ctx, req); err != nil {
					return nil, err
				}
				var found bool
				req, found = liquidityKeeper.GetWithdrawRequest(ctx, oldPoolId, req.Id)
				if !found { // sanity check
					panic("withdraw request not found")
				}
				// If no coins have been withdrawn, skip this account.
				if req.WithdrawnCoins.IsZero() {
					continue
				}

				oldPool := oldPools[oldPoolId]
				newPoolId := newPoolIdByPairId[pairIdByOldPoolId[oldPoolId]]
				var lowerPrice, upperPrice sdk.Dec
				if oldPool.Type == liquiditytypes.PoolTypeBasic {
					lowerPrice = exchangetypes.PriceAtTick(ammtypes.MinTick)
					upperPrice = exchangetypes.PriceAtTick(ammtypes.MaxTick)
				} else { // the pool is a ranged pool
					lowerPrice = AdjustPriceToTickSpacing(*oldPool.MinPrice, defaultTickSpacing, false)
					upperPrice = AdjustPriceToTickSpacing(*oldPool.MaxPrice, defaultTickSpacing, true)
				}
				_, _, _, err = ammKeeper.AddLiquidity(
					ctx, addr, addr, newPoolId, lowerPrice, upperPrice, req.WithdrawnCoins)
				if err != nil {
					if errors.Is(err, sdkerrors.ErrInsufficientFunds) &&
						strings.Contains(err.Error(), "minted liquidity is zero") {
						// It is the case the withdrawn coins contains only one
						// denom, which happens because the withdrawal from the
						// old pool truncated the withdrawn coins.
						// Just skip this case and let the withdrawn coins be
						// in the account's balance.
						continue
					}
					return nil, err
				}
			}
		}

		// Migrate farming plans.
		var lastFarmingPlanId uint64
		numActivePrivateFarmingPlans := uint32(0)
		lpFarmKeeper.IterateAllPlans(ctx, func(plan lpfarmtypes.Plan) (stop bool) {
			var newRewardAllocs []ammtypes.FarmingRewardAllocation
			for _, rewardAlloc := range plan.RewardAllocations {
				if rewardAlloc.PairId != 0 { // Migrate pair reward allocations only
					// It's possible the old reward allocation to have a pair's
					// id which has no related new pool.
					// It happens when the pair has no active(not disabled) old
					// pools.
					// So just skip the migration of those pair reward allocations.
					newPoolId, ok := newPoolIdByPairId[rewardAlloc.PairId]
					if !ok {
						continue
					}
					newRewardAllocs = append(newRewardAllocs,
						ammtypes.NewFarmingRewardAllocation(newPoolId, rewardAlloc.RewardsPerDay))
				}
			}
			// Save the new farming plan with the same id as the old plan.
			ammKeeper.SetFarmingPlan(ctx, ammtypes.NewFarmingPlan(
				plan.Id, plan.Description, plan.GetFarmingPoolAddress(), plan.GetTerminationAddress(),
				newRewardAllocs, plan.StartTime, plan.EndTime, plan.IsPrivate))
			if plan.IsPrivate && !plan.IsTerminated {
				numActivePrivateFarmingPlans++
			}
			lpFarmKeeper.DeletePlan(ctx, plan)
			lastFarmingPlanId = plan.Id
			return false
		})
		// Sets the global counter states.
		ammKeeper.SetLastFarmingPlanId(ctx, lastFarmingPlanId)
		ammKeeper.SetNumPrivateFarmingPlans(ctx, numActivePrivateFarmingPlans)
		// Clean lpfarm module states.
		lpFarmKeeper.DeleteLastPlanId(ctx)
		lpFarmKeeper.DeleteNumPrivatePlans(ctx)
		lpFarmKeeper.IterateAllFarms(ctx, func(denom string, farm lpfarmtypes.Farm) (stop bool) {
			lpFarmKeeper.DeleteFarm(ctx, denom)
			return false
		})

		return vm, nil
	}
}

var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		markertypes.StoreKey,
		exchangetypes.StoreKey,
		ammtypes.StoreKey,
	},
}
