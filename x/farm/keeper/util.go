package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

func (k Keeper) PoolRewardWeight(ctx sdk.Context, pool liquiditytypes.Pool, pair liquiditytypes.Pair) sdk.Dec {
	// TODO: further optimize gas usage by using BankKeeper.SpendableCoin()
	spendable := k.bankKeeper.SpendableCoins(ctx, pool.GetReserveAddress())
	rx := spendable.AmountOf(pair.QuoteCoinDenom)
	ry := spendable.AmountOf(pair.BaseCoinDenom)
	// TODO: use sqrt(k) instead of poolReserveAmtByX and ampFactor
	poolReserveAmtByX := rx.Add(pair.LastPrice.MulInt(ry).TruncateInt())
	var ampFactor sdk.Dec
	oneDec := sdk.OneDec()
	switch pool.Type {
	case liquiditytypes.PoolTypeBasic:
		ampFactor = oneDec
	case liquiditytypes.PoolTypeRanged:
		lastPrice := *pair.LastPrice
		minPrice := *pool.MinPrice
		maxPrice := *pool.MaxPrice
		if lastPrice.LT(minPrice) || lastPrice.GT(maxPrice) {
			ampFactor = sdk.ZeroDec()
			break
		}
		sqrt := utils.DecApproxSqrt
		ampFactor = oneDec.Quo(
			oneDec.Sub(sqrt(minPrice.Quo(lastPrice)).Add(sqrt(lastPrice.Quo(maxPrice))).QuoInt64(2)))
	default:
		panic("invalid pool type")
	}
	return ampFactor.MulInt(poolReserveAmtByX)
}
