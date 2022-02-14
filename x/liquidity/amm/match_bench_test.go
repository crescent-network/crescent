package amm_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

func BenchmarkFindMatchPrice(b *testing.B) {
	minPrice, maxPrice := squad.ParseDec("0.0000001"), squad.ParseDec("10000000")
	minAmt, maxAmt := sdk.NewInt(100), sdk.NewInt(10000000)
	minReserveAmt, maxReserveAmt := sdk.NewInt(500), sdk.NewInt(1000000000)

	for seed := int64(0); seed < 5; seed++ {
		b.Run(fmt.Sprintf("seed/%d", seed), func(b *testing.B) {
			r := rand.New(rand.NewSource(seed))
			ob := amm.NewOrderBook()
			for i := 0; i < 10000; i++ {
				ob.Add(newOrder(amm.Buy, squad.RandomDec(r, minPrice, maxPrice), squad.RandomInt(r, minAmt, maxAmt)))
				ob.Add(newOrder(amm.Sell, squad.RandomDec(r, minPrice, maxPrice), squad.RandomInt(r, minAmt, maxAmt)))
			}
			var poolOrderSources []amm.OrderSource
			for i := 0; i < 1000; i++ {
				rx, ry := squad.RandomInt(r, minReserveAmt, maxReserveAmt), squad.RandomInt(r, minReserveAmt, maxReserveAmt)
				pool := amm.NewBasicPool(rx, ry, sdk.ZeroInt())
				poolOrderSources = append(poolOrderSources, amm.NewMockPoolOrderSource(pool, "denom1", "denom2"))
			}
			os := amm.MergeOrderSources(append(poolOrderSources, ob)...)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				amm.FindMatchPrice(os, defTickPrec)
			}
		})
	}
}
