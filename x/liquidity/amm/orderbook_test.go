package amm_test

//func TestOrderBook(t *testing.T) {
//	ob := amm.NewOrderBook(
//		newOrder(amm.Buy, utils.ParseDec("10.01"), sdk.NewInt(10000)),
//		newOrder(amm.Buy, utils.ParseDec("10.00"), sdk.NewInt(10000)),
//		newOrder(amm.Buy, utils.ParseDec("9.999"), sdk.NewInt(10000)),
//		newOrder(amm.Sell, utils.ParseDec("9.999"), sdk.NewInt(10000)),
//		newOrder(amm.Buy, utils.ParseDec("9.998"), sdk.NewInt(10000)),
//		newOrder(amm.Sell, utils.ParseDec("9.998"), sdk.NewInt(10000)),
//		newOrder(amm.Sell, utils.ParseDec("9.997"), sdk.NewInt(10000)),
//		newOrder(amm.Sell, utils.ParseDec("9.996"), sdk.NewInt(10000)),
//	)
//
//	highest, found := ob.HighestBuyPrice()
//	require.True(t, found)
//	require.True(sdk.DecEq(t, utils.ParseDec("10.01"), highest))
//	lowest, found := ob.LowestSellPrice()
//	require.True(t, found)
//	require.True(sdk.DecEq(t, utils.ParseDec("9.996"), lowest))
//}
