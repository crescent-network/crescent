package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestBatchMatchingWithLastPrice() {
	market := s.CreateMarket("ucre", "uusd") // Default Fee = Maker: 0.15%, Taker: 0.3%, Order source: Taker * 50%

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("1"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(2, enoughCoins)
	ordererAddr2 := s.FundedAccount(3, enoughCoins)

	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1"), sdk.NewInt(5_000000), time.Hour)
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.001"), sdk.NewInt(10_000000), time.Hour)
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.002"), sdk.NewInt(10_000000), time.Hour)

	order4 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1"), sdk.NewInt(10_000000), time.Hour)
	order5 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.01"), sdk.NewInt(1_000000), time.Hour)

	// Order book looks like:
	//                | 1.010 | # (5)
	// (3) ########## | 1.002 |
	// (2) ########## | 1.001 |
	// (1)      ##### | 1.000 | ########## (4) <-- last price

	// After batch matching, it should look like(phase 1 only):
	//                | 1.010 |
	// (3) ########## | 1.002 |
	// (2) ########## | 1.001 |
	//                | 1.000 | ###### (4)    <-- last price

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order4.Id)
	s.AssertEqual(sdk.NewInt(4_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("4_000000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("3_988000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("12000ucre"))

	ev = s.getEventOrderFilled(order5.Id)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_000000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("997000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("3000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order3.RemainingDeposit)

	order4 = s.keeper.MustGetOrder(s.Ctx, order4.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(6_000000), order4.OpenQuantity)
	s.AssertEqual(sdk.NewInt(6_000000), order4.RemainingDeposit)

	_, found = s.keeper.GetOrder(s.Ctx, order5.Id)
	s.Require().False(found)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("15000ucre"), utils.ParseCoin("15000uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)
}

// Test for the case where the last price does not exist
func (s *KeeperTestSuite) TestBatchMatchingWithoutLastPrice1() {
	market := s.CreateMarket("ucre", "uusd") // Default Fee = Maker: 0.15%, Taker: 0.3%, Order source: Taker * 50%

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0001"), sdk.NewInt(6_000000), time.Hour)
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)
	order4 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0100"), sdk.NewInt(1_000000), time.Hour)

	// Order book looks like:
	//                | 1.0100 | # (4)
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	//     (1) ###### | 1.0001 |

	// After batch matching, it should look like: single price auction
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ##### (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |      <-- last price = (1.0040 + 1.0001)/2 = 1.00205 --> banker's rounding
	//                | 1.0001 |

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0020"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(6_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("6_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("5_993964uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("18036uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_010000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000ucre"))

	ev = s.getEventOrderFilled(order4.Id)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_002000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("997000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("3000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(5_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(5_030000), order3.RemainingDeposit)

	_, found = s.keeper.GetOrder(s.Ctx, order4.Id)
	s.Require().False(found)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("18000ucre"), utils.ParseCoin("18036uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)
}

func (s *KeeperTestSuite) TestBatchMatchingWithoutLastPrice2() {
	market := s.CreateMarket("ucre", "uusd")

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)
	order4 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0100"), sdk.NewInt(1_000000), time.Hour)

	// Order book looks like:
	//                | 1.0100 | #   (4)
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	// (1)     ###### | 1.0003 |

	// After batch matching, it should look like: single price auction
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ##### (3)
	//                | 1.0022 |      <-- last price = (1.0040 + 1.0003)/2 = 1.00215 --> banker's rounding
	//                | 1.0021 |
	//                | 1.0020 |
	//                | 1.0003 |

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0022"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(6_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("6_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("5_995160uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("18040uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_011000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000ucre"))

	ev = s.getEventOrderFilled(order4.Id)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_002200uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("997000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("3000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(5_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(5_029000), order3.RemainingDeposit)

	_, found = s.keeper.GetOrder(s.Ctx, order4.Id)
	s.Require().False(found)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("18000ucre"), utils.ParseCoin("18040uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)
}

func (s *KeeperTestSuite) TestBatchMatchingWithLastPrice099() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("0.99"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(2, enoughCoins)
	ordererAddr2 := s.FundedAccount(3, enoughCoins)

	// Order book looks like:
	//                | 1.0100 | #   (4)
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	// (1)     ###### | 1.0003 |
	//                | 0.9900 |  --> last price

	// After batch matching, it should look like:
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ##### (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	//                | 1.0003 |           <-- last price (maker's price)

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)
	order4 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0100"), sdk.NewInt(1_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0003"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(6_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("6_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("5_992797uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("9003uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_001500uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000ucre"))

	ev = s.getEventOrderFilled(order4.Id)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_000300uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("997000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("3000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(5_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(5_038500), order3.RemainingDeposit)

	_, found = s.keeper.GetOrder(s.Ctx, order4.Id)
	s.Require().False(found)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("18000ucre"), utils.ParseCoin("9003uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)
}

func (s *KeeperTestSuite) TestBatchMatchingWithLastPrice1002() {
	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("1.002"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(2, enoughCoins)
	ordererAddr2 := s.FundedAccount(3, enoughCoins)

	// Order book looks like:
	//                | 1.0100 | #   (4)
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |  --> last price
	// (1)     ###### | 1.0003 |
	//                | 0.9900 |

	// After batch matching, it should look like:
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ##### (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |  --> last price
	//                | 1.0003 |

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)
	order4 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0100"), sdk.NewInt(1_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.002"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(6_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("6_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("5_993964uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("18036uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(5_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_010000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000ucre"))

	ev = s.getEventOrderFilled(order4.Id)
	s.AssertEqual(sdk.NewInt(1_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("1_002000uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("997000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("3000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(5_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(5_030000), order3.RemainingDeposit)

	_, found = s.keeper.GetOrder(s.Ctx, order4.Id)
	s.Require().False(found)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("18000ucre"), utils.ParseCoin("18036uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)
}

func (s *KeeperTestSuite) TestBatchMatchingWithOrderSourceOrderWithoutMatchingWithoutLastPrice() {
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.002"), sdk.NewInt(1_000000)),
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.001"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// Order book looks like:
	//                | 1.0200 |
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	// (1)     ###### | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0020 | #    (os)  ---> order source order
	//                | 1.0010 | ##   (os)  ---> order source order
	//                | 1.0000 |

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0060"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.Require().Nil(s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.Require().Nil(ev)

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.Require().Equal((*types.EventOrderSourceOrdersFilled)(nil), ev_os)

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(6_000000), order1.OpenQuantity)
	s.AssertEqual(sdk.NewInt(6_000000), order1.RemainingDeposit)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_040000), order3.RemainingDeposit)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff := osBalanceAfterMatching.Sub(osBalanceBeforeMatching)
	expectedOsBalancDiff := sdk.Coins{}
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestBatchMatchingWithOrderSourceOrderWithoutMatchingWithLastPrice() {
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.002"), sdk.NewInt(1_000000)),
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.001"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("1.004"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)
	expectedFee := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// Order book looks like:
	//                | 1.0200 |
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	// (1)     ###### | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0020 | #    (os)  ---> order source order
	//                | 1.0010 | ##   (os)  ---> order source order
	//                | 1.0000 |

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0060"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.004"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.Require().Nil(ev)

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.Require().Equal((*types.EventOrderSourceOrdersFilled)(nil), ev_os)

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(6_000000), order1.OpenQuantity)
	s.AssertEqual(sdk.NewInt(6_000000), order1.RemainingDeposit)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_040000), order3.RemainingDeposit)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff := osBalanceAfterMatching.Sub(osBalanceBeforeMatching)
	expectedOsBalancDiff := sdk.Coins{}
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestBatchMatchingWithOrderSourceOrderWithoutLastPrice() {
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.01"), sdk.NewInt(1_000000)),
		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.02"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)
	expectedFee := sdk.Coins{}
	expectedOsBalanceDiffManual := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// Order book looks like:
	//                | 1.0200 | ##   (os)  ---> order source order (taker)
	//                | 1.0100 | #    (os)  ---> order source order (taker)
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	// (1)     ###### | 1.0003 |

	// After batch matching, it should look like: single price auction
	//                | 1.0200 |
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ####### (3)
	//                | 1.0022 |      <-- last price = (1.0040 + 1.0003)/2 = 1.00215 --> banker's rounding
	//                | 1.0021 |
	//                | 1.0020 |
	//                | 1.0003 |

	// make sell orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000), time.Hour) // only difference from the above
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0022"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(6_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("6_000000ucre"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("5_995160uusd"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("18040uusd"))

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(3_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("3_006600uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("2_991000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("9000ucre"))

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(3_000000), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("3_006600uusd"), ev_os.Paid)
	s.AssertEqual(utils.ParseCoin("3_000000ucre"), ev_os.Received)
	expectedOsBalanceDiffManual, _ = expectedOsBalanceDiffManual.SafeSub(sdk.Coins{utils.ParseCoin("3_006600uusd")})
	expectedOsBalanceDiffManual = expectedOsBalanceDiffManual.Add(utils.ParseCoin("3_000000ucre"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(7_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(7_033400), order3.RemainingDeposit)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("18040uusd"), utils.ParseCoin("9000ucre")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancDiff := sdk.Coins{sdk.Coin{Denom: "uusd", Amount: sdk.NewInt(-3006600)},
		utils.ParseCoin("3_000000ucre")}
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
	s.AssertEqual(expectedOsBalanceDiffManual, osBalanceDiff)
}

func (s *KeeperTestSuite) TestBatchMatchingWithOrderSourceOrderAsMaker() {
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("0.99"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)
	expectedFee := sdk.Coins{}
	expectedOsBalanceDiffManual := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	ordererAddr3 := s.FundedAccount(3, enoughCoins)

	// Order book looks like:
	//                | 1.0200 | ###   (1) (taker)
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ########## (3) (taker)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	// (os)    ###### | 1.0003 |   (maker)
	//                | 0.9900 |      <-- last price

	// After batch matching, it should look like:
	//                | 1.0200 |
	//                | 1.0100 |
	// (2) ########## | 1.0080 |
	//                | 1.0060 |
	//                | 1.0040 | ####### (3)
	//                | 1.0022 |
	//                | 1.0021 |
	//                | 1.0020 |
	//                | 1.0003 |      --> new last price

	// make sell orders
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

	// make buy orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("1.02"), sdk.NewInt(3_000000), time.Hour)
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr3, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0003"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(3_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("3_000900uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("2_991000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("4500ucre")) // half fee due to order source order

	ev = s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev)

	ev = s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(3_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("3_000900uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("2_991000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("4500ucre")) // half fee due to order source order

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(6_000000), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_991000ucre"), ev_os.Paid)
	s.AssertEqual(utils.ParseCoin("6_001800uusd"), ev_os.Received)
	expectedOsBalanceDiffManual, _ = expectedOsBalanceDiffManual.SafeSub(sdk.Coins{utils.ParseCoin("5_991000ucre")})
	expectedOsBalanceDiffManual = expectedOsBalanceDiffManual.Add(utils.ParseCoin("6_001800uusd"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(7_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(7_039100), order3.RemainingDeposit)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("9000ucre")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancDiff := sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-5991000)},
		utils.ParseCoin("6_001800uusd")}
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
	s.AssertEqual(expectedOsBalanceDiffManual, osBalanceDiff)
}

func (s *KeeperTestSuite) TestBatchMatchingEdgeCase() {
	// Order book looks like:
	//                | 1.0200 | ##########   (1) (taker)
	//                | 1.0100 |
	//                | 1.0080 |
	//                | 1.0060 |
	// (os)       #.1 | 1.0040 | #####.2      (2) (taker)
	// (os)     ##### | 1.0022 |
	// (os)      #### | 1.0021 |
	// (os)       ### | 1.0020 |
	// (os)        ## | 1.0003 |
	//                | 0.9900 | ##          (os)         <-- last price

	// After batch matching, it should look like:
	//                | 1.0200 | ##########   (1) (taker)
	//                | 1.0100 |
	//                | 1.0080 |
	//                | 1.0060 |
	// (os)       #.1 | 1.0040 | #####.2      (2) (taker)  <-- new last price
	// (os)     ##### | 1.0022 |
	// (os)      #### | 1.0021 |
	// (os)       ### | 1.0020 |
	// (os)        ## | 1.0003 |
	//                | 0.9900 | ##          (os)
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, utils.ParseDec("0.9900"), sdk.NewInt(2_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0040"), sdk.NewInt(1_000001)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0022"), sdk.NewInt(5_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0021"), sdk.NewInt(4_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0020"), sdk.NewInt(3_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0003"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("0.99"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)
	expectedFee := sdk.Coins{}
	expectedOsBalanceDiffManual := sdk.Coins{}

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	// make buy orders
	order1 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("1.02"), sdk.NewInt(10_000000), time.Hour)
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, true, utils.ParseDec("1.0040"), sdk.NewInt(5_000002), time.Hour)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	s.AssertEqual(utils.ParseDec("1.0040"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	ev := s.getEventOrderFilled(order1.Id)
	s.AssertEqual(sdk.NewInt(10_000000), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("10_017200uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("9_970000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("15000ucre")) // half fee due to order source order

	ev = s.getEventOrderFilled(order2.Id)
	s.AssertEqual(sdk.NewInt(5_000001), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_012802uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("4_985000ucre"), ev.Received)
	expectedFee = expectedFee.Add(utils.ParseCoin("7501ucre"), utils.ParseCoin("1uusd")) // half fee due to order source order + dust

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(15_000001), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("14_977501ucre"), ev_os.Paid)
	s.AssertEqual(utils.ParseCoin("15_030001uusd"), ev_os.Received)
	expectedOsBalanceDiffManual, _ = expectedOsBalanceDiffManual.SafeSub(sdk.Coins{utils.ParseCoin("14_977501ucre")})
	expectedOsBalanceDiffManual = expectedOsBalanceDiffManual.Add(utils.ParseCoin("15_030001uusd"))

	_, found := s.keeper.GetOrder(s.Ctx, order1.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(1), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(7201), order2.RemainingDeposit)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	expectedFeeManual := sdk.Coins{utils.ParseCoin("22501ucre"), utils.ParseCoin("1uusd")}
	s.AssertEqual(expectedFee, feeAmount)
	s.AssertEqual(expectedFeeManual, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancDiff := sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-14977501)},
		utils.ParseCoin("15_030001uusd")}
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
	s.AssertEqual(expectedOsBalanceDiffManual, osBalanceDiff)
}

// func (s *KeeperTestSuite) TestBatchMatchingWithMultipleOrderSourceOrder() {
// 	os1 := types.NewMockOrderSource(
// 		"mockOrderSource1",
// 		types.NewMockOrderSourceOrder(false, utils.ParseDec("1.0003"), sdk.NewInt(6_000000)))
// 	os2 := types.NewMockOrderSource(
// 		"mockOrderSource2",
// 		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.01"), sdk.NewInt(1_000000)),
// 		types.NewMockOrderSourceOrder(true, utils.ParseDec("1.02"), sdk.NewInt(2_000000)))
// 	s.FundAccount(os1.Address, enoughCoins)
// 	s.FundAccount(os2.Address, enoughCoins)
// 	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os1, os2)
// 	s.keeper = s.App.ExchangeKeeper

// 	market := s.CreateMarket("ucre", "uusd")
// 	mmAddr := s.FundedAccount(1, enoughCoins)
// 	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("0.99"))

// 	feeCollector := market.MustGetFeeCollectorAddress()
// 	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
// 	os1BalanceBeforeMatching := s.GetAllBalances(os1.Address)
// 	os2BalanceBeforeMatching := s.GetAllBalances(os1.Address)
// 	expectedFee := sdk.Coins{}
// 	expectedOs1BalanceDiffManual := sdk.Coins{}
// 	expectedOs2BalanceDiffManual := sdk.Coins{}

// 	ordererAddr2 := s.FundedAccount(2, enoughCoins)
// 	ordererAddr3 := s.FundedAccount(3, enoughCoins)

// 	// Order book looks like:
// 	//                | 1.0200 | ##   (os2) (taker)
// 	//                | 1.0100 | #    (os2) (taker)
// 	// (2) ########## | 1.0080 |
// 	//                | 1.0060 |
// 	//                | 1.0040 | ########## (3) (taker)
// 	//                | 1.0022 |
// 	//                | 1.0021 |
// 	//                | 1.0020 |
// 	// (os1)   ###### | 1.0003 |   (maker)
// 	//                | 0.9900 |      <-- last price

// 	// After batch matching, it should look like:
// 	//                | 1.0200 |
// 	//                | 1.0100 |
// 	// (2) ########## | 1.0080 |
// 	//                | 1.0060 |
// 	//                | 1.0040 | ####### (3)
// 	//                | 1.0022 |
// 	//                | 1.0021 |
// 	//                | 1.0020 |
// 	//                | 1.0003 |      --> new last price

// 	// make sell orders
// 	order2 := s.PlaceBatchLimitOrder(
// 		market.Id, ordererAddr2, false, utils.ParseDec("1.0080"), sdk.NewInt(10_000000), time.Hour)

// 	// make buy orders
// 	order3 := s.PlaceBatchLimitOrder(
// 		market.Id, ordererAddr3, true, utils.ParseDec("1.0040"), sdk.NewInt(10_000000), time.Hour)

// 	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
// 	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

// 	s.AssertEqual(utils.ParseDec("1.0003"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

// 	ev := s.getEventOrderFilled(order2.Id)
// 	s.Require().Nil(ev)

// 	ev = s.getEventOrderFilled(order3.Id)
// 	s.AssertEqual(sdk.NewInt(3_000000), ev.ExecutedQuantity)
// 	s.AssertEqual(utils.ParseCoin("3_000900uusd"), ev.Paid)
// 	s.AssertEqual(utils.ParseCoin("2_991000ucre"), ev.Received)
// 	expectedFee = expectedFee.Add(utils.ParseCoin("4500ucre")) // half fee due to order source order

// 	ev_os1 := s.getEventOrderSourceOrdersFilled("mockOrderSource1") // maker
// 	s.AssertEqual(sdk.NewInt(6_000000), ev_os1.ExecutedQuantity)
// 	s.AssertEqual(utils.ParseCoin("5_995500ucre"), ev_os1.Paid)
// 	s.AssertEqual(utils.ParseCoin("6_001800uusd"), ev_os1.Received)
// 	expectedOs1BalanceDiffManual, _ = expectedOs1BalanceDiffManual.SafeSub(sdk.Coins{utils.ParseCoin("5_991000ucre")})
// 	expectedOs1BalanceDiffManual = expectedOs1BalanceDiffManual.Add(utils.ParseCoin("6_001800uusd"))

// 	ev_os2 := s.getEventOrderSourceOrdersFilled("mockOrderSource2") // taker
// 	s.AssertEqual(sdk.NewInt(3_000000), ev_os2.ExecutedQuantity)
// 	s.AssertEqual(utils.ParseCoin("3_009000uusd"), ev_os2.Paid)
// 	s.AssertEqual(utils.ParseCoin("3_000000ucre"), ev_os2.Received)
// 	expectedOs2BalanceDiffManual, _ = expectedOs2BalanceDiffManual.SafeSub(sdk.Coins{utils.ParseCoin("3_009000uusd")})
// 	expectedOs2BalanceDiffManual = expectedOs2BalanceDiffManual.Add(utils.ParseCoin("3_000000ucre"))

// 	order2, found := s.keeper.GetOrder(s.Ctx, order2.Id)
// 	s.Require().True(found)
// 	s.AssertEqual(sdk.NewInt(10_000000), order2.OpenQuantity)
// 	s.AssertEqual(sdk.NewInt(10_000000), order2.RemainingDeposit)

// 	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
// 	s.Require().True(found)
// 	s.AssertEqual(sdk.NewInt(7_000000), order3.OpenQuantity)
// 	s.AssertEqual(sdk.NewInt(7_039100), order3.RemainingDeposit)

// 	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
// 	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
// 	expectedFeeManual := sdk.Coins{utils.ParseCoin("4500ucre")}
// 	s.AssertEqual(expectedFee, feeAmount)
// 	s.AssertEqual(expectedFeeManual, feeAmount)

// 	os1BalanceAfterMatching := s.GetAllBalances(os1.Address)
// 	os1BalanceDiff, _ := os1BalanceAfterMatching.SafeSub(os1BalanceBeforeMatching)
// 	expectedOs1BalancDiff := sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-5995500)},
// 		utils.ParseCoin("6_001800uusd")}
// 	s.AssertEqual(expectedOs1BalancDiff, os1BalanceDiff)
// 	s.AssertEqual(expectedOs1BalanceDiffManual, os1BalanceDiff)

// 	os2BalanceAfterMatching := s.GetAllBalances(os2.Address)
// 	os2BalanceDiff, _ := os2BalanceAfterMatching.SafeSub(os2BalanceBeforeMatching)
// 	expectedOs2BalancDiff := sdk.Coins{sdk.Coin{Denom: "uusd", Amount: sdk.NewInt(-3009000)},
// 		utils.ParseCoin("3_000000ucre")}
// 	s.AssertEqual(expectedOs2BalancDiff, os2BalanceDiff)
// 	s.AssertEqual(expectedOs2BalanceDiffManual, os2BalanceDiff)
// }
