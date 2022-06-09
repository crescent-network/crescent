package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity/amm"
)

// NewUserOrder returns a new user order.
func NewUserOrder(order Order) *amm.UserOrder {
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
	return amm.NewUserOrder(order.Id, order.BatchId, dir, order.Price, amt)
}
