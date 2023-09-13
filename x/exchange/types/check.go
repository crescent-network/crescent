package types

import (
	"fmt"

	utils "github.com/crescent-network/crescent/v5/types"
)

// ValidateOrderResult validates the result of order matching.
// TODO: remove after enough testing
func ValidateOrderResult(memOrder *MemOrder) {
	if memOrder.isBuy {
		if memOrder.ReceivedWithoutFee().IsPositive() {
			price := memOrder.PaidWithoutFee().Sub(utils.SmallestDec).QuoTruncate(memOrder.ReceivedWithoutFee()).
				Sub(utils.SmallestDec)
			if price.GT(memOrder.price) {
				panic(fmt.Errorf("match price %s > order price %s", price, memOrder.price))
			}
		}
	} else {
		if memOrder.PaidWithoutFee().IsPositive() {
			price := memOrder.ReceivedWithoutFee().Add(utils.SmallestDec).QuoRoundUp(memOrder.PaidWithoutFee()).
				Add(utils.SmallestDec)
			if price.LT(memOrder.price) {
				panic(fmt.Errorf("match price %s < order price %s", price, memOrder.price))
			}
		}
	}
}
