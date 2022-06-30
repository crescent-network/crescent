package amm

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
)

var (
	zeroInt = sdk.ZeroInt()
	oneDec  = sdk.OneDec()
	fourDec = sdk.NewDec(4)
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
	if price.MulInt(matchableAmt).TruncateInt().IsZero() {
		matchableAmt = zeroInt
	}
	return
}

// TotalAmount returns total amount of orders.
func TotalAmount(orders []Order) sdk.Int {
	amt := sdk.ZeroInt()
	for _, order := range orders {
		amt = amt.Add(order.GetAmount())
	}
	return amt
}

// TotalMatchableAmount returns total matchable amount of orders.
func TotalMatchableAmount(orders []Order, price sdk.Dec) (amt sdk.Int) {
	amt = sdk.ZeroInt()
	for _, order := range orders {
		amt = amt.Add(MatchableAmount(order, price))
	}
	return
}

// OrderGroup represents a group of orders with same batch id.
type OrderGroup struct {
	BatchId uint64
	Orders  []Order
}

// GroupOrdersByBatchId groups orders by their batch id and returns a
// slice of OrderGroup.
func GroupOrdersByBatchId(orders []Order) (groups []*OrderGroup) {
	groupByBatchId := map[uint64]*OrderGroup{}
	for _, order := range orders {
		group, ok := groupByBatchId[order.GetBatchId()]
		if !ok {
			i := sort.Search(len(groups), func(i int) bool {
				if order.GetBatchId() == 0 {
					return groups[i].BatchId == 0
				}
				if groups[i].BatchId == 0 {
					return true
				}
				return order.GetBatchId() <= groups[i].BatchId
			})
			group = &OrderGroup{BatchId: order.GetBatchId()}
			groupByBatchId[order.GetBatchId()] = group
			groups = append(groups[:i], append([]*OrderGroup{group}, groups[i:]...)...)
		}
		group.Orders = append(group.Orders, order)
	}
	return
}

// SortOrders sorts orders using its HasPriority condition.
func SortOrders(orders []Order) {
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].HasPriority(orders[j])
	})
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

var (
	// Pool price gap ratio function thresholds
	t1 = utils.ParseDec("0.01")
	t2 = utils.ParseDec("0.02")
	t3 = utils.ParseDec("0.1")

	// Pool price gap ratio function coefficients
	a1, b1 = utils.ParseDec("0.007"), utils.ParseDec("0.00003")
	a2, b2 = utils.ParseDec("0.09"), utils.ParseDec("-0.0008")
	a3     = utils.ParseDec("0.05")
	b4     = utils.ParseDec("0.005")
)

func poolOrderPriceGapRatio(poolPrice, currentPrice sdk.Dec) (r sdk.Dec) {
	if poolPrice.IsZero() {
		poolPrice = sdk.NewDecWithPrec(1, sdk.Precision) // lowest possible sdk.Dec
	}
	x := currentPrice.Sub(poolPrice).Abs().Quo(poolPrice)
	switch {
	case x.LTE(t1):
		return a1.Mul(x).Add(b1)
	case x.LTE(t2):
		return a2.Mul(x).Add(b2)
	case x.LTE(t3):
		return a3.Mul(x)
	default:
		return b4
	}
}
