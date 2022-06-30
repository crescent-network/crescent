package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
)

// OrderDirectionFromAMM converts amm.OrderDirection to liquidity module's
// OrderDirection.
func OrderDirectionFromAMM(dir amm.OrderDirection) OrderDirection {
	switch dir {
	case amm.Buy:
		return OrderDirectionBuy
	case amm.Sell:
		return OrderDirectionSell
	default:
		panic(fmt.Errorf("invalid order direction: %s", dir))
	}
}

type UserOrder struct {
	*amm.BaseOrder
	Orderer                         sdk.AccAddress
	OrderId                         uint64
	BatchId                         uint64
	OfferCoinDenom, DemandCoinDenom string
}

// NewUserOrder returns a new user order.
func NewUserOrder(order Order) *UserOrder {
	var dir amm.OrderDirection
	var amt sdk.Int
	switch order.Direction {
	case OrderDirectionBuy:
		dir = amm.Buy
		utils.SafeMath(func() {
			amt = sdk.MinInt(
				order.OpenAmount,
				order.RemainingOfferCoin.Amount.ToDec().QuoTruncate(order.Price).TruncateInt(),
			)
		}, func() {
			amt = order.OpenAmount
		})
	case OrderDirectionSell:
		dir = amm.Sell
		amt = order.OpenAmount
	}
	return &UserOrder{
		BaseOrder:       amm.NewBaseOrder(dir, order.Price, amt, order.RemainingOfferCoin.Amount),
		Orderer:         order.GetOrderer(),
		OrderId:         order.Id,
		BatchId:         order.BatchId,
		OfferCoinDenom:  order.OfferCoin.Denom,
		DemandCoinDenom: order.ReceivedCoin.Denom,
	}
}

func (order *UserOrder) GetBatchId() uint64 {
	return order.BatchId
}

func (order *UserOrder) HasPriority(other amm.Order) bool {
	if !order.Amount.Equal(other.GetAmount()) {
		return order.BaseOrder.HasPriority(other)
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
	*amm.BaseOrder
	PoolId                          uint64
	ReserveAddress                  sdk.AccAddress
	OfferCoinDenom, DemandCoinDenom string
}

func NewPoolOrder(
	poolId uint64, reserveAddr sdk.AccAddress, dir amm.OrderDirection, price sdk.Dec, amt sdk.Int,
	offerCoinDenom, demandCoinDenom string) *PoolOrder {
	return &PoolOrder{
		BaseOrder:       amm.NewBaseOrder(dir, price, amt, amm.OfferCoinAmount(dir, price, amt)),
		PoolId:          poolId,
		ReserveAddress:  reserveAddr,
		OfferCoinDenom:  offerCoinDenom,
		DemandCoinDenom: demandCoinDenom,
	}
}

func (order *PoolOrder) HasPriority(other amm.Order) bool {
	if !order.Amount.Equal(other.GetAmount()) {
		return order.BaseOrder.HasPriority(other)
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
