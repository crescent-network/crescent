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
	GetRemainingOfferCoin() sdk.Coin
	DecrRemainingOfferCoin(amt sdk.Int) Order // Decrement remaining offer coin amount
	GetReceivedDemandCoin() sdk.Coin
	IncrReceivedDemandCoin(amt sdk.Int) Order // Increment received demand coin amount
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
	Direction          OrderDirection
	Price              sdk.Dec
	Amount             sdk.Int
	OpenAmount         sdk.Int
	RemainingOfferCoin sdk.Coin
	ReceivedDemandCoin sdk.Coin
	Matched            bool
}

func NewBaseOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int, offerCoin sdk.Coin, demandCoinDenom string) *BaseOrder {
	return &BaseOrder{
		Direction:          dir,
		Price:              price,
		Amount:             amt,
		OpenAmount:         amt,
		RemainingOfferCoin: offerCoin,
		ReceivedDemandCoin: sdk.NewCoin(demandCoinDenom, sdk.ZeroInt()),
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

func (order *BaseOrder) GetRemainingOfferCoin() sdk.Coin {
	return order.RemainingOfferCoin
}

func (order *BaseOrder) DecrRemainingOfferCoin(amt sdk.Int) Order {
	order.RemainingOfferCoin = order.RemainingOfferCoin.SubAmount(amt)
	return order
}

func (order *BaseOrder) GetReceivedDemandCoin() sdk.Coin {
	return order.ReceivedDemandCoin
}

func (order *BaseOrder) IncrReceivedDemandCoin(amt sdk.Int) Order {
	order.ReceivedDemandCoin = order.ReceivedDemandCoin.AddAmount(amt)
	return order
}

func (order *BaseOrder) IsMatched() bool {
	return order.Matched
}

func (order *BaseOrder) SetMatched(matched bool) Order {
	order.Matched = matched
	return order
}
