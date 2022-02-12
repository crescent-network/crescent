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
	amm.BaseOrder
	RequestId uint64
	Orderer   sdk.AccAddress
}

func NewUserOrder(req SwapRequest) *UserOrder {
	var dir amm.OrderDirection
	switch req.Direction {
	case SwapDirectionBuy:
		dir = amm.Buy
	case SwapDirectionSell:
		dir = amm.Sell
	}
	return &UserOrder{
		BaseOrder: amm.NewBaseOrder(dir, req.Price, req.OpenAmount, req.RemainingOfferCoin.Amount),
		RequestId: req.Id,
		Orderer:   req.GetOrderer(),
	}
}

func (order *UserOrder) SetOpenAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetOpenAmount(amt)
	return order
}

func (order *UserOrder) SetRemainingOfferCoinAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetRemainingOfferCoinAmount(amt)
	return order
}

func (order *UserOrder) SetReceivedDemandCoinAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetReceivedDemandCoinAmount(amt)
	return order
}

type PoolOrder struct {
	amm.BaseOrder
	PoolId          uint64
	ReserveAddress  sdk.AccAddress
	OfferCoinAmount sdk.Int
}

func NewPoolOrder(pool Pool, dir amm.OrderDirection, price sdk.Dec, amt, offerCoinAmt sdk.Int) *PoolOrder {
	return &PoolOrder{
		BaseOrder:      amm.NewBaseOrder(dir, price, amt, offerCoinAmt),
		PoolId:         pool.Id,
		ReserveAddress: pool.GetReserveAddress(),
	}
}

func (order *PoolOrder) SetOpenAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetOpenAmount(amt)
	return order
}

func (order *PoolOrder) SetRemainingOfferCoinAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetRemainingOfferCoinAmount(amt)
	return order
}

func (order *PoolOrder) SetReceivedDemandCoinAmount(amt sdk.Int) amm.Order {
	order.BaseOrder.SetReceivedDemandCoinAmount(amt)
	return order
}
