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
			ordererAddr1, market.Id, true, utils.ParseDec("5.1"), sdk.NewDec(10_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.OrderId)
	s.Require().Equal(sdk.NewDec(0), resp.ExecutedQuantity)
	s.Require().Equal(sdk.NewInt64DecCoin("uusd", 0), resp.Paid)
	s.Require().Equal(sdk.NewInt64DecCoin("ucre", 0), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"0.000000000000000000"`),
		"paid":              []byte(`{"denom":"uusd","amount":"0.000000000000000000"}`),
		"received":          []byte(`{"denom":"ucre","amount":"0.000000000000000000"}`),
	})

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	resp, err = msgServer.PlaceLimitOrder(
		sdk.WrapSDKContext(s.Ctx), types.NewMsgPlaceLimitOrder(
			ordererAddr2, market.Id, false, utils.ParseDec("5"), sdk.NewDec(5_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(2, resp.OrderId)
	s.Require().Equal(sdk.NewDec(5_000000), resp.ExecutedQuantity)
	s.Require().Equal(sdk.NewInt64DecCoin("ucre", 5_000000), resp.Paid)
	// Matched at 5.1
	s.Require().Equal(sdk.NewInt64DecCoin("uusd", 25_423500), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"5000000.000000000000000000"`),
		"paid":              []byte(`{"denom":"ucre","amount":"5000000.000000000000000000"}`),
		"received":          []byte(`{"denom":"uusd","amount":"25423500.000000000000000000"}`),
	})
}

func (s *KeeperTestSuite) TestPlaceBatchLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	ordererBalances1Before := s.GetAllBalances(ordererAddr1)
	ordererBalances2Before := s.GetAllBalances(ordererAddr2)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewDec(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewDec(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.Require().Equal("5.050000000000000000", marketState.LastPrice.String())
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
		market.Id, ordererAddr1, true, utils.ParseDec("5.2"), sdk.NewDec(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5.07"), sdk.NewDec(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState = s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.Require().Equal("5.070000000000000000", marketState.LastPrice.String())
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
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.1"), sdk.NewDec(100_000000), time.Hour)

	for i := uint32(0); i < maxNumMMOrders; i++ {
		price := utils.ParseDec("5").Sub(utils.ParseDec("0.001").MulInt64(int64(i)))
		s.PlaceMMLimitOrder(
			market.Id, ordererAddr1, true, price, sdk.NewDec(10_000000), time.Hour)
	}
	_, _, _, err := s.keeper.PlaceMMLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewDec(10_00000), time.Hour)
	s.Require().EqualError(err, "16 > 15: number of MM orders exceeded the limit")

	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("4.9"), sdk.NewDec(30_000000), 0)

	s.PlaceMMLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.9"), sdk.NewDec(10_00000), time.Hour)
}

func (s *KeeperTestSuite) TestPlaceMMBatchLimitOrder() {
	market := s.CreateMarket("ucre", "uusd")
	maxNumMMOrders := s.keeper.GetMaxNumMMOrders(s.Ctx)
	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5.1"), sdk.NewDec(100_000000), time.Hour)

	for i := uint32(0); i < maxNumMMOrders; i++ {
		price := utils.ParseDec("5").Sub(utils.ParseDec("0.001").MulInt64(int64(i)))
		s.PlaceMMBatchLimitOrder(
			market.Id, ordererAddr1, true, price, sdk.NewDec(10_000000), time.Hour)
	}
	_, err := s.keeper.PlaceMMBatchLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewDec(10_00000), time.Hour)
	s.Require().EqualError(err, "16 > 15: number of MM orders exceeded the limit")

	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("4.9"), sdk.NewDec(30_000000), 0)
	s.NextBlock()

	s.PlaceMMBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.9"), sdk.NewDec(10_00000), time.Hour)
}

