package amm

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	zeroInt = sdk.ZeroInt()
	zeroDec = sdk.ZeroDec()
	oneDec  = sdk.OneDec()
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

// MatchableAmount returns matchable amount of an order considering
// remaining offer coin and price.
func MatchableAmount(order Order, price sdk.Dec) (matchableAmt sdk.Int) {
	switch order.GetDirection() {
	case Buy:
		remainingOfferCoinAmt := order.GetOfferCoinAmount().Sub(order.GetPaidOfferCoinAmount())
		matchableAmt = sdk.MinInt(
			order.GetOpenAmount(),
			remainingOfferCoinAmt.ToDec().QuoTruncate(price).TruncateInt(),
		)
	case Sell:
		matchableAmt = order.GetOpenAmount()
	}
	return
}

// findFirstTrueCondition uses the binary search to find the first index
// where f(i) is true, while searching in range [start, end].
// It assumes that f(j) == false where j < i and f(j) == true where j >= i.
// start can be greater than end.
func findFirstTrueCondition(start, end int, f func(i int) bool) (i int, found bool) {
	if start < end {
		i = start + sort.Search(end-start+1, func(i int) bool {
			return f(start + i)
		})
		if i > end {
			return 0, false
		}
		return i, true
	}
	i = start - sort.Search(start-end+1, func(i int) bool {
		return f(start - i)
	})
	if i < end {
		return 0, false
	}
	return i, true
}

// inv returns the inverse of x.
func inv(x sdk.Dec) (r sdk.Dec) {
	r = oneDec.Quo(x)
	return
}
