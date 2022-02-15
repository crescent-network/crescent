package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

var (
	_ amm.Order = (*UserOrder)(nil)
	_ amm.Order = (*PoolOrder)(nil)

	DescendingPrice PriceComparator = func(a, b amm.Order) bool {
		return a.GetPrice().GT(b.GetPrice())
	}
	AscendingPrice PriceComparator = func(a, b amm.Order) bool {
		return a.GetPrice().LT(b.GetPrice())
	}
)

type PriceComparator func(a, b amm.Order) bool

func SortOrders(orders []amm.Order, cmp PriceComparator) {
	sort.SliceStable(orders, func(i, j int) bool {
		switch orderA := orders[i].(type) {
		case *UserOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return orderA.RequestId < orderB.RequestId
			case *PoolOrder:
				return true
			}
		case *PoolOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return false
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

type UserOrder struct {
	*amm.BaseOrder
	RequestId uint64
	Orderer   sdk.AccAddress
}

func NewUserOrder(order Order) *UserOrder {
	var dir amm.OrderDirection
	switch order.Direction {
	case OrderDirectionBuy:
		dir = amm.Buy
	case OrderDirectionSell:
		dir = amm.Sell
	}
	return &UserOrder{
		BaseOrder: amm.NewBaseOrder(dir, order.Price, order.OpenAmount, order.RemainingOfferCoin, order.ReceivedCoin.Denom),
		RequestId: order.Id,
		Orderer:   order.GetOrderer(),
	}
}

type PoolOrder struct {
	*amm.BaseOrder
	PoolId         uint64
	ReserveAddress sdk.AccAddress
	OfferCoin      sdk.Coin
}

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
