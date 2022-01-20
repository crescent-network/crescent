package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Order = (*BaseOrder)(nil)
	_ Order = (*UserOrder)(nil)
	_ Order = (*PoolOrder)(nil)
)

type Order interface {
	GetDirection() SwapDirection
	GetPrice() sdk.Dec
	GetOpenBaseCoinAmount() sdk.Int
	SetOpenBaseCoinAmount(amount sdk.Int) Order
	GetRemainingOfferCoinAmount() sdk.Int
	SetRemainingOfferCoinAmount(amount sdk.Int) Order
	GetReceivedAmount() sdk.Int
	SetReceivedAmount(amount sdk.Int) Order
}

type Orders []Order

func (orders Orders) OpenBaseCoinAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.GetOpenBaseCoinAmount())
	}
	return amount
}

type BaseOrder struct {
	Direction          SwapDirection
	Price              sdk.Dec
	OpenBaseCoinAmount       sdk.Int
	RemainingOfferCoinAmount sdk.Int
	ReceivedAmount           sdk.Int
}

func NewBaseOrder(dir SwapDirection, price sdk.Dec, baseCoinAmt, offerCoinAmt sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction:                dir,
		Price:                    price,
		OpenBaseCoinAmount:       baseCoinAmt,
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

func (order *BaseOrder) GetOpenBaseCoinAmount() sdk.Int {
	return order.OpenBaseCoinAmount
}

func (order *BaseOrder) SetOpenBaseCoinAmount(amount sdk.Int) Order {
	order.OpenBaseCoinAmount = amount
	return order
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
	OfferCoinAmount sdk.Int
	ReserveAddress  sdk.AccAddress
}

func NewPoolOrder(reserveAddr sdk.AccAddress, dir SwapDirection, price sdk.Dec, amount sdk.Int) *PoolOrder {
	return &PoolOrder{
		BaseOrder: BaseOrder{
			Direction:                dir,
			Price:                    price,
			OpenBaseCoinAmount:       amount,
			RemainingOfferCoinAmount: amount,
			ReceivedAmount:           sdk.ZeroInt(),
		},
		OfferCoinAmount: amount,
		ReserveAddress:  reserveAddr,
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
