package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func TestOrders_Sort(t *testing.T) {
	for seed := int64(0); seed < 10; seed++ {
		r := rand.New(rand.NewSource(seed))

		const n = 1000

		reqIds := make([]uint64, n)
		for i := uint64(0); i < n; i++ {
			reqIds[i] = i + 1
		}
		rand.Shuffle(len(reqIds), func(i, j int) {
			reqIds[i], reqIds[j] = reqIds[j], reqIds[i]
		})
		poolIds := make([]uint64, n)
		for i := uint64(0); i < n; i++ {
			poolIds[i] = i + 1
		}
		rand.Shuffle(len(poolIds), func(i, j int) {
			poolIds[i], poolIds[j] = poolIds[j], poolIds[i]
		})

		orders := make(types.Orders, n)
		for i := 0; i < n; i++ {
			price := types.TickFromIndex(r.Intn(100)+10000, tickPrec)
			amt := newInt(r.Int63n(500) + 100)
			if r.Intn(2) == 0 {
				var reqId uint64
				reqId, reqIds = reqIds[0], reqIds[1:]
				orders[i] = newBuyUserOrder(reqId, price, amt)
			} else {
				var poolId uint64
				poolId, poolIds = poolIds[0], poolIds[1:]
				orders[i] = newBuyPoolOrder(poolId, price, amt)
			}
		}

		const ascendingPrice, descendingPrice = 1, 2
		for _, priceCmp := range []int{1, 2} {
			switch priceCmp {
			case ascendingPrice:
				orders.Sort(types.AscendingPrice)
			case descendingPrice:
				orders.Sort(types.DescendingPrice)
			}
			for i := 1; i < n; i++ {
				switch priceCmp {
				case ascendingPrice:
					require.True(t, orders[i].GetPrice().GTE(orders[i-1].GetPrice()))
				case descendingPrice:
					require.True(t, orders[i].GetPrice().LTE(orders[i-1].GetPrice()))
				}
				if orders[i].GetPrice().Equal(orders[i-1].GetPrice()) {
					require.True(t, orders[i].GetAmount().LTE(orders[i-1].GetAmount()))
					if orders[i].GetAmount().Equal(orders[i-1].GetAmount()) {
						switch orderA := orders[i].(type) {
						case *types.UserOrder:
							switch orderB := orders[i-1].(type) {
							case *types.UserOrder:
								require.Greater(t, orderA.RequestId, orderB.RequestId)
							case *types.PoolOrder:
								t.Error("not sorted")
								t.FailNow()
							}
						case *types.PoolOrder:
							switch orderB := orders[i-1].(type) {
							case *types.UserOrder:
								// ok
							case *types.PoolOrder:
								require.Greater(t, orderA.PoolId, orderB.PoolId)
							}
						}
					}
				}
			}
		}
	}
}
