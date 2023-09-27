package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestPlaceLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	msgServer := keeper.NewMsgServerImpl(s.keeper)
	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	resp, err := msgServer.PlaceLimitOrder(
		sdk.WrapSDKContext(s.Ctx), types.NewMsgPlaceLimitOrder(
			ordererAddr1, market.Id, true, utils.ParseDec("5.1"), sdk.NewInt(10_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.OrderId)
	s.AssertEqual(sdk.NewInt(0), resp.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("0uusd"), resp.Paid)
	s.AssertEqual(utils.ParseCoin("0ucre"), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"0"`),
		"paid":              []byte(`{"denom":"uusd","amount":"0"}`),
		"received":          []byte(`{"denom":"ucre","amount":"0"}`),
	})

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	resp, err = msgServer.PlaceLimitOrder(
		sdk.WrapSDKContext(s.Ctx), types.NewMsgPlaceLimitOrder(
			ordererAddr2, market.Id, false, utils.ParseDec("5"), sdk.NewInt(5_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(2, resp.OrderId)
	s.AssertEqual(sdk.NewInt(5_000000), resp.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_000000ucre"), resp.Paid)
	// Matched at 5.1
	s.AssertEqual(utils.ParseCoin("25_423500uusd"), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"5000000"`),
		"paid":              []byte(`{"denom":"ucre","amount":"5000000"}`),
		"received":          []byte(`{"denom":"uusd","amount":"25423500"}`),
	})
}

func (s *KeeperTestSuite) TestPlaceBatchLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	ordererBalances1Before := s.GetAllBalances(ordererAddr1)
	ordererBalances2Before := s.GetAllBalances(ordererAddr2)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewInt(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.AssertEqual(utils.ParseDec("5.05"), *marketState.LastPrice)
	ordererBalances1After := s.GetAllBalances(ordererAddr1)
	ordererBalances2After := s.GetAllBalances(ordererAddr2)
	ordererBalances1Diff, _ := ordererBalances1After.SafeSub(ordererBalances1Before)
	ordererBalances2Diff, _ := ordererBalances2After.SafeSub(ordererBalances2Before)
	// Both taker
	s.Require().Equal("9970000ucre,-50500000uusd", ordererBalances1Diff.String())
	s.Require().Equal("-10000000ucre,50348500uusd", ordererBalances2Diff.String())

	ordererBalances1Before = ordererBalances1After
	ordererBalances2Before = ordererBalances2After
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("5.2"), sdk.NewInt(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5.07"), sdk.NewInt(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState = s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.AssertEqual(utils.ParseDec("5.07"), *marketState.LastPrice)
	ordererBalances1After = s.GetAllBalances(ordererAddr1)
	ordererBalances2After = s.GetAllBalances(ordererAddr2)
	ordererBalances1Diff, _ = ordererBalances1After.SafeSub(ordererBalances1Before)
	ordererBalances2Diff, _ = ordererBalances2After.SafeSub(ordererBalances2Before)
	s.Require().Equal("9970000ucre,-50700000uusd", ordererBalances1Diff.String())
	// Maker
	s.Require().Equal("-10000000ucre,50623950uusd", ordererBalances2Diff.String())
}

func (s *KeeperTestSuite) TestPlaceMMLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")
	maxNumMMOrders := s.keeper.GetMaxNumMMOrders(s.Ctx)
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.1"), sdk.NewInt(100_000000), time.Hour)

	for i := uint32(0); i < maxNumMMOrders; i++ {
		price := utils.ParseDec("5").Sub(utils.ParseDec("0.001").MulInt64(int64(i)))
		s.PlaceMMLimitOrder(
			market.Id, ordererAddr1, true, price, sdk.NewInt(10_000000), time.Hour)
	}
	_, _, _, err := s.keeper.PlaceMMLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewInt(10_00000), time.Hour)
	s.Require().EqualError(err, "16 > 15: number of MM orders exceeded the limit")

	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("4.9"), sdk.NewInt(30_000000), 0)

	s.PlaceMMLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.9"), sdk.NewInt(10_00000), time.Hour)
}

func (s *KeeperTestSuite) TestPlaceMMBatchLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")
	maxNumMMOrders := s.keeper.GetMaxNumMMOrders(s.Ctx)
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.1"), sdk.NewInt(100_000000), time.Hour)

	for i := uint32(0); i < maxNumMMOrders; i++ {
		price := utils.ParseDec("5").Sub(utils.ParseDec("0.001").MulInt64(int64(i)))
		s.PlaceMMBatchLimitOrder(
			market.Id, ordererAddr1, true, price, sdk.NewInt(10_000000), time.Hour)
	}
	_, err := s.keeper.PlaceMMBatchLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewInt(10_00000), time.Hour)
	s.Require().EqualError(err, "16 > 15: number of MM orders exceeded the limit")

	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("4.9"), sdk.NewInt(30_000000), 0)
	s.NextBlock()

	s.PlaceMMBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.9"), sdk.NewInt(10_00000), time.Hour)
}

