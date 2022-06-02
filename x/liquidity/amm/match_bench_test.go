package amm_test

//func BenchmarkFindMatchPrice(b *testing.B) {
//	minPrice, maxPrice := utils.ParseDec("0.0000001"), utils.ParseDec("10000000")
//	minAmt, maxAmt := sdk.NewInt(100), sdk.NewInt(10000000)
//	minReserveAmt, maxReserveAmt := sdk.NewInt(500), sdk.NewInt(1000000000)
//
//	for seed := int64(0); seed < 5; seed++ {
//		b.Run(fmt.Sprintf("seed/%d", seed), func(b *testing.B) {
//			r := rand.New(rand.NewSource(seed))
//			ob := amm.NewOrderBook()
//			for i := 0; i < 10000; i++ {
//				ob.AddOrder(newOrder(amm.Buy, utils.RandomDec(r, minPrice, maxPrice), utils.RandomInt(r, minAmt, maxAmt)))
//				ob.AddOrder(newOrder(amm.Sell, utils.RandomDec(r, minPrice, maxPrice), utils.RandomInt(r, minAmt, maxAmt)))
//			}
//			var poolOrderSources []amm.OrderSource
//			for i := 0; i < 1000; i++ {
//				rx, ry := utils.RandomInt(r, minReserveAmt, maxReserveAmt), utils.RandomInt(r, minReserveAmt, maxReserveAmt)
//				pool := amm.NewBasicPool(rx, ry, sdk.Int{})
//				poolOrderSources = append(poolOrderSources, amm.NewMockPoolOrderSource(pool, "denom1", "denom2"))
//			}
//			os := amm.MergeOrderSources(append(poolOrderSources, ob)...)
//			b.ResetTimer()
//			for i := 0; i < b.N; i++ {
//				amm.FindMatchPrice(os, int(defTickPrec))
//			}
//		})
//	}
//}
