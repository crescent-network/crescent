package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
)

var (
	_ amm.Order = (*UserOrder)(nil)
	_ amm.Order = (*PoolOrder)(nil)

	// PriceDescending defines a price comparator which is used to sort orders
	// by price in descending order.
	PriceDescending PriceComparator = func(a, b amm.Order) bool {
		return a.GetPrice().GT(b.GetPrice())
	}
	// PriceAscending defines a price comparator which is used to sort orders
	// by price in ascending order.
	PriceAscending PriceComparator = func(a, b amm.Order) bool {
		return a.GetPrice().LT(b.GetPrice())
	}
)

// PriceComparator is used to sort orders by price.
type PriceComparator func(a, b amm.Order) bool

// SortOrders sorts orders by these four criteria:
// 1. Price - descending/ascending based on PriceComparator
// 2. Amount - Larger amount takes higher priority than smaller amount
// 3. Order type - pool orders take higher priority than user orders
// 4. Time - early orders take higher priority. For pools, the pool with
//    lower pool id takes higher priority
func SortOrders(orders []amm.Order, cmp PriceComparator) {
	sort.SliceStable(orders, func(i, j int) bool {
		switch orderA := orders[i].(type) {
		case *UserOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return orderA.OrderId < orderB.OrderId
			case *PoolOrder:
				return false
			}
		case *PoolOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return true
			case *PoolOrder:
				return orderA.PoolId < orderB.PoolId
			}
		}
		panic(fmt.Sprintf("unknown order types: (%T, %T)", orders[i], orders[j]))
	})
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].GetAmount().GT(orders[j].GetAmount())
	})
	sort.SliceStable(orders, func(i, j int) bool {
		return cmp(orders[i], orders[j])
	})
}

// UserOrder is the user order type.
type UserOrder struct {
	*amm.BaseOrder
	OrderId uint64
	Orderer sdk.AccAddress
}

// NewUserOrder returns a new user order.
func NewUserOrder(order Order) *UserOrder {
	var dir amm.OrderDirection
	var amt sdk.Int
	switch order.Direction {
	case OrderDirectionBuy:
		dir = amm.Buy
		utils.SafeMath(func() {
			amt = sdk.MinInt(
				order.OpenAmount,
				order.RemainingOfferCoin.Amount.ToDec().QuoTruncate(order.Price).TruncateInt(),
			)
		}, func() {
			amt = order.OpenAmount
		})
	case OrderDirectionSell:
		dir = amm.Sell
		amt = order.OpenAmount
	}
	return &UserOrder{
		BaseOrder: amm.NewBaseOrder(dir, order.Price, amt, order.RemainingOfferCoin, order.ReceivedCoin.Denom),
		OrderId:   order.Id,
		Orderer:   order.GetOrderer(),
	}
}

// PoolOrder is the pool order type.
type PoolOrder struct {
	*amm.BaseOrder
	PoolId         uint64
	ReserveAddress sdk.AccAddress
	OfferCoin      sdk.Coin
}

// NewPoolOrder returns a new pool order.
func NewPoolOrder(
	poolId uint64, reserveAddr sdk.AccAddress, dir amm.OrderDirection, price sdk.Dec, amt sdk.Int,
	offerCoin sdk.Coin, demandCoinDenom string) *PoolOrder {
	return &PoolOrder{
		BaseOrder:      amm.NewBaseOrder(dir, price, amt, offerCoin, demandCoinDenom),
		PoolId:         poolId,
		ReserveAddress: reserveAddr,
		OfferCoin:      offerCoin,
	}
}
