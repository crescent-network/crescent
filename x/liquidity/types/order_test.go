package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func newUserOrder(dir amm.OrderDirection, orderId uint64, price sdk.Dec, amt sdk.Int) *types.UserOrder {
	return &types.UserOrder{
		BaseOrder: amm.NewBaseOrder(dir, price, amt, sdk.Coin{}, "denom"),
		OrderId:   orderId,
	}
}

func newPoolOrder(dir amm.OrderDirection, poolId uint64, price sdk.Dec, amt sdk.Int) *types.PoolOrder {
	return &types.PoolOrder{
		BaseOrder: amm.NewBaseOrder(dir, price, amt, sdk.Coin{}, "denom"),
		PoolId:    poolId,
	}
}

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

		orders := make([]amm.Order, n)
		for i := 0; i < n; i++ {
			price := amm.TickFromIndex(r.Intn(100)+10000, 3)
			amt := newInt(r.Int63n(500) + 100)
			if r.Intn(2) == 0 {
				var reqId uint64
				reqId, reqIds = reqIds[0], reqIds[1:]
				orders[i] = newUserOrder(amm.Buy, reqId, price, amt)
			} else {
				var poolId uint64
				poolId, poolIds = poolIds[0], poolIds[1:]
				orders[i] = newPoolOrder(amm.Buy, poolId, price, amt)
			}
		}

		const ascendingPrice, descendingPrice = 1, 2
		for _, priceCmp := range []int{1, 2} {
			switch priceCmp {
			case ascendingPrice:
				types.SortOrders(orders, types.PriceAscending)
			case descendingPrice:
				types.SortOrders(orders, types.PriceDescending)
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
								require.Greater(t, orderA.OrderId, orderB.OrderId)
							case *types.PoolOrder:
								// ok
							}
						case *types.PoolOrder:
							switch orderB := orders[i-1].(type) {
							case *types.UserOrder:
								t.Error("not sorted")
								t.FailNow()
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
