package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

// cachingKeeper acts like a proxy to keeper methods,
// but caches the result to avoid unnecessary gas consumptions and
// store read operations.
// TODO: use generic with go1.18
type cachingKeeper struct {
	k         Keeper
	pairCache map[uint64]*liquiditytypes.Pair
	poolCache map[uint64]*liquiditytypes.Pool
	farmCache map[string]*types.Farm
}

func newCachingKeeper(k Keeper) *cachingKeeper {
	return &cachingKeeper{
		k:         k,
		pairCache: map[uint64]*liquiditytypes.Pair{},
		poolCache: map[uint64]*liquiditytypes.Pool{},
		farmCache: map[string]*types.Farm{},
	}
}

func (ck *cachingKeeper) getPair(ctx sdk.Context, id uint64) (pair liquiditytypes.Pair, found bool) {
	p, ok := ck.pairCache[id]
	if !ok {
		pair, found = ck.k.liquidityKeeper.GetPair(ctx, id)
		if found {
			p = &pair
		}
		ck.pairCache[id] = p
	}
	if p == nil {
		return liquiditytypes.Pair{}, false
	}
	return *p, true
}

func (ck *cachingKeeper) getFarm(ctx sdk.Context, denom string) (farm types.Farm, found bool) {
	p, ok := ck.farmCache[denom]
	if !ok {
		farm, found = ck.k.GetFarm(ctx, denom)
		if found {
			p = &farm
		}
		ck.farmCache[denom] = p
	}
	if p == nil {
		return types.Farm{}, false
	}
	return *p, true
}

type rewardAllocator struct {
	ctx                       sdk.Context
	k                         Keeper
	ck                        *cachingKeeper
	allocatedRewards          map[string]map[string]sdk.DecCoins // farming pool => (denom => rewards)
	totalRewardsByFarmingPool map[string]sdk.Coins               // farming pool => total rewards
	farmingPoolAddrs          []sdk.AccAddress
	poolInfosByPairId         map[uint64][]*poolInfo
}

type poolInfo struct {
	poolCoinDenom string
	rewardWeight  sdk.Dec
	rewardsShare  sdk.Dec
}

func newRewardAllocator(ctx sdk.Context, k Keeper, ck *cachingKeeper) *rewardAllocator {
	return &rewardAllocator{
		ctx:                       ctx,
		k:                         k,
		ck:                        ck,
		allocatedRewards:          map[string]map[string]sdk.DecCoins{},
		totalRewardsByFarmingPool: map[string]sdk.Coins{},
		poolInfosByPairId:         map[uint64][]*poolInfo{},
	}
}

func (ra *rewardAllocator) allocateRewardsToPair(farmingPoolAddr sdk.AccAddress, pair liquiditytypes.Pair, rewards sdk.Coins) {
	poolInfos, ok := ra.poolInfosByPairId[pair.Id]
	if !ok {
		totalRewardWeight := sdk.ZeroDec()
		_ = ra.k.liquidityKeeper.IteratePoolsByPair(ra.ctx, pair.Id, func(pool liquiditytypes.Pool) (stop bool, err error) {
			if pool.Disabled {
				return false, nil
			}
			// If the pool is a ranged pool and its pair's last price is out of
			// its price range, skip the pool.
			// This is because the amplification factor would be zero
			// so its reward weight would eventually be zero, too.
			if pool.Type == liquiditytypes.PoolTypeRanged &&
				(pair.LastPrice.LT(*pool.MinPrice) || pair.LastPrice.GT(*pool.MaxPrice)) {
				return false, nil
			}
			farm, found := ra.ck.getFarm(ra.ctx, pool.PoolCoinDenom)
			if !found || !farm.TotalFarmingAmount.IsPositive() {
				return false, nil
			}
			rewardWeight := ra.k.PoolRewardWeight(ra.ctx, pool, pair)
			totalRewardWeight = totalRewardWeight.Add(rewardWeight)
			poolInfos = append(poolInfos, &poolInfo{
				poolCoinDenom: pool.PoolCoinDenom,
				rewardWeight:  rewardWeight,
			})
			return false, nil
		})
		for _, pi := range poolInfos {
			pi.rewardsShare = pi.rewardWeight.QuoTruncate(totalRewardWeight)
		}
		ra.poolInfosByPairId[pair.Id] = poolInfos
	}
	if len(poolInfos) > 0 {
		farmingPool := farmingPoolAddr.String()
		rewardsByDenom, ok := ra.allocatedRewards[farmingPool]
		if !ok {
			rewardsByDenom = map[string]sdk.DecCoins{}
			ra.allocatedRewards[farmingPool] = rewardsByDenom
			ra.farmingPoolAddrs = append(ra.farmingPoolAddrs, farmingPoolAddr)
		}
		for _, pi := range poolInfos {
			rewardsByDenom[pi.poolCoinDenom] = rewardsByDenom[pi.poolCoinDenom].Add(
				sdk.NewDecCoinsFromCoins(rewards...).MulDecTruncate(pi.rewardsShare)...)
		}
		ra.totalRewardsByFarmingPool[farmingPool] =
			ra.totalRewardsByFarmingPool[farmingPool].Add(rewards...)
	}
}

func (ra *rewardAllocator) allocateRewardsToDenom(farmingPoolAddr sdk.AccAddress, denom string, rewards sdk.Coins) {
	farm, found := ra.ck.getFarm(ra.ctx, denom)
	if !found || !farm.TotalFarmingAmount.IsPositive() {
		return
	}
	farmingPool := farmingPoolAddr.String()
	rewardsByDenom, ok := ra.allocatedRewards[farmingPool]
	if !ok {
		rewardsByDenom = map[string]sdk.DecCoins{}
		ra.allocatedRewards[farmingPool] = rewardsByDenom
		ra.farmingPoolAddrs = append(ra.farmingPoolAddrs, farmingPoolAddr)
	}
	rewardsByDenom[denom] = rewardsByDenom[denom].Add(
		sdk.NewDecCoinsFromCoins(rewards...)...)
	ra.totalRewardsByFarmingPool[farmingPool] =
		ra.totalRewardsByFarmingPool[farmingPool].Add(rewards...)
}

// PoolRewardWeight returns the pool's reward weight.
func (k Keeper) PoolRewardWeight(ctx sdk.Context, pool liquiditytypes.Pool, pair liquiditytypes.Pair) sdk.Dec {
	if pool.Type == liquiditytypes.PoolTypeRanged &&
		(pair.LastPrice.LT(*pool.MinPrice) || pair.LastPrice.GT(*pool.MaxPrice)) {
		return sdk.ZeroDec()
	}
	// TODO: further optimize gas usage by using BankKeeper.SpendableCoin()
	spendable := k.bankKeeper.SpendableCoins(ctx, pool.GetReserveAddress())
	rx := spendable.AmountOf(pair.QuoteCoinDenom)
	ry := spendable.AmountOf(pair.BaseCoinDenom)
	return types.PoolRewardWeight(pool.AMMPool(rx, ry, sdk.Int{}))
}
