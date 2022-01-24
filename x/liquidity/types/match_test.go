package types_test

import (
	"fmt"
	"testing"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestMatchOrders(t *testing.T) {
	ob := types.NewOrderBook(tickPrec)

	ob.AddOrders(
		newBuyOrder(parseDec("0.9"), newInt(7500)),
		newBuyOrder(parseDec("0.8"), newInt(5000)),
		newSellOrder(parseDec("0.7"), newInt(10000)),
	)

	types.MatchOrders(ob.BuyTicks.AllOrders(), ob.SellTicks.AllOrders(), parseDec("0.7137"))

	for _, order := range ob.AllOrders() {
		fmt.Printf("(%s, %s(%s), paid %s, received %s)\n",
			order.GetDirection(), order.GetAmount(), order.GetOpenAmount(), order.GetOfferCoinAmount().Sub(order.GetRemainingOfferCoinAmount()), order.GetReceivedAmount())
	}
}