func (s *KeeperTestSuite) TestOrderMatching() {
	aliceAddr := s.FundedAccount(1, utils.ParseCoins("1000000ucre,1000000uusd"))
	bobAddr := s.FundedAccount(2, utils.ParseCoins("1000000ucre,1000000uusd"))

	market := s.CreateMarket("ucre", "uusd")

	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("100"), sdk.NewDec(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("99"), sdk.NewDec(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("97"), sdk.NewDec(1000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("98"), sdk.NewDec(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1001497ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("998500ucre,1149051uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins("3ucre,146949uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("96"), sdk.NewDec(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1002994ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("997000ucre,1295111uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins("6ucre,889uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))
}

func (s *KeeperTestSuite) TestMinMaxPrice() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)
	maxPrice := types.MaxPrice
	for price := types.MinPrice; price.LT(maxPrice); price = price.MulInt64(10) {
		s.PlaceLimitOrder(
			market.Id, ordererAddr, false, price, sdk.NewDec(1000000), time.Hour)
	}

	// HERE! No need to check something?
}

func (s *KeeperTestSuite) TestCancelOrder() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)

	_, err := s.keeper.CancelOrder(s.Ctx, ordererAddr, 1)
	s.Require().EqualError(err, "order not found: not found")

	balancesBefore := s.GetAllBalances(ordererAddr)
	_, order, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("5"), sdk.NewDec(10_000000), time.Hour)
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

	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("4.999"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, true, utils.ParseDec("4.998"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.1"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.101"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.102"), sdk.NewDec(10_000000), time.Hour)

	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.1"), sdk.NewDec(10_00000), time.Hour)
	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.09"), sdk.NewDec(10_00000), time.Hour)
	s.PlaceLimitOrder(market2.Id, ordererAddr1, true, utils.ParseDec("10.08"), sdk.NewDec(10_00000), time.Hour)

	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("5"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("4.999"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr2, true, utils.ParseDec("4.998"), sdk.NewDec(10_000000), time.Hour)

	s.NextBlock()

	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.2"), sdk.NewDec(10_00000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.21"), sdk.NewDec(10_00000), time.Hour)
	s.PlaceLimitOrder(market1.Id, ordererAddr1, false, utils.ParseDec("5.22"), sdk.NewDec(10_00000), time.Hour)

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
		market.Id, ordererAddr1, true, utils.ParseDec("1.2"), sdk.NewDec(10000), 0)
	_, order2, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.2"), sdk.NewDec(5000), 0)

	s.PlaceLimitOrder(
		market.Id, ordererAddr3, false, utils.ParseDec("1.1"), sdk.NewDec(9000), 0)

	order1, _ = s.keeper.GetOrder(s.Ctx, order1.Id)
	order2, _ = s.keeper.GetOrder(s.Ctx, order2.Id)

	s.AssertEqual(utils.ParseDec("4000"), order1.OpenQuantity) // 9000*2/3 matched
	s.AssertEqual(utils.ParseDec("2000"), order2.OpenQuantity) // 9000*1/3 matched

	s.NextBlock()

	_, order1, _ = s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("1.1"), sdk.NewDec(7000), 0)
	_, order2, _ = s.PlaceLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.1"), sdk.NewDec(3000), 0)

	s.PlaceMarketOrder(market.Id, ordererAddr3, false, sdk.NewDec(101))

	order1, _ = s.keeper.GetOrder(s.Ctx, order1.Id)
	order2, _ = s.keeper.GetOrder(s.Ctx, order2.Id)

	s.AssertEqual(utils.ParseDec("6929.3"), order1.OpenQuantity) // 101*7/10 matched
	s.AssertEqual(utils.ParseDec("2969.7"), order2.OpenQuantity) // 101*3/10 matched
}

func (s *KeeperTestSuite) TestDecQuantity() {
	market := s.CreateMarket("ucre", "uusd")
	// Set the fees to zero for now.
	market.MakerFeeRate = sdk.ZeroDec()
	market.TakerFeeRate = sdk.ZeroDec()
	s.keeper.SetMarket(s.Ctx, market)

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	ordererAddr3 := s.FundedAccount(3, enoughCoins)

	_, order1, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("0.14076"), sdk.NewDec(7000), time.Hour)
	_, order2, _ := s.PlaceLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("0.14076"), sdk.NewDec(3000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, ordererAddr3, false, utils.ParseDec("0.14"), sdk.NewDec(1001), 0)

	order1, _ = s.keeper.GetOrder(s.Ctx, order1.Id)
	order2, _ = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.AssertEqual(utils.ParseDec("6299.3"), order1.OpenQuantity)
	s.AssertEqual(utils.ParseDec("2699.7"), order2.OpenQuantity)

	// Cancel the first order.
	s.NextBlock()
	s.CancelOrder(ordererAddr1, order1.Id)

	orderer3BalancesBefore := s.GetAllBalances(ordererAddr3)
	// Fill the rest.
	s.PlaceMarketOrder(market.Id, ordererAddr3, false, sdk.NewDec(10000))
	orderer3BalancesAfter := s.GetAllBalances(ordererAddr3)

	// order2 remaining deposit = 380uusd
	// order2 executable quantity ~= 2699.6305768
	// order2 executable quantity * 0.14076 ~= 379.9999999
	diff, _ := orderer3BalancesAfter.SafeSub(orderer3BalancesBefore)
	s.AssertEqual(sdk.NewInt(379), diff.AmountOf("uusd"))
}

