package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/cremath"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestPoolOrders(t *testing.T) {
	types.PoolOrders(
		true, cremath.MustNewBigDecFromStr("1"), sdk.NewInt(10000000), sdk.NewDec(1000000000),
		-1000, 10, sdk.NewDec(10000), sdk.NewDec(10000), func(price, qty, openQty sdk.Dec) (stop bool) {
			fmt.Println(price, qty, openQty)
			return false
		})
}
