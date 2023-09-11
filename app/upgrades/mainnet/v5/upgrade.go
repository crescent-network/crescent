package v5

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	utils "github.com/crescent-network/crescent/v5/types"
	ammkeeper "github.com/crescent-network/crescent/v5/x/amm/keeper"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	claimkeeper "github.com/crescent-network/crescent/v5/x/claim/keeper"
	claimtypes "github.com/crescent-network/crescent/v5/x/claim/types"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
	farmingkeeper "github.com/crescent-network/crescent/v5/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v5/x/farming/types"
	liquidammtypes "github.com/crescent-network/crescent/v5/x/liquidamm/types"
	liquiditykeeper "github.com/crescent-network/crescent/v5/x/liquidity/keeper"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmkeeper "github.com/crescent-network/crescent/v5/x/lpfarm/keeper"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
	markerkeeper "github.com/crescent-network/crescent/v5/x/marker/keeper"
	markertypes "github.com/crescent-network/crescent/v5/x/marker/types"
)

const UpgradeName = "v5"

var StoreUpgrades = store.StoreUpgrades{
	Added: []string{
		markertypes.StoreKey,
		exchangetypes.StoreKey,
		ammtypes.StoreKey,
		liquidammtypes.StoreKey,
	},
}

func UpgradeHandler(
	mm *module.Manager, configurator module.Configurator, accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper, distrKeeper distrkeeper.Keeper, liquidityKeeper liquiditykeeper.Keeper,
	lpFarmKeeper lpfarmkeeper.Keeper, exchangeKeeper exchangekeeper.Keeper, ammKeeper ammkeeper.Keeper,
	markerKeeper markerkeeper.Keeper, farmingKeeper farmingkeeper.Keeper, claimKeeper claimkeeper.Keeper,
	disableUpgradeEvents bool) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		if disableUpgradeEvents {
			ctx = ctx.WithEventManager(sdk.NewEventManager())
		}

		// Set new module parameters.
		markerKeeper.SetParams(ctx, markertypes.DefaultParams())
		exchangeParams := exchangetypes.DefaultParams()
		exchangeParams.MarketCreationFee = sdk.NewCoins(sdk.NewInt64Coin("ucre", 2000_000000))
		exchangeParams.Fees = exchangetypes.NewFees(
			sdk.NewDecWithPrec(1, 3), // Maker: 0.1%
			sdk.NewDecWithPrec(2, 3), // Taker: 0.2%
			sdk.NewDecWithPrec(8, 1)) // Order source: Taker * 80%
		exchangeKeeper.SetParams(ctx, exchangeParams)
		ammParams := ammtypes.DefaultParams()
		ammParams.PoolCreationFee = sdk.NewCoins(sdk.NewInt64Coin("ucre", 1000_000000))
		ammParams.DefaultMinOrderQuantity = sdk.NewDec(10000)
		ammParams.DefaultMinOrderQuote = sdk.NewDec(10000)
		ammParams.PrivateFarmingPlanCreationFee = sdk.NewCoins(sdk.NewInt64Coin("ucre", 1000_000000))
		ammKeeper.SetParams(ctx, ammParams)

		// Migrate farming plans and staked coins to the new amm module.

		// Unstake all staked coins from x/farming and start farming on x/lpfarm.
		// NOTE: This code is taken from v3 upgrade handler.
		// First we migrate x/farming's staking positions to x/lpfarm's positions
		// and migrate the position again to x/amm.
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
				if _, err := lpFarmKeeper.Farm(ctx, farmerAddr, stakedCoin); err != nil {
					return nil, err
				}
			}
		}
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, farmingtypes.RewardsReserveAcc); err != nil {
			return nil, err
		}
		feeCollectorAddr := sdk.MustAccAddressFromBech32(farmingKeeper.GetParams(ctx).FarmingFeeCollector)
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, feeCollectorAddr); err != nil {
			return nil, err
		}

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
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, lpfarmtypes.RewardsPoolAddress); err != nil {
			return nil, err
		}
		feeCollectorAddr = sdk.MustAccAddressFromBech32(lpFarmKeeper.GetFeeCollector(ctx))
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, feeCollectorAddr); err != nil {
			return nil, err
		}

		// Migrate pairs to markets.
		// To avoid confusions, we use the same id as the pair id for markets.
		// Note that the new pool ids differ from the old ones.
		liquidityKeeper.DeleteOutdatedRequests(ctx)
		defaultMakerFeeRate := exchangeParams.Fees.DefaultMakerFeeRate
		defaultTakerFeeRate := exchangeParams.Fees.DefaultTakerFeeRate
		defaultOrderSourceFeeRatio := exchangeParams.Fees.DefaultOrderSourceFeeRatio
		pairs := map[uint64]liquiditytypes.Pair{}
		var pairIds []uint64 // For ordered map access
		var lastMarketId uint64
		if err := liquidityKeeper.IterateAllPairs(ctx, func(pair liquiditytypes.Pair) (stop bool, err error) {
			// Cache pairs for later iteration.
			pairs[pair.Id] = pair
			pairIds = append(pairIds, pair.Id)

			// Cancel all the orders. Note that the x/upgrade begin blocker is
			// the first begin blocker to run, so orders that should be deleted in
			// the current block haven't been deleted yet.
			// So we ran liquidityKeeper.DeleteOutdatedRequests first to sure that
			// all remaining orders are safe to be cancelled.
			if err := liquidityKeeper.IterateOrdersByPair(ctx, pair.Id, func(order liquiditytypes.Order) (stop bool, err error) {
				if err := liquidityKeeper.CancelOrder(
					ctx, liquiditytypes.NewMsgCancelOrder(order.GetOrderer(), pair.Id, order.Id)); err != nil {
					return true, err
				}
				return false, nil
			}); err != nil {
				return true, err
			}

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
				defaultMakerFeeRate, defaultTakerFeeRate, defaultOrderSourceFeeRatio)
			exchangeKeeper.SetMarket(ctx, market)
			exchangeKeeper.SetMarketByDenomsIndex(ctx, market)
			exchangeKeeper.SetMarketState(ctx, market.Id, exchangetypes.NewMarketState(pair.LastPrice))
			lastMarketId = pair.Id
			return false, nil
		}); err != nil {
			return nil, err
		}
		exchangeKeeper.SetLastMarketId(ctx, lastMarketId)

		// Create a new pool for each market if the market's corresponding pair
		// had at least one active pool.
		defaultTickSpacing := ammParams.DefaultTickSpacing
		defaultMinOrderQty := ammParams.DefaultMinOrderQuantity
		defaultMinOrderQuote := ammParams.DefaultMinOrderQuote
		newPoolIdByPairId := map[uint64]uint64{}
		pairIdByOldPoolId := map[uint64]uint64{}
		oldPoolsById := map[uint64]liquiditytypes.Pool{}
		var oldPoolIds []uint64
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
				if _, ok := oldPoolsById[pool.Id]; !ok {
					oldPoolIds = append(oldPoolIds, pool.Id)
				}
				oldPoolsById[pool.Id] = pool
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
					newPoolId, marketId, pair.BaseCoinDenom, pair.QuoteCoinDenom, defaultTickSpacing,
					defaultMinOrderQty, defaultMinOrderQuote)
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

		// Migrate liquidity. We do this by withdrawing all liquidity
		// from old pools and add the liquidity to the newly created pools.
		migrateLiquidity := func() error {
			accsBalances := bankKeeper.GetAccountsBalances(ctx)
			for _, accBalances := range accsBalances {
				addr := accBalances.GetAddress()
				// Skip if the account is potentially a module account.
				acc := accountKeeper.GetAccount(ctx, addr)
				if acc.GetSequence() == 0 || acc.GetPubKey() == nil {
					for _, coin := range accBalances.Coins {
						_, err := liquiditytypes.ParsePoolCoinDenom(coin.Denom)
						if err != nil { // Skip
							continue
						}
					}
					continue
				}
				for _, coin := range accBalances.Coins {
					oldPoolId, err := liquiditytypes.ParsePoolCoinDenom(coin.Denom)
					if err != nil { // Skip
						continue
					}

					// Withdraw from the old pool.
					// This is done by creating a new withdrawal request and
					// execute it immediately.
					// Within the same block, x/liquidity module's BeginBlocker will
					// delete all the requests stored in
					req, err := liquidityKeeper.Withdraw(
						ctx, liquiditytypes.NewMsgWithdraw(addr, oldPoolId, coin))
					if err != nil {
						return err
					}
					if err := liquidityKeeper.ExecuteWithdrawRequest(ctx, req); err != nil {
						return err
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

					oldPool := oldPoolsById[oldPoolId]
					newPoolId := newPoolIdByPairId[pairIdByOldPoolId[oldPoolId]]
					var lowerPrice, upperPrice sdk.Dec
					if oldPool.Type == liquiditytypes.PoolTypeBasic {
						lowerPrice = ammtypes.MinPrice
						upperPrice = ammtypes.MaxPrice
					} else { // the pool is a ranged pool
						lowerPrice = exchangetypes.PriceAtTick(
							ammtypes.AdjustPriceToTickSpacing(*oldPool.MinPrice, defaultTickSpacing, false))
						upperPrice = exchangetypes.PriceAtTick(
							ammtypes.AdjustPriceToTickSpacing(*oldPool.MaxPrice, defaultTickSpacing, true))
					}
					if _, _, _, err := ammKeeper.AddLiquidity(
						ctx, addr, addr, newPoolId, lowerPrice, upperPrice, req.WithdrawnCoins); err != nil {
						if errors.Is(err, sdkerrors.ErrInsufficientFunds) &&
							strings.Contains(err.Error(), "minted liquidity is zero") {
							// It is the case the withdrawn coins contains only one
							// denom, which happens because the withdrawal from the
							// old pool truncated the withdrawn coins.
							// Just skip this case and let the withdrawn coins be
							// in the account's balance.
							continue
						}
						return err
					}
				}
			}
			// Delete outdated requests again so that the withdrawal requests
			// we made above can be deleted.
			liquidityKeeper.DeleteOutdatedRequests(ctx)
			return nil
		}
		// We run migrateLiquidity() multiple times here.
		// This is because a withdrawal request with too small pool coin
		// could fail in the first run, but in the second run it
		// can be completed since the last withdrawer gets all remaining coins
		// in the pool's reserve, regardless how small the pool coin amount
		// was.
		for i := 0; i < 2; i++ {
			if err := migrateLiquidity(); err != nil {
				return nil, err
			}
		}

		// Send the remaining coins in each pool's reserve account to the
		// community pool and delete pools.
		for _, oldPoolId := range oldPoolIds {
			oldPool := oldPoolsById[oldPoolId]
			if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, oldPool.GetReserveAddress()); err != nil {
				return nil, err
			}
			// Delete the pool.
			liquidityKeeper.DeletePool(ctx, oldPoolId)
			liquidityKeeper.DeletePoolByReserveIndex(ctx, oldPool)
			liquidityKeeper.DeletePoolsByPairIndex(ctx, oldPool)
		}
		// Send the remaining coins in each pair's escrow account to the
		// community pool and delete pairs.
		for _, pairId := range pairIds {
			pair := pairs[pairId]
			if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, pair.GetEscrowAddress()); err != nil {
				return nil, err
			}
			liquidityKeeper.DeletePair(ctx, pairId)
			liquidityKeeper.DeletePairIndex(ctx, pair.BaseCoinDenom, pair.QuoteCoinDenom)
			liquidityKeeper.DeletePairLookupIndex(ctx, pair)
		}
		liquidityKeeper.DeleteLastPairId(ctx)
		liquidityKeeper.DeleteLastPoolId(ctx)

		// Send the remaining coins in the dust collector to the community
		// pool.
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, liquidityKeeper.GetDustCollector(ctx)); err != nil {
			return nil, err
		}
		if err := fundCommunityPool(ctx, bankKeeper, distrKeeper, liquidityKeeper.GetFeeCollector(ctx)); err != nil {
			return nil, err
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
			ammKeeper.SetFarmingPlan(ctx, ammtypes.FarmingPlan{
				Id:                 plan.Id,
				Description:        plan.Description,
				FarmingPoolAddress: plan.FarmingPoolAddress,
				TerminationAddress: plan.TerminationAddress,
				RewardAllocations:  newRewardAllocs,
				StartTime:          plan.StartTime,
				EndTime:            plan.EndTime,
				IsPrivate:          plan.IsPrivate,
				IsTerminated:       plan.IsTerminated,
			})
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
		lpFarmKeeper.IterateAllHistoricalRewards(ctx, func(denom string, period uint64, hist lpfarmtypes.HistoricalRewards) (stop bool) {
			lpFarmKeeper.DeleteHistoricalRewards(ctx, denom, period)
			return false
		})

		// Delete airdrops and claim records.
		var airdrops []claimtypes.Airdrop
		claimKeeper.IterateAllAirdrops(ctx, func(airdrop claimtypes.Airdrop) (stop bool) {
			airdrops = append(airdrops, airdrop)
			return false
		})
		for _, airdrop := range airdrops {
			claimKeeper.IterateAllClaimRecordsByAirdropId(ctx, airdrop.Id, func(record claimtypes.ClaimRecord) (stop bool) {
				claimKeeper.DeleteClaimRecord(ctx, record)
				return false
			})
			claimKeeper.DeleteAirdrop(ctx, airdrop.Id)
		}

		// TODO: remove temporary checks below before mainnet upgrade
		// No legacy pairs.
		ok := true
		_ = liquidityKeeper.IterateAllPairs(ctx, func(pair liquiditytypes.Pair) (stop bool, err error) {
			ok = false
			return true, nil
		})
		if !ok {
			panic("legacy pair exists")
		}
		if liquidityKeeper.GetLastPairId(ctx) != 0 {
			panic("legacy last pair id exists")
		}
		// No legacy pools.
		ok = true
		_ = liquidityKeeper.IterateAllPools(ctx, func(pool liquiditytypes.Pool) (stop bool, err error) {
			ok = false
			return true, nil
		})
		if !ok {
			panic("legacy pool exists")
		}
		if liquidityKeeper.GetLastPoolId(ctx) != 0 {
			panic("legacy last pool id exists")
		}
		// No legacy deposit/withdraw/order requests.
		ok = true
		_ = liquidityKeeper.IterateAllDepositRequests(ctx, func(req liquiditytypes.DepositRequest) (stop bool, err error) {
			ok = false
			return true, nil
		})
		if !ok {
			panic("legacy deposit request exists")
		}
		ok = true
		_ = liquidityKeeper.IterateAllWithdrawRequests(ctx, func(req liquiditytypes.WithdrawRequest) (stop bool, err error) {
			ok = false
			return true, nil
		})
		if !ok {
			panic("legacy withdraw request exists")
		}
		ok = true
		_ = liquidityKeeper.IterateAllOrders(ctx, func(order liquiditytypes.Order) (stop bool, err error) {
			ok = false
			return true, nil
		})
		if !ok {
			panic("legacy order exists")
		}
		// No legacy pool coin supply after the upgrade.
		ok = true
		bankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
			if _, err := liquiditytypes.ParsePoolCoinDenom(coin.Denom); err == nil {
				ok = false
				return true
			}
			return false
		})
		if !ok {
			panic("has pool coin supply")
		}

		// Similar farming rewards accrued after the upgrade.
		// No legacy farming plans.
		ok = true
		lpFarmKeeper.IterateAllPlans(ctx, func(plan lpfarmtypes.Plan) (stop bool) {
			ok = false
			return true
		})
		if !ok {
			panic("legacy farming plan exists")
		}
		// No legacy farms.
		ok = true
		lpFarmKeeper.IterateAllFarms(ctx, func(denom string, farm lpfarmtypes.Farm) (stop bool) {
			ok = false
			return true
		})
		if !ok {
			panic("legacy farm exists")
		}
		// No legacy lpfarm positions.
		ok = true
		lpFarmKeeper.IterateAllPositions(ctx, func(position lpfarmtypes.Position) (stop bool) {
			ok = false
			return true
		})
		if !ok {
			panic("legacy position exists")
		}
		// No legacy historical rewards.
		ok = true
		lpFarmKeeper.IterateAllHistoricalRewards(ctx, func(denom string, period uint64, hist lpfarmtypes.HistoricalRewards) (stop bool) {
			ok = false
			return true
		})
		if !ok {
			panic("legacy historical rewards exists")
		}

		// Set default market/pool parameters for the upgrade.
		changedPairIds := maps.Keys(ParamChanges)
		slices.Sort(changedPairIds)
		for _, pairId := range changedPairIds {
			if ParamChanges[pairId].MakerFeeRate != nil || ParamChanges[pairId].TakerFeeRate != nil {
				market, found := exchangeKeeper.GetMarket(ctx, pairId) // marketId == pairId
				if !found {                                            // maybe in test
					continue
				}
				market.MakerFeeRate = *ParamChanges[pairId].MakerFeeRate
				market.TakerFeeRate = *ParamChanges[pairId].TakerFeeRate
				exchangeKeeper.SetMarket(ctx, market)
			}
			if ParamChanges[pairId].TickSpacing != nil {
				poolId, ok := newPoolIdByPairId[pairId]
				if !ok { // maybe in test
					continue
				}
				pool := ammKeeper.MustGetPool(ctx, poolId)
				pool.TickSpacing = *ParamChanges[pairId].TickSpacing
				ammKeeper.SetPool(ctx, pool)
			}
		}

		return vm, nil
	}
}

func fundCommunityPool(
	ctx sdk.Context, bankKeeper bankkeeper.Keeper, distrKeeper distrkeeper.Keeper, addr sdk.AccAddress) error {
	remainingCoins := bankKeeper.SpendableCoins(ctx, addr)
	if remainingCoins.IsAllPositive() {
		return distrKeeper.FundCommunityPool(ctx, remainingCoins, addr)
	}
	return nil
}