func (s *KeeperTestSuite) TestOrderMatching() {
	aliceAddr := s.FundedAccount(1, utils.ParseCoins("1000000ucre,1000000uusd"))
	bobAddr := s.FundedAccount(2, utils.ParseCoins("1000000ucre,1000000uusd"))

	market := s.CreateMarket("ucre", "uusd")

	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("100"), sdk.NewInt(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("99"), sdk.NewInt(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("97"), sdk.NewInt(1000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("98"), sdk.NewInt(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1001497ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("998500ucre,1149051uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins("146500uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))
	s.AssertEqual(utils.ParseCoins("3ucre,449uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetFeeCollectorAddress()))

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("96"), sdk.NewInt(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1002994ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("997000ucre,1295111uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins(""), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))
	s.AssertEqual(utils.ParseCoins("6ucre,889uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetFeeCollectorAddress()))
}

func (s *KeeperTestSuite) TestMinMaxPrice() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)
	ctx := s.Ctx
	s.Ctx, _ = ctx.CacheContext()
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, types.MinPrice, sdk.NewIntWithDecimal(1, 18), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, types.MaxPrice, sdk.NewInt(10000), time.Hour)
	s.Ctx, _ = ctx.CacheContext()
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, types.MinPrice, sdk.NewIntWithDecimal(1, 18), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, types.MaxPrice, sdk.NewInt(10000), time.Hour)
}

func (s *KeeperTestSuite) TestCancelOrder() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)

	_, err := s.keeper.CancelOrder(s.Ctx, ordererAddr, 1)
	s.Require().EqualError(err, "order not found: not found")

	balancesBefore := s.GetAllBalances(ordererAddr)
	_, order, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)
	s.Require().EqualValues(1, order.Id)
	_, err = s.keeper.CancelOrder(s.Ctx, ordererAddr, order.Id)
	s.Require().EqualError(err, "cannot cancel order placed in the same block: invalid request")

	s.NextBlock()
	s.CancelOrder(ordererAddr, order.Id)
	balancesAfter := s.GetAllBalances(ordererAddr)

	s.Require().Equal(balancesBefore, balancesAfter)
}

func (s *KeeperTestSuite) TestCancelAllOrders() {
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "uusd")
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("4.999"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("4.998"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.1"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.101"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.102"), sdk.NewInt(10_000000), time.Hour)

	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.1"), sdk.NewInt(10_00000), time.Hour)
	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.09"), sdk.NewInt(10_00000), time.Hour)
	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.08"), sdk.NewInt(10_00000), time.Hour)

	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("4.999"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("4.998"), sdk.NewInt(10_000000), time.Hour)

	s.NextBlock()

	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.2"), sdk.NewInt(10_00000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.21"), sdk.NewInt(10_00000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.22"), sdk.NewInt(10_00000), time.Hour)

	cancelledOrders := s.CancelAllOrders(ordererAddr1, market1.Id)
	s.Require().Len(cancelledOrders, 6)
	s.Require().EqualValues(1, cancelledOrders[0].Id)
	s.Require().EqualValues(2, cancelledOrders[1].Id)
	s.Require().EqualValues(3, cancelledOrders[2].Id)
	s.Require().EqualValues(4, cancelledOrders[3].Id)
	s.Require().EqualValues(5, cancelledOrders[4].Id)
	s.Require().EqualValues(6, cancelledOrders[5].Id)
}

func (s *KeeperTestSuite) TestFairMatching() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	ordererAddr3 := s.FundedAccount(3, enoughCoins)

	_, order1, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("1.2"), sdk.NewInt(10000), 0)
	_, order2, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.2"), sdk.NewInt(5000), 0)

	s.PlaceLimitOrder(
		market.Id, ordererAddr3, false, utils.ParseDec("1.1"), sdk.NewInt(9000), 0)

	order1, _ = s.keeper.GetOrder(s.Ctx, order1.Id)
	order2, _ = s.keeper.GetOrder(s.Ctx, order2.Id)

	s.AssertEqual(sdk.NewInt(4000), order1.OpenQuantity) // 9000*2/3 matched
	s.AssertEqual(sdk.NewInt(2000), order2.OpenQuantity) // 9000*1/3 matched

	s.NextBlock()

	_, order1, _ = s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("1.1"), sdk.NewInt(7000), 0)
	_, order2, _ = s.PlaceLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.1"), sdk.NewInt(3000), 0)

	s.PlaceMarketOrder(market.Id, ordererAddr3, false, sdk.NewInt(101))

	order1, _ = s.keeper.GetOrder(s.Ctx, order1.Id)
	order2, _ = s.keeper.GetOrder(s.Ctx, order2.Id)

	s.AssertEqual(sdk.NewInt(6929), order1.OpenQuantity) // 101*7/10 matched
	s.AssertEqual(sdk.NewInt(2970), order2.OpenQuantity) // 101*3/10 matched
}

