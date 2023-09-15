package keeper_test

//func (s *KeeperTestSuite) TestAfterOrdersExecuted_NegativeAmountInDiffEdgecase() {
//	r := rand.New(rand.NewSource(1))
//	enoughCoins := sdk.NewCoins(
//		sdk.NewCoin("ucre", sdk.NewIntWithDecimal(1, 60)),
//		sdk.NewCoin("uusd", sdk.NewIntWithDecimal(1, 60)))
//
//	for i := 0; i < 20; i++ {
//		r := rand.New(rand.NewSource(r.Int63()))
//
//		s.SetupTest()
//		s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
//
//		currentPrice := sdk.NewDec(2).Add(simtypes.RandomDecAmount(r, sdk.NewDec(7)))
//		market, pool := s.CreateMarketAndPool("ucre", "uusd", currentPrice)
//
//		lpAddr := s.FundedAccount(1, enoughCoins)
//		s.MakeLastPrice(
//			market.Id, lpAddr,
//			exchangetypes.PriceAtTick(exchangetypes.TickAtPrice(currentPrice)))
//
//		liquidity := sdk.NewIntWithDecimal(1, 30)
//		s.AddLiquidityByLiquidity(lpAddr, pool.Id, types.MinPrice, types.MaxPrice, liquidity)
//
//		ordererAddr := s.FundedAccount(2, enoughCoins)
//
//		for j := 0; j < 500; j++ {
//			isBuy := r.Float64() < 0.5
//			obs := s.App.ExchangeKeeper.ConstructMemOrderBookSide(s.Ctx, market, exchangetypes.MemOrderBookSideOptions{
//				IsBuy:             !isBuy,
//				MaxNumPriceLevels: 5,
//			}, nil)
//			if len(obs.Levels()) == 0 {
//				continue
//			}
//			var qty sdk.Dec
//			if r.Float64() < 0.5 {
//				qty = obs.Levels()[0].Orders()[0].OpenQuantity()
//			} else {
//				t := obs.Levels()[0].Orders()[0].OpenQuantity()
//				qty = utils.RandomDec(r, t.Mul(utils.ParseDec("0.99")), t)
//			}
//			qty = qty.TruncateDec()
//			s.PlaceMarketOrder(market.Id, ordererAddr, isBuy, qty)
//		}
//	}
//}
