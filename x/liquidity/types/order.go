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
	GetAmount() sdk.Int
	GetOpenAmount() sdk.Int
	SetOpenAmount(amount sdk.Int) Order
	GetOfferCoinAmount() sdk.Int
	GetRemainingOfferCoinAmount() sdk.Int
	SetRemainingOfferCoinAmount(amount sdk.Int) Order
	GetReceivedAmount() sdk.Int
	SetReceivedAmount(amount sdk.Int) Order
}

type PriceComparator func(a, b Order) bool

type Orders []Order

func (orders Orders) OpenAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.GetOpenAmount())
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
		return orders[i].GetAmount().GT(orders[j].GetAmount())
	})
	sort.SliceStable(orders, func(i, j int) bool {
		return cmp(orders[i], orders[j])
	})
}

type BaseOrder struct {
	Direction                SwapDirection
	Price                    sdk.Dec
	Amount                   sdk.Int
	OpenAmount               sdk.Int
	OfferCoinAmount          sdk.Int
	RemainingOfferCoinAmount sdk.Int
	ReceivedAmount           sdk.Int
}

func NewBaseOrder(dir SwapDirection, price sdk.Dec, amt, offerCoinAmt sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction:                dir,
		Price:                    price,
		Amount:                   amt,
		OpenAmount:               amt,
		OfferCoinAmount:          offerCoinAmt,
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

func (order *BaseOrder) GetAmount() sdk.Int {
	return order.Amount
}

func (order *BaseOrder) GetOpenAmount() sdk.Int {
	return order.OpenAmount
}

func (order *BaseOrder) SetOpenAmount(amount sdk.Int) Order {
	order.OpenAmount = amount
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
			Amount:                   req.OpenAmount,
			OpenAmount:               req.OpenAmount,
			OfferCoinAmount:          req.RemainingOfferCoin.Amount,
			RemainingOfferCoinAmount: req.RemainingOfferCoin.Amount,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		RequestId: req.Id,
		Orderer:   req.GetOrderer(),
	}
}

func (order *UserOrder) SetOpenAmount(amount sdk.Int) Order {
	order.BaseOrder.SetOpenAmount(amount)
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

func NewPoolOrder(poolId uint64, reserveAddr sdk.AccAddress, dir SwapDirection, price sdk.Dec, amt sdk.Int) *PoolOrder {
	var offerCoinAmt sdk.Int
	switch dir {
	case SwapDirectionBuy:
		offerCoinAmt = price.MulInt(amt).Ceil().TruncateInt()
	case SwapDirectionSell:
		offerCoinAmt = amt
	}
	return &PoolOrder{
		BaseOrder: BaseOrder{
			Direction:                dir,
			Price:                    price,
			Amount:                   amt,
			OpenAmount:               amt,
			OfferCoinAmount:          offerCoinAmt,
			RemainingOfferCoinAmount: offerCoinAmt,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		PoolId:          poolId,
		ReserveAddress:  reserveAddr,
		OfferCoinAmount: amt,
	}
}

func (order *PoolOrder) SetOpenAmount(amount sdk.Int) Order {
	order.BaseOrder.SetOpenAmount(amount)
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
