package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Order = (*BaseOrder)(nil)
	_ Order = (*UserOrder)(nil)
	_ Order = (*PoolOrder)(nil)

	DescendingPrice PriceComparator = func(a, b Order) bool {
		return a.GetPrice().GT(b.GetPrice())
	}
	AscendingPrice PriceComparator = func(a, b Order) bool {
		return a.GetPrice().LT(b.GetPrice())
	}
)

type Order interface {
	GetDirection() SwapDirection
	GetPrice() sdk.Dec
	GetBaseCoinAmount() sdk.Int
	GetOpenBaseCoinAmount() sdk.Int
	SetOpenBaseCoinAmount(amount sdk.Int) Order
	GetOfferCoinAmount() sdk.Int
	GetRemainingOfferCoinAmount() sdk.Int
	SetRemainingOfferCoinAmount(amount sdk.Int) Order
	GetReceivedAmount() sdk.Int
	SetReceivedAmount(amount sdk.Int) Order
}

type PriceComparator func(a, b Order) bool

type Orders []Order

func (orders Orders) OpenBaseCoinAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.GetOpenBaseCoinAmount())
	}
	return amount
}

func (orders Orders) Sort(cmp PriceComparator) {
	sort.SliceStable(orders, func(i, j int) bool {
		switch orderA := orders[i].(type) {
		case *UserOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return orderA.RequestId > orderB.RequestId
			case *PoolOrder:
				return true
			}
		case *PoolOrder:
			switch orderB := orders[j].(type) {
			case *UserOrder:
				return false
			case *PoolOrder:
				return orderA.PoolId > orderB.PoolId
			}
		}
		return false // not reachable
	})
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].GetBaseCoinAmount().GT(orders[j].GetBaseCoinAmount())
	})
	sort.SliceStable(orders, func(i, j int) bool {
		return cmp(orders[i], orders[j])
	})
}

type BaseOrder struct {
	Direction                SwapDirection
	Price                    sdk.Dec
	BaseCoinAmount           sdk.Int
	OpenBaseCoinAmount       sdk.Int
	OfferCoinAmount sdk.Int
	RemainingOfferCoinAmount sdk.Int
	ReceivedAmount           sdk.Int
}

func NewBaseOrder(dir SwapDirection, price sdk.Dec, baseCoinAmt, offerCoinAmt sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction:                dir,
		Price:                    price,
		BaseCoinAmount:           baseCoinAmt,
		OpenBaseCoinAmount:       baseCoinAmt,
		OfferCoinAmount: offerCoinAmt,
		RemainingOfferCoinAmount: offerCoinAmt,
		ReceivedAmount:           sdk.ZeroInt(),
	}
}

func (order *BaseOrder) GetDirection() SwapDirection {
	return order.Direction
}

func (order *BaseOrder) GetPrice() sdk.Dec {
	return order.Price
}

func (order *BaseOrder) GetBaseCoinAmount() sdk.Int {
	return order.BaseCoinAmount
}

func (order *BaseOrder) GetOpenBaseCoinAmount() sdk.Int {
	return order.OpenBaseCoinAmount
}

func (order *BaseOrder) SetOpenBaseCoinAmount(amount sdk.Int) Order {
	order.OpenBaseCoinAmount = amount
	return order
}

func (order *BaseOrder) GetOfferCoinAmount() sdk.Int {
	return order.OfferCoinAmount
}

func (order *BaseOrder) GetRemainingOfferCoinAmount() sdk.Int {
	return order.RemainingOfferCoinAmount
}

func (order *BaseOrder) SetRemainingOfferCoinAmount(amount sdk.Int) Order {
	order.RemainingOfferCoinAmount = amount
	return order
}

func (order *BaseOrder) GetReceivedAmount() sdk.Int {
	return order.ReceivedAmount
}

func (order *BaseOrder) SetReceivedAmount(amount sdk.Int) Order {
	order.ReceivedAmount = amount
	return order
}

type UserOrder struct {
	BaseOrder
	RequestId uint64
	Orderer   sdk.AccAddress
}

func NewUserOrder(req SwapRequest) *UserOrder {
	return &UserOrder{
		BaseOrder: BaseOrder{
			Direction:                req.Direction,
			Price:                    req.Price,
			OpenBaseCoinAmount:       req.OpenBaseCoinAmount,
			RemainingOfferCoinAmount: req.RemainingOfferCoin.Amount,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		RequestId: req.Id,
		Orderer:   req.GetOrderer(),
	}
}

func (order *UserOrder) SetOpenBaseCoinAmount(amount sdk.Int) Order {
	order.BaseOrder.SetOpenBaseCoinAmount(amount)
	return order
}

func (order *UserOrder) SetRemainingOfferCoinAmount(amount sdk.Int) Order {
	order.BaseOrder.SetRemainingOfferCoinAmount(amount)
	return order
}

func (order *UserOrder) SetReceivedAmount(amount sdk.Int) Order {
	order.BaseOrder.SetReceivedAmount(amount)
	return order
}

type PoolOrder struct {
	BaseOrder
	PoolId          uint64
	ReserveAddress  sdk.AccAddress
	OfferCoinAmount sdk.Int
}

func NewPoolOrder(poolId uint64, reserveAddr sdk.AccAddress, dir SwapDirection, price sdk.Dec, amount sdk.Int) *PoolOrder {
	return &PoolOrder{
		BaseOrder: BaseOrder{
			Direction:                dir,
			Price:                    price,
			OpenBaseCoinAmount:       amount,
			RemainingOfferCoinAmount: amount,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		PoolId:          poolId,
		ReserveAddress:  reserveAddr,
		OfferCoinAmount: amount,
	}
}

func (order *PoolOrder) SetOpenBaseCoinAmount(amount sdk.Int) Order {
	order.BaseOrder.SetOpenBaseCoinAmount(amount)
	return order
}

func (order *PoolOrder) SetRemainingOfferCoinAmount(amount sdk.Int) Order {
	order.BaseOrder.SetRemainingOfferCoinAmount(amount)
	return order
}

func (order *PoolOrder) SetReceivedAmount(amount sdk.Int) Order {
	order.BaseOrder.SetReceivedAmount(amount)
	return order
}
