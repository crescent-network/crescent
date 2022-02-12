package amm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Order = (*BaseOrder)(nil)

type OrderDirection int

const (
	Buy OrderDirection = iota + 1
	Sell
)

func (dir OrderDirection) String() string {
	switch dir {
	case Buy:
		return "buy"
	case Sell:
		return "sell"
	default:
		return fmt.Sprintf("OrderDirection(%d)", dir)
	}
}

type Order interface {
	GetDirection() OrderDirection
	GetPrice() sdk.Dec
	GetAmount() sdk.Int
	GetOpenAmount() sdk.Int
	SetOpenAmount(amt sdk.Int) Order
	GetRemainingOfferCoinAmount() sdk.Int
	SetRemainingOfferCoinAmount(amt sdk.Int) Order
	GetReceivedDemandCoinAmount() sdk.Int
	SetReceivedDemandCoinAmount(amt sdk.Int) Order
	IsMatched() bool
	SetMatched(matched bool) Order
}

func TotalOpenAmount(orders []Order) sdk.Int {
	amt := sdk.ZeroInt()
	for _, order := range orders {
		amt = amt.Add(order.GetOpenAmount())
	}
	return amt
}

type BaseOrder struct {
	Direction                OrderDirection
	Price                    sdk.Dec
	Amount                   sdk.Int
	OpenAmount               sdk.Int
	RemainingOfferCoinAmount sdk.Int
	ReceivedDemandCoinAmount sdk.Int
	Matched                  bool
}

func NewBaseOrder(dir OrderDirection, price sdk.Dec, amt, offerCoinAmt sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction:                dir,
		Price:                    price,
		Amount:                   amt,
		OpenAmount:               amt,
		RemainingOfferCoinAmount: offerCoinAmt,
		ReceivedDemandCoinAmount: sdk.ZeroInt(),
	}
}

func (order *BaseOrder) GetDirection() OrderDirection {
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

func (order *BaseOrder) SetOpenAmount(amt sdk.Int) Order {
	order.OpenAmount = amt
	return order
}

func (order *BaseOrder) GetRemainingOfferCoinAmount() sdk.Int {
	return order.RemainingOfferCoinAmount
}

func (order *BaseOrder) SetRemainingOfferCoinAmount(amount sdk.Int) Order {
	order.RemainingOfferCoinAmount = amount
	return order
}

func (order *BaseOrder) GetReceivedDemandCoinAmount() sdk.Int {
	return order.ReceivedDemandCoinAmount
}

func (order *BaseOrder) SetReceivedDemandCoinAmount(amt sdk.Int) Order {
	order.ReceivedDemandCoinAmount = amt
	return order
}

func (order *BaseOrder) IsMatched() bool {
	return order.Matched
}

func (order *BaseOrder) SetMatched(matched bool) Order {
	order.Matched = matched
	return order
}
