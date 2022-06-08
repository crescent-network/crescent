package amm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Order = (*BaseOrder)(nil)
	_ Order = (*UserOrder)(nil)
	_ Order = (*PoolOrder)(nil)
)

// OrderDirection specifies an order direction, either buy or sell.
type OrderDirection int

// OrderDirection enumerations.
const (
	Buy OrderDirection = iota + 1
	Sell
)

func (dir OrderDirection) String() string {
	switch dir {
	case Buy:
		return "Buy"
	case Sell:
		return "Sell"
	default:
		return fmt.Sprintf("OrderDirection(%d)", dir)
	}
}

// Order is the universal interface of an order.
type Order interface {
	GetDirection() OrderDirection
	// GetBatchId returns the batch id where the order was created.
	// Batch id of 0 means the current batch.
	GetBatchId() uint64
	GetPrice() sdk.Dec
	GetAmount() sdk.Int // The original order amount
	// HasPriority returns true if the order has higher priority
	// than the other order.
	HasPriority(other Order) bool
}

// BaseOrder is the base struct for an Order.
type BaseOrder struct {
	Direction          OrderDirection
	Price              sdk.Dec
	Amount             sdk.Int
	OpenAmount         sdk.Int
	OfferCoin          sdk.Coin
	RemainingOfferCoin sdk.Coin
	ReceivedDemandCoin sdk.Coin
	Matched            bool
}

// NewBaseOrder returns a new BaseOrder.
func NewBaseOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int, offerCoin sdk.Coin, demandCoinDenom string) *BaseOrder {
	return &BaseOrder{
		Direction:          dir,
		Price:              price,
		Amount:             amt,
		OpenAmount:         amt,
		OfferCoin:          offerCoin,
		RemainingOfferCoin: offerCoin,
		ReceivedDemandCoin: sdk.NewCoin(demandCoinDenom, sdk.ZeroInt()),
	}
}

// GetDirection returns the order direction.
func (order *BaseOrder) GetDirection() OrderDirection {
	return order.Direction
}

func (order *BaseOrder) GetBatchId() uint64 {
	return 0
}

// GetPrice returns the order price.
func (order *BaseOrder) GetPrice() sdk.Dec {
	return order.Price
}

// GetAmount returns the order amount.
func (order *BaseOrder) GetAmount() sdk.Int {
	return order.Amount
}

// HasPriority returns whether the order has higher priority than
// the other order.
func (order *BaseOrder) HasPriority(other Order) bool {
	return order.Amount.GT(other.GetAmount())
}

// GetOpenAmount returns open(not matched) amount of the order.
func (order *BaseOrder) GetOpenAmount() sdk.Int {
	return order.OpenAmount
}

// SetOpenAmount sets open amount of the order.
func (order *BaseOrder) SetOpenAmount(amt sdk.Int) {
	order.OpenAmount = amt
}

func (order *BaseOrder) GetOfferCoin() sdk.Coin {
	return order.OfferCoin
}

// GetRemainingOfferCoin returns remaining offer coin of the order.
func (order *BaseOrder) GetRemainingOfferCoin() sdk.Coin {
	return order.RemainingOfferCoin
}

// DecrRemainingOfferCoin decrements remaining offer coin amount of the order.
func (order *BaseOrder) DecrRemainingOfferCoin(amt sdk.Int) {
	order.RemainingOfferCoin = order.RemainingOfferCoin.SubAmount(amt)
}

// GetReceivedDemandCoin returns received demand coin of the order.
func (order *BaseOrder) GetReceivedDemandCoin() sdk.Coin {
	return order.ReceivedDemandCoin
}

// IncrReceivedDemandCoin increments received demand coin amount of the order.
func (order *BaseOrder) IncrReceivedDemandCoin(amt sdk.Int) {
	order.ReceivedDemandCoin = order.ReceivedDemandCoin.AddAmount(amt)
}

// IsMatched returns whether the order is matched or not.
func (order *BaseOrder) IsMatched() bool {
	return order.Matched
}

// SetMatched sets whether the order is matched or not.
func (order *BaseOrder) SetMatched(matched bool) {
	order.Matched = matched
}

type UserOrder struct {
	BaseOrder
	OrderId uint64
	BatchId uint64
}

func (order *UserOrder) GetBatchId() uint64 {
	return order.BatchId
}

func (order *UserOrder) HasPriority(other Order) bool {
	if !order.Amount.Equal(other.GetAmount()) {
		return order.Amount.GT(other.GetAmount())
	}
	switch other := other.(type) {
	case *UserOrder:
		return order.OrderId < other.OrderId
	case *PoolOrder:
		return true
	default:
		panic(fmt.Errorf("invalid order type: %T", other))
	}
}

type PoolOrder struct {
	BaseOrder
	PoolId uint64
}

func (order *PoolOrder) HasPriority(other Order) bool {
	if !order.Amount.Equal(other.GetAmount()) {
		return order.Amount.GT(other.GetAmount())
	}
	switch other := other.(type) {
	case *UserOrder:
		return false
	case *PoolOrder:
		return order.PoolId < other.PoolId
	default:
		panic(fmt.Errorf("invalid order type: %T", other))
	}
}

func TotalAmount(orders []Order) sdk.Int {
	amt := sdk.ZeroInt()
	for _, order := range orders {
		amt = amt.Add(order.GetAmount())
	}
	return amt
}