func (s *KeeperTestSuite) TestNumMMOrdersEdgecase() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr := s.FundedAccount(1, enoughCoins)
	// Place 2 MM orders
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.9"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.85"), sdk.NewInt(10_000000), time.Hour)

	// Match against own orders
	s.PlaceMMLimitOrder(market.Id, ordererAddr, false, utils.ParseDec("4.5"), sdk.NewInt(30_000000), time.Hour)

	numMMOrders, _ := s.keeper.GetNumMMOrders(s.Ctx, ordererAddr, market.Id)
	// The number of MM orders should be 1, since previous order are fully matched
	// and deleted.
	s.Require().Equal(uint32(1), numMMOrders)
}

func (s *KeeperTestSuite) TestSwapEdgecase() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, ordererAddr1, utils.ParseDec("5"))

	for i := 0; i < 10; i++ {
		buyPrice := utils.ParseDec("5").Sub(utils.ParseDec("0.01").MulInt64(int64(i + 1)))
		sellPrice := utils.ParseDec("5").Add(utils.ParseDec("0.01").MulInt64(int64(i + 1)))
		s.PlaceLimitOrder(market.Id, ordererAddr1, true, buyPrice, sdk.NewInt(1_000000), time.Hour)
		s.PlaceLimitOrder(market.Id, ordererAddr1, false, sellPrice, sdk.NewInt(1_000000), time.Hour)
	}

	ordererAddr2 := s.FundedAccount(2, utils.ParseCoins("30_000000uusd"))
	s.SwapExactAmountIn(
		ordererAddr2, []uint64{market.Id}, utils.ParseCoin("30_000000uusd"), utils.ParseCoin("0ucre"), false)
}

func (s *KeeperTestSuite) TestPlaceMarketOrder_SellInsufficientFunds() {
	market := s.CreateMarket("ucre", "uusd")

	// Create last price.
	ordererAddr1 := s.FundedAccount(1, utils.ParseCoins("5_000000uusd"))
	ordererAddr2 := s.FundedAccount(2, utils.ParseCoins("1_200000ucre"))
	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)

	_, _, err := s.keeper.PlaceMarketOrder(s.Ctx, market.Id, ordererAddr2, false, sdk.NewInt(500000))
	s.Require().EqualError(err, "200000ucre is smaller than 500000ucre: insufficient funds")
}

func (s *KeeperTestSuite) TestMaxOrderPriceRatio() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)

	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().Equal(int64(1), marketState.LastMatchingHeight)
	s.AssertEqual(utils.ParseDec("5"), *marketState.LastPrice)

	// 5.6 > 5 * 1.1 (not allowed for buy orders)
	_, _, _, err := s.keeper.PlaceLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.6"), sdk.NewInt(1_000000), time.Hour)
	s.Require().EqualError(err, "price is higher than the limit 5.500000000000000000: order price out of range")
	// 4 < 5 * 0.9 (allowed for buy orders)
	s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4"), sdk.NewInt(1_000000), time.Hour)

	// 4.4 < 5 * 0.9 (not allowed for sell orders)
	_, _, _, err = s.keeper.PlaceLimitOrder(
		s.Ctx, market.Id, ordererAddr2, false, utils.ParseDec("4.4"), sdk.NewInt(1_000000), time.Hour)
	s.Require().EqualError(err, "price is lower than the limit 4.500000000000000000: order price out of range")
	// 6 > 5 * 1.1 (allowed for sell orders)
	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("6"), sdk.NewInt(1_000000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.5"), sdk.NewInt(1_000000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5.5"), sdk.NewInt(1_000000), time.Hour)

	_, res := s.PlaceMarketOrder(market.Id, ordererAddr1, true, sdk.NewInt(2_000000))
	s.Require().True(res.IsMatched())
	// Order at 6 not matched because of MaxOrderPriceRatio
	s.AssertEqual(utils.ParseDec("5.5"), res.LastPrice)
	s.AssertEqual(sdk.NewInt(1_000000), res.ExecutedQuantity)

	// Reset last price to 5
	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(1_000000), time.Hour)

	_, res = s.PlaceMarketOrder(market.Id, ordererAddr2, false, sdk.NewInt(2_000000))
	s.Require().True(res.IsMatched())
	// Order at 4 not matched because of MaxOrderPriceRatio
	s.AssertEqual(utils.ParseDec("4.5"), res.LastPrice)
	s.AssertEqual(sdk.NewInt(1_000000), res.ExecutedQuantity)
}
