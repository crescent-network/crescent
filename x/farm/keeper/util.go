package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// PoolRewardWeight returns the pool's reward weight.
// TODO: check if last price is in price range for ranged pools
func (k Keeper) PoolRewardWeight(ctx sdk.Context, pool liquiditytypes.Pool, pair liquiditytypes.Pair) sdk.Dec {
	// TODO: further optimize gas usage by using BankKeeper.SpendableCoin()
	spendable := k.bankKeeper.SpendableCoins(ctx, pool.GetReserveAddress())
	rx := spendable.AmountOf(pair.QuoteCoinDenom)
	ry := spendable.AmountOf(pair.BaseCoinDenom)
	return types.PoolRewardWeight(pool.AMMPool(rx, ry, sdk.Int{}))
}
