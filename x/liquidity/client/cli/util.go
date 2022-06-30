package cli

import (
	"fmt"
	"strings"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

// excConditions returns true when exactly one condition is true.
func excConditions(conditions ...bool) bool {
	cnt := 0
	for _, condition := range conditions {
		if condition {
			cnt++
		}
	}
	return cnt == 1
}

// parseOrderDirection parses order direction string and returns
// types.OrderDirection.
func parseOrderDirection(s string) (types.OrderDirection, error) {
	switch strings.ToLower(s) {
	case "buy", "b":
		return types.OrderDirectionBuy, nil
	case "sell", "s":
		return types.OrderDirectionSell, nil
	}
	return 0, fmt.Errorf("invalid order direction: %s", s)
}
