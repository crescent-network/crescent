package amm_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

func randInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.Sub(min).BigInt())))
}

func randDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(sdk.NewDecFromBigIntWithPrec(new(big.Int).Rand(r, max.Sub(min).BigInt()), sdk.Precision))
}

func BenchmarkFindMatchPrice(b *testing.B) {
	minPrice, maxPrice := squad.ParseDec("0.0000001"), squad.ParseDec("10000000")
	minAmt, maxAmt := sdk.NewInt(100), sdk.NewInt(10000000)
	minReserveAmt, maxReserveAmt := sdk.NewInt(500), sdk.NewInt(1000000000)

	for seed := int64(0); seed < 5; seed++ {
		b.Run(fmt.Sprintf("seed/%d", seed), func(b *testing.B) {
			r := rand.New(rand.NewSource(seed))
			ob := amm.NewOrderBook()
			for i := 0; i < 10000; i++ {
				ob.Add(newOrder(amm.Buy, randDec(r, minPrice, maxPrice), randInt(r, minAmt, maxAmt)))
				ob.Add(newOrder(amm.Sell, randDec(r, minPrice, maxPrice), randInt(r, minAmt, maxAmt)))
			}
			var poolOrderSources []amm.OrderSource
			for i := 0; i < 1000; i++ {
				rx, ry := randInt(r, minReserveAmt, maxReserveAmt), randInt(r, minReserveAmt, maxReserveAmt)
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
