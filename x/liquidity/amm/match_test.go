package amm_test

//func TestFindMatchPrice(t *testing.T) {
//	for _, tc := range []struct {
//		name       string
//		os         amm.OrderSource
//		found      bool
//		matchPrice sdk.Dec
//	}{
//		{
//			"happy case",
//			amm.NewOrderBook(
//				newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
//				newOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
//			),
//			true,
//			utils.ParseDec("1.0"),
//		},
//		{
//			"buy order only",
//			amm.NewOrderBook(newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000))),
//			false,
//			sdk.Dec{},
//		},
//		{
//			"sell order only",
//			amm.NewOrderBook(newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000))),
//			false,
//			sdk.Dec{},
//		},
//		{
//			"highest buy price is lower than lowest sell price",
//			amm.NewOrderBook(
//				newOrder(amm.Buy, utils.ParseDec("0.9"), sdk.NewInt(10000)),
//				newOrder(amm.Sell, utils.ParseDec("1.1"), sdk.NewInt(10000)),
//			),
//			false,
//			sdk.Dec{},
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//			matchPrice, found := amm.FindMatchPrice(tc.os, int(defTickPrec))
//			require.Equal(t, tc.found, found)
//			if found {
//				require.Equal(t, tc.matchPrice, matchPrice)
//			}
//		})
//	}
//}

//func TestFindMatchPrice_Rounding(t *testing.T) {
//	basePrice := utils.ParseDec("0.9990")
//
//	for i := 0; i < 50; i++ {
//		ob := amm.NewOrderBook(
//			newOrder(amm.Buy, defTickPrec.UpTick(defTickPrec.UpTick(basePrice)), sdk.NewInt(80)),
//			newOrder(amm.Sell, defTickPrec.UpTick(basePrice), sdk.NewInt(20)),
//			newOrder(amm.Buy, basePrice, sdk.NewInt(10)), newOrder(amm.Sell, basePrice, sdk.NewInt(10)),
//			newOrder(amm.Sell, defTickPrec.DownTick(basePrice), sdk.NewInt(70)),
//		)
//		matchPrice, found := amm.FindMatchPrice(ob, int(defTickPrec))
//		require.True(t, found)
//		require.True(sdk.DecEq(t,
//			defTickPrec.RoundPrice(basePrice.Add(defTickPrec.UpTick(basePrice)).QuoInt64(2)),
//			matchPrice))
//
//		basePrice = defTickPrec.UpTick(basePrice)
//	}
//}
//
//func TestFindLastMatchableOrders(t *testing.T) {
//	_, _, _, _, found := amm.FindLastMatchableOrders(nil, nil, utils.ParseDec("1.0"))
//	require.False(t, found)
//
//	for seed := int64(0); seed < 100; seed++ {
//		r := rand.New(rand.NewSource(seed))
//
//		minPrice, maxPrice := utils.ParseDec("0.01"), utils.ParseDec("1.0")
//		minAmt, maxAmt := sdk.NewInt(30), sdk.NewInt(300)
//
//		for i := 0; i < 100; i++ {
//			var buyOrders, sellOrders []amm.Order
//			numBuyOrders := 1 + r.Intn(5)
//			numSellOrders := 1 + r.Intn(5)
//			for j := 0; j < numBuyOrders; j++ {
//				price := utils.ParseDec("1.0") // Price is not important.
//				amt := utils.RandomInt(r, minAmt, maxAmt)
//				buyOrders = append(buyOrders, newOrder(amm.Buy, price, amt))
//			}
//			for j := 0; j < numSellOrders; j++ {
//				price := utils.ParseDec("1.0") // Price is not important.
//				amt := utils.RandomInt(r, minAmt, maxAmt)
//				sellOrders = append(sellOrders, newOrder(amm.Sell, price, amt))
//			}
//			matchPrice := defTickPrec.PriceToDownTick(utils.RandomDec(r, minPrice, maxPrice))
//			// We don't sort orders like in real situations, and it doesn't
//			// actually matter.
//			bi, si, pmb, pms, found := amm.FindLastMatchableOrders(buyOrders, sellOrders, matchPrice)
//			if found {
//				buyAmt := amm.TotalAmount(buyOrders[:bi]).Add(pmb)
//				sellAmt := amm.TotalAmount(sellOrders[:si]).Add(pms)
//				require.True(sdk.IntEq(t, buyAmt, sellAmt))
//				require.False(t, matchPrice.MulInt(pms).TruncateInt().IsZero())
//			}
//		}
//	}
//}
//
//func TestMatchOrders(t *testing.T) {
//	_, matched := amm.MatchOrders(nil, nil, utils.ParseDec("1.0"))
//	require.False(t, matched)
//
//	for _, tc := range []struct {
//		name          string
//		os            amm.OrderSource
//		matchPrice    sdk.Dec
//		matched       bool
//		quoteCoinDust sdk.Int
//	}{
//		{
//			"happy case",
//			amm.NewOrderBook(
//				newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000)),
//				newOrder(amm.Sell, utils.ParseDec("1.0"), sdk.NewInt(10000)),
//			),
//			utils.ParseDec("1.0"),
//			true,
//			sdk.ZeroInt(),
//		},
//		{
//			"happy case #2",
//			amm.NewOrderBook(
//				newOrder(amm.Buy, utils.ParseDec("1.1"), sdk.NewInt(10000)),
//				newOrder(amm.Sell, utils.ParseDec("0.9"), sdk.NewInt(10000)),
//			),
//			utils.ParseDec("1.0"),
//			true,
//			sdk.ZeroInt(),
//		},
//		{
//			"positive quote coin dust",
//			amm.NewOrderBook(
//				newOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
//				newOrder(amm.Buy, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
//				newOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
//				newOrder(amm.Sell, utils.ParseDec("0.9999"), sdk.NewInt(1000)),
//			),
//			utils.ParseDec("0.9999"),
//			true,
//			sdk.NewInt(2),
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//			buyOrders := tc.os.BuyOrdersOver(tc.matchPrice)
//			sellOrders := tc.os.SellOrdersUnder(tc.matchPrice)
//			matchPrice, found := amm.FindMatchPrice(tc.os, int(defTickPrec))
//			if tc.matched {
//				require.True(t, found)
//			} else {
//				require.False(t, found)
//				return
//			}
//			require.True(sdk.DecEq(t, tc.matchPrice, matchPrice))
//			quoteCoinDust, matched := amm.MatchOrders(buyOrders, sellOrders, tc.matchPrice)
//			require.Equal(t, tc.matched, matched)
//			if matched {
//				require.True(sdk.IntEq(t, tc.quoteCoinDust, quoteCoinDust))
//				for _, order := range append(buyOrders, sellOrders...) {
//					if order.IsMatched() {
//						paid := order.GetOfferCoin().Sub(order.GetRemainingOfferCoin())
//						received := order.GetReceivedDemandCoin()
//						var effPrice sdk.Dec // Effective swap price
//						switch order.GetDirection() {
//						case amm.Buy:
//							effPrice = paid.Amount.ToDec().QuoInt(received.Amount)
//						case amm.Sell:
//							effPrice = received.Amount.ToDec().QuoInt(paid.Amount)
//						}
//						require.True(t, utils.DecApproxEqual(tc.matchPrice, effPrice))
//					}
//				}
//			}
//		})
//	}
//}
