package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Order       = (*BaseOrder)(nil)
	_ Order       = (*UserOrder)(nil)
	_ Order       = (*PoolOrder)(nil)
)

type Order interface {
	GetDirection() SwapDirection
	GetPrice() sdk.Dec
	GetAmount() sdk.Int
	GetRemainingAmount() sdk.Int
	SetRemainingAmount(amount sdk.Int) Order
	GetReceivedAmount() sdk.Int
	SetReceivedAmount(amount sdk.Int) Order
	IsMatched() bool
	SetMatched(matched bool) Order
}

type Orders []Order

func (orders Orders) RemainingAmount() sdk.Int {
	amount := sdk.ZeroInt()
	for _, order := range orders {
		amount = amount.Add(order.GetRemainingAmount())
	}
	return amount
}

type BaseOrder struct {
	Direction       SwapDirection
	Price           sdk.Dec
	Amount          sdk.Int
	RemainingAmount sdk.Int
	ReceivedAmount  sdk.Int
	Matched         bool
}

func NewBaseOrder(dir SwapDirection, price sdk.Dec, amount sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction:       dir,
		Price:           price,
		Amount:          amount,
		RemainingAmount: amount,
		ReceivedAmount:  sdk.ZeroInt(),
		Matched:         false,
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

func (order *BaseOrder) GetRemainingAmount() sdk.Int {
	return order.RemainingAmount
}

func (order *BaseOrder) SetRemainingAmount(amount sdk.Int) Order {
	order.RemainingAmount = amount
	return order
}

func (order *BaseOrder) GetReceivedAmount() sdk.Int {
	return order.ReceivedAmount
}

func (order *BaseOrder) SetReceivedAmount(amount sdk.Int) Order {
	order.ReceivedAmount = amount
	return order
}

func (order *BaseOrder) IsMatched() bool {
	return order.Matched
}

func (order *BaseOrder) SetMatched(matched bool) Order {
	order.Matched = matched
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
			Direction:       req.Direction,
			Price:           req.Price,
			Amount:          req.RemainingCoin.Amount,
			RemainingAmount: req.RemainingCoin.Amount,
			ReceivedAmount:  sdk.ZeroInt(),
			Matched:         false,
		},
		RequestId: req.Id,
		Orderer:   req.GetOrderer(),
	}
}

type PoolOrder struct {
	BaseOrder
	ReserveAddress sdk.AccAddress
}

func NewPoolOrder(reserveAddr sdk.AccAddress, dir SwapDirection, price sdk.Dec, amount sdk.Int) *PoolOrder {
	return &PoolOrder{
		BaseOrder: BaseOrder{
			Direction:       dir,
			Price:           price,
			Amount:          amount,
			RemainingAmount: amount,
			ReceivedAmount:  sdk.ZeroInt(),
			Matched:         false,
		},
		ReserveAddress: reserveAddr,
	}
}
