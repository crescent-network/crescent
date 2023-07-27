package types

import (
	"fmt"

	utils "github.com/crescent-network/crescent/v5/types"
)

// ValidateOrderResult validates the result of order matching.
// TODO: remove after enough testing
func ValidateOrderResult(memOrder *MemOrder) {
	if memOrder.isBuy {
		price := memOrder.PaidWithoutFee().Sub(utils.SmallestDec).Quo(memOrder.ReceivedWithoutFee())
		if price.GT(memOrder.price) {
			panic(fmt.Errorf("match price %s > order price %s", price, memOrder.price))
		}
	} else {
		price := memOrder.ReceivedWithoutFee().Add(utils.SmallestDec).Quo(memOrder.PaidWithoutFee())
		if price.LT(memOrder.price) {
			panic(fmt.Errorf("match price %s < order price %s", price, memOrder.price))
		}
	}
}
