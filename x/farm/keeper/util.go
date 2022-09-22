package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// CachedKeeper provides cached getters for gas optimization.
// TODO: use generic with go1.18
type CachedKeeper struct {
	k                   Keeper
	pairCache           map[uint64]*liquiditytypes.Pair
	farmCache           map[string]*types.Farm
	spendableCoinsCache map[string]sdk.Coins
}

func NewCachedKeeper(k Keeper) *CachedKeeper {
	return &CachedKeeper{
		k:                   k,
		pairCache:           map[uint64]*liquiditytypes.Pair{},
		farmCache:           map[string]*types.Farm{},
		spendableCoinsCache: map[string]sdk.Coins{},
	}
}

func (cache *CachedKeeper) GetPair(ctx sdk.Context, id uint64) (pair liquiditytypes.Pair, found bool) {
	p, ok := cache.pairCache[id]
	if !ok {
		pair, found = cache.k.liquidityKeeper.GetPair(ctx, id)
		if found {
			p = &pair
		}
		cache.pairCache[id] = p
	}
	if p == nil {
		return liquiditytypes.Pair{}, false
	}
	return *p, true
}

func (cache *CachedKeeper) GetFarm(ctx sdk.Context, denom string) (farm types.Farm, found bool) {
	p, ok := cache.farmCache[denom]
	if !ok {
		farm, found = cache.k.GetFarm(ctx, denom)
		if found {
			p = &farm
		}
		cache.farmCache[denom] = p
	}
	if p == nil {
		return types.Farm{}, false
	}
	return *p, true
}

func (cache *CachedKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	addrStr := addr.String()
	balances, ok := cache.spendableCoinsCache[addrStr]
	if !ok {
		balances = cache.k.bankKeeper.SpendableCoins(ctx, addr)
		cache.spendableCoinsCache[addrStr] = balances
	}
	return balances
}

// PoolRewardWeight returns the pool's reward weight.
// TODO: check if last price is in price range for ranged pools
func (k Keeper) PoolRewardWeight(ctx sdk.Context, pool liquiditytypes.Pool, pair liquiditytypes.Pair) sdk.Dec {
	// TODO: further optimize gas usage by using BankKeeper.SpendableCoin()
	spendable := k.bankKeeper.SpendableCoins(ctx, pool.GetReserveAddress())
	rx := spendable.AmountOf(pair.QuoteCoinDenom)
	ry := spendable.AmountOf(pair.BaseCoinDenom)
	return types.PoolRewardWeight(pool.AMMPool(rx, ry, sdk.Int{}))
}
