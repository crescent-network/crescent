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
	String() string
}

// BaseOrder is the base struct for an Order.
type BaseOrder struct {
	Direction OrderDirection
	Price     sdk.Dec
	Amount    sdk.Int
}

// NewBaseOrder returns a new BaseOrder.
func NewBaseOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) *BaseOrder {
	return &BaseOrder{
		Direction: dir,
		Price:     price,
		Amount:    amt,
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

func (order *BaseOrder) String() string {
	return fmt.Sprintf("BaseOrder(%s,%s,%s)", order.Direction, order.Price, order.Amount)
}

type UserOrder struct {
	BaseOrder
	OrderId uint64
	BatchId uint64
}

func NewUserOrder(orderId, batchId uint64, dir OrderDirection, price sdk.Dec, amt sdk.Int) *UserOrder {
	return &UserOrder{
		BaseOrder: BaseOrder{
			Direction: dir,
			Price:     price,
			Amount:    amt,
		},
		OrderId: orderId,
		BatchId: batchId,
	}
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

func (order *UserOrder) String() string {
	return fmt.Sprintf("UserOrder(%d,%d,%s,%s,%s)",
		order.OrderId, order.BatchId, order.Direction, order.Price, order.Amount)
}

type PoolOrder struct {
	BaseOrder
	PoolId uint64
}

func NewPoolOrder(poolId uint64, dir OrderDirection, price sdk.Dec, amt sdk.Int) *PoolOrder {
	return &PoolOrder{
		BaseOrder: BaseOrder{
			Direction: dir,
			Price:     price,
			Amount:    amt,
		},
		PoolId: poolId,
	}
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

func (order *PoolOrder) String() string {
	return fmt.Sprintf("PoolOrder(%d,%s,%s,%s)",
		order.PoolId, order.Direction, order.Price, order.Amount)
}

func TotalAmount(orders []Order) sdk.Int {
	amt := sdk.ZeroInt()
	for _, order := range orders {
		amt = amt.Add(order.GetAmount())
	}
	return amt
}
