package amm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OfferCoinAmount returns the minimum offer coin amount for
// given order direction, price and order amount.
func OfferCoinAmount(dir OrderDirection, price sdk.Dec, amt sdk.Int) sdk.Int {
	switch dir {
	case Buy:
		return price.MulInt(amt).Ceil().TruncateInt()
	case Sell:
		return amt
	default:
		panic(fmt.Sprintf("invalid order direction: %s", dir))
	}
}

func PoolOrderPriceGapRatio(min, max, priceDiffRatio, maxPriceLimitRatio sdk.Dec) sdk.Dec {
	if priceDiffRatio.IsZero() {
		return min
	}
	a := max.Sub(min).Quo(maxPriceLimitRatio)
	return a.Mul(priceDiffRatio).Add(min)
}