func (s *KeeperTestSuite) TestNumMMOrdersEdgecase() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr := s.FundedAccount(1, enoughCoins)
	// Place 2 MM orders
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.9"), sdk.NewDec(10_000000), time.Hour)
	s.PlaceMMLimitOrder(market.Id, ordererAddr, true, utils.ParseDec("4.85"), sdk.NewDec(10_000000), time.Hour)

	// Match against own orders
	s.PlaceMMLimitOrder(market.Id, ordererAddr, false, utils.ParseDec("4.5"), sdk.NewDec(30_000000), time.Hour)

	numMMOrders, _ := s.keeper.GetNumMMOrders(s.Ctx, ordererAddr, market.Id)
	// The number of MM orders should be 1, since previous order are fully matched
	// and deleted.
	s.Require().Equal(uint32(1), numMMOrders)
}

// HERE! Test for the limit on the MM order number
func (s *KeeperTestSuite) TestLimitNumMMOrders() {

}

func (s *KeeperTestSuite) TestSwapEdgecase() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, ordererAddr1, utils.ParseDec("5"))

	for i := 0; i < 10; i++ {
		buyPrice := utils.ParseDec("5").Sub(utils.ParseDec("0.01").MulInt64(int64(i + 1)))
		sellPrice := utils.ParseDec("5").Add(utils.ParseDec("0.01").MulInt64(int64(i + 1)))
		s.PlaceLimitOrder(market.Id, ordererAddr1, true, buyPrice, sdk.NewDec(1_000000), time.Hour)
		s.PlaceLimitOrder(market.Id, ordererAddr1, false, sellPrice, sdk.NewDec(1_000000), time.Hour)
	}

	ordererAddr2 := s.FundedAccount(2, utils.ParseCoins("30_000000uusd"))
	s.SwapExactAmountIn(
		ordererAddr2, []uint64{market.Id}, utils.ParseDecCoin("30_000000uusd"), utils.ParseDecCoin("0ucre"), false)
}

func (s *KeeperTestSuite) TestPlaceMarketOrder_SellInsufficientFunds() {
	market := s.CreateMarket("ucre", "uusd")

	// Create last price.
	ordererAddr1 := s.FundedAccount(1, utils.ParseCoins("5_000000uusd"))
	ordererAddr2 := s.FundedAccount(2, utils.ParseCoins("1_200000ucre"))
	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)

	_, _, err := s.keeper.PlaceMarketOrder(s.Ctx, market.Id, ordererAddr2, false, sdk.NewDec(500000))
	s.Require().EqualError(err, "200000ucre is smaller than 500000.000000000000000000ucre: insufficient funds")
}

func (s *KeeperTestSuite) TestMaxOrderPriceRatio() {
	market := s.CreateMarket("ucre", "uusd")

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)

	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().Equal(int64(1), marketState.LastMatchingHeight)
	s.AssertEqual(utils.ParseDec("5"), *marketState.LastPrice)

	// 5.6 > 5 * 1.1 (not allowed for buy orders)
	_, _, _, err := s.keeper.PlaceLimitOrder(
		s.Ctx, market.Id, ordererAddr1, true, utils.ParseDec("5.6"), sdk.NewDec(1_000000), time.Hour)
	s.Require().EqualError(err, "price is higher than the limit 5.500000000000000000: order price out of range")
	// 4 < 5 * 0.9 (allowed for buy orders)
	s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4"), sdk.NewDec(1_000000), time.Hour)

	// 4.4 < 5 * 0.9 (not allowed for sell orders)
	_, _, _, err = s.keeper.PlaceLimitOrder(
		s.Ctx, market.Id, ordererAddr2, false, utils.ParseDec("4.4"), sdk.NewDec(1_000000), time.Hour)
	s.Require().EqualError(err, "price is lower than the limit 4.500000000000000000: order price out of range")
	// 6 > 5 * 1.1 (allowed for sell orders)
	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("6"), sdk.NewDec(1_000000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("4.5"), sdk.NewDec(1_000000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5.5"), sdk.NewDec(1_000000), time.Hour)

	_, res := s.PlaceMarketOrder(market.Id, ordererAddr1, true, sdk.NewDec(2_000000))
	s.Require().True(res.Executed())
	// Order at 6 not matched because of MaxOrderPriceRatio
	s.AssertEqual(utils.ParseDec("5.5"), res.LastPrice)
	s.AssertEqual(sdk.NewDec(1_000000), res.ExecutedQuantity)

	// Reset last price to 5
	s.PlaceLimitOrder(market.Id, ordererAddr1, true, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewDec(1_000000), time.Hour)

	_, res = s.PlaceMarketOrder(market.Id, ordererAddr2, false, sdk.NewDec(2_000000))
	s.Require().True(res.Executed())
	// Order at 4 not matched because of MaxOrderPriceRatio
	s.AssertEqual(utils.ParseDec("4.5"), res.LastPrice)
	s.AssertEqual(sdk.NewDec(1_000000), res.ExecutedQuantity)
}
