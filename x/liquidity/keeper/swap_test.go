package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestLimitOrder() {
	// Create a denom1/denom2 pair and set last price to 1.0
	pair1 := s.createPair(s.addr(0), "denom1", "denom2", true)
	lastPrice := utils.ParseDec("1.0")
	pair1.LastPrice = &lastPrice
	s.keeper.SetPair(s.ctx, pair1)

	// denom2/denom1 pair doesn't have last price
	pair2 := s.createPair(s.addr(0), "denom2", "denom1", true)

	orderer := s.addr(1)
	s.fundAddr(orderer, utils.ParseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLimitOrder
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"), "denom1",
				utils.ParseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"wrong offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom1"), "denom2",
				utils.ParseDec("1.0"), newInt(1000000), 0),
			"denom pair (denom2, denom1) != (denom1, denom2): wrong denom pair",
		},
		{
			"correct offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom1"), "denom2",
				utils.ParseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"price not fit in ticks",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionSell, utils.ParseCoin("1000000denom1"), "denom2",
				utils.ParseDec("1.0005"), newInt(1000000), 0),
			"",
		},
		{
			"too long order lifespan",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionSell, utils.ParseCoin("1000000denom1"), "denom2",
				utils.ParseDec("1.0"), newInt(1000000), 48*time.Hour),
			"48h0m0s is longer than 24h0m0s: order lifespan is too long",
		},
		{
			"pair not found",
			types.NewMsgLimitOrder(
				orderer, 3, types.OrderDirectionBuy, utils.ParseCoin("1000000denom1"), "denom2",
				utils.ParseDec("1.0"), newInt(1000000), 0),
			"pair 3 not found: not found",
		},
		{
			"price out of lower limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"), "denom1",
				utils.ParseDec("0.8"), newInt(1000000), 0),
			"0.800000000000000000 is lower than 0.900000000000000000: price out of range limit",
		},
		{
			"price out of upper limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, utils.ParseCoin("2000000denom2"), "denom1",
				utils.ParseDec("1.2"), newInt(1000000), 0),
			"1.200000000000000000 is higher than 1.100000000000000000: price out of range limit",
		},
		{
			"no price limit without last price",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.OrderDirectionSell, utils.ParseCoin("1000000denom2"), "denom1",
				utils.ParseDec("100.0"), newInt(1000000), 0),
			"",
		},
	} {
		s.Run(tc.name, func() {
			// The msg is valid, but may cause an error when it's being handled in the msg server.
			s.Require().NoError(tc.msg.ValidateBasic())
			req, err := s.keeper.LimitOrder(s.ctx, tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				switch tc.msg.Direction {
				case types.OrderDirectionBuy:
					s.Require().True(req.Price.LTE(tc.msg.Price))
				case types.OrderDirectionSell:
					s.Require().True(req.Price.GTE(tc.msg.Price))
				}
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestLimitOrderRefund() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	orderer := s.addr(1)
	s.fundAddr(orderer, utils.ParseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		msg          *types.MsgLimitOrder
		refundedCoin sdk.Coin
	}{
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"), "denom1",
				utils.ParseDec("1.0"), newInt(1000000), 0),
			utils.ParseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"), "denom1",
				utils.ParseDec("1.0"), newInt(10000), 0),
			utils.ParseCoin("990000denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1000denom2"), "denom1",
				utils.ParseDec("0.9999"), newInt(1000), 0),
			utils.ParseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("102denom2"), "denom1",
				utils.ParseDec("1.001"), newInt(100), 0),
			utils.ParseCoin("1denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionSell, utils.ParseCoin("1000denom1"), "denom2",
				utils.ParseDec("1.100"), newInt(1000), 0),
			utils.ParseCoin("0denom1"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionSell, utils.ParseCoin("1000denom1"), "denom2",
				utils.ParseDec("1.100"), newInt(100), 0),
			utils.ParseCoin("900denom1"),
		},
	} {
		s.Run("", func() {
			s.Require().NoError(tc.msg.ValidateBasic())

			balanceBefore := s.getBalance(orderer, tc.msg.OfferCoin.Denom)
			_, err := s.keeper.LimitOrder(s.ctx, tc.msg)
			s.Require().NoError(err)

			balanceAfter := s.getBalance(orderer, tc.msg.OfferCoin.Denom)

			refundedCoin := balanceAfter.Sub(balanceBefore.Sub(tc.msg.OfferCoin))
			s.Require().True(coinEq(tc.refundedCoin, refundedCoin))
		})
	}
}

func (s *KeeperTestSuite) TestMarketOrder() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// When there is no last price in the pair, only limit orders can be made.
	// These two orders will be matched.
	s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), 0, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), 0, true)
	s.nextBlock()

	// Now users can make market orders.
	// In this case, addr(3) user's order takes higher priority than addr(4) user's,
	// because market buy orders have 10% higher price than the last price(1.0).
	s.buyMarketOrder(s.addr(3), pair.Id, sdk.NewInt(10000), 0, true)
	s.buyLimitOrder(s.addr(4), pair.Id, utils.ParseDec("1.08"), sdk.NewInt(10000), 0, true)
	s.sellLimitOrder(s.addr(5), pair.Id, utils.ParseDec("1.07"), sdk.NewInt(10000), 0, true)
	s.nextBlock()

	// Check the result.
	s.Require().True(coinEq(utils.ParseCoin("10000denom1"), s.getBalance(s.addr(3), "denom1")))
	s.Require().True(coinsEq(utils.ParseCoins("10800denom2"), s.getBalances(s.addr(4))))
}

func (s *KeeperTestSuite) TestMarketOrderRefund() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	p := utils.ParseDec("1.0")
	pair.LastPrice = &p
	s.keeper.SetPair(s.ctx, pair)
	orderer := s.addr(1)
	s.fundAddr(orderer, utils.ParseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		msg          *types.MsgMarketOrder
		refundedCoin sdk.Coin
	}{
		{
			types.NewMsgMarketOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1100000denom2"), "denom1",
				newInt(1000000), 0),
			utils.ParseCoin("0denom2"),
		},
		{
			types.NewMsgMarketOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"), "denom1",
				newInt(10000), 0),
			utils.ParseCoin("989000denom2"),
		},
		{
			types.NewMsgMarketOrder(
				orderer, pair.Id, types.OrderDirectionSell, utils.ParseCoin("1000000denom1"), "denom2",
				newInt(10000), 0),
			utils.ParseCoin("990000denom1"),
		},
	} {
		s.Run("", func() {
			s.Require().NoError(tc.msg.ValidateBasic())

			balanceBefore := s.getBalance(orderer, tc.msg.OfferCoin.Denom)
			_, err := s.keeper.MarketOrder(s.ctx, tc.msg)
			s.Require().NoError(err)

			balanceAfter := s.getBalance(orderer, tc.msg.OfferCoin.Denom)

			refundedCoin := balanceAfter.Sub(balanceBefore.Sub(tc.msg.OfferCoin))
			s.Require().True(coinEq(tc.refundedCoin, refundedCoin))
		})
	}
}

func (s *KeeperTestSuite) TestMarketOrderWithNoLastPrice() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.Require().Nil(pair.LastPrice)
	offerCoin := utils.ParseCoin("10000denom2")
	s.fundAddr(s.addr(1), sdk.NewCoins(offerCoin))
	msg := types.NewMsgMarketOrder(
		s.addr(1), pair.Id, types.OrderDirectionBuy, offerCoin, "denom1", sdk.NewInt(10000), 0)
	_, err := s.keeper.MarketOrder(s.ctx, msg)
	s.Require().ErrorIs(err, types.ErrNoLastPrice)
}

func (s *KeeperTestSuite) TestSingleOrderNoMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	// Execute matching
	liquidity.EndBlocker(ctx, k)

	order, found := k.GetOrder(ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusNotMatched, order.Status)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Second))
	// Expire the order, here BeginBlocker is not called to check
	// the request's changed status
	liquidity.EndBlocker(ctx, k)

	order, _ = k.GetOrder(ctx, order.PairId, order.Id)
	s.Require().Equal(types.OrderStatusExpired, order.Status)

	s.Require().True(coinsEq(utils.ParseCoins("1000000denom2"), s.getBalances(s.addr(1))))
}

func (s *KeeperTestSuite) TestTwoOrderExactMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req1 := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), newInt(10000), time.Hour, true)
	req2 := s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.0"), newInt(10000), time.Hour, true)
	liquidity.EndBlocker(ctx, k)

	req1, _ = k.GetOrder(ctx, req1.PairId, req1.Id)
	s.Require().Equal(types.OrderStatusCompleted, req1.Status)
	req2, _ = k.GetOrder(ctx, req2.PairId, req2.Id)
	s.Require().Equal(types.OrderStatusCompleted, req2.Status)

	s.Require().True(coinsEq(utils.ParseCoins("10000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(utils.ParseCoins("10000denom2"), s.getBalances(s.addr(2))))

	pair, _ = k.GetPair(ctx, pair.Id)
	s.Require().NotNil(pair.LastPrice)
	s.Require().True(decEq(utils.ParseDec("1.0"), *pair.LastPrice))
}

func (s *KeeperTestSuite) TestPartialMatch() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Hour, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(5000), 0, true)
	s.nextBlock()

	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusPartiallyMatched, order.Status)
	s.Require().True(coinEq(utils.ParseCoin("5000denom2"), order.RemainingOfferCoin))
	s.Require().True(coinEq(utils.ParseCoin("5000denom1"), order.ReceivedCoin))
	s.Require().True(intEq(sdk.NewInt(5000), order.OpenAmount))

	s.sellMarketOrder(s.addr(3), pair.Id, sdk.NewInt(5000), 0, true)
	s.nextBlock()

	// Now completely matched.
	_, found = s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestMatchWithLowPricePool() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	// Create a pool with very low price.
	s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000000000000000000000000000000000000000denom1,1000000denom2"), true)
	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.000000000000001000"), sdk.NewInt(100000000000000000), 10*time.Second, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusNotMatched, order.Status)
}

func (s *KeeperTestSuite) TestCancelOrder() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), newInt(10000), types.DefaultMaxOrderLifespan, true)

	// Cannot cancel an order within a same batch
	err := k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), order.PairId, order.Id))
	s.Require().ErrorIs(err, types.ErrSameBatch)

	s.nextBlock()

	// Now an order can be canceled
	err = k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), order.PairId, order.Id))
	s.Require().NoError(err)

	order, found := k.GetOrder(ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusCanceled, order.Status)

	// Coins are refunded
	s.Require().True(coinsEq(utils.ParseCoins("10000denom2"), s.getBalances(s.addr(1))))

	s.nextBlock()

	// Order is deleted
	_, found = k.GetOrder(ctx, order.PairId, order.Id)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestCancelAllOrders() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Hour, true)
	s.cancelAllOrders(s.addr(1), nil) // CancelAllOrders doesn't cancel orders within in same batch
	s.nextBlock()

	// The order is still alive.
	_, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)

	s.cancelAllOrders(s.addr(1), nil) // This time, it cancels the order.
	order, found = s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	// Canceling an order doesn't delete the order immediately.
	s.Require().True(found)
	// Instead, the order becomes canceled.
	s.Require().Equal(types.OrderStatusCanceled, order.Status)

	// The order won't be matched with this market order, since the order is
	// already canceled.
	s.sellLimitOrder(s.addr(3), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), 0, true)
	s.nextBlock()
	s.Require().True(coinsEq(utils.ParseCoins("10000denom2"), s.getBalances(s.addr(1))))

	pair2 := s.createPair(s.addr(0), "denom2", "denom3", true)
	s.buyLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Hour, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.5"), sdk.NewInt(10000), time.Hour, true)
	s.sellLimitOrder(s.addr(2), pair2.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Hour, true)
	s.nextBlock()
	// CancelAllOrders can cancel orders in specific pairs.
	s.cancelAllOrders(s.addr(2), []uint64{pair.Id})
	// Coins from first two orders are refunded, but not from the last order.
	s.Require().True(coinsEq(utils.ParseCoins("10000denom2,10000denom1"), s.getBalances(s.addr(2))))
}

func (s *KeeperTestSuite) TestDustCollector() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.9005"), newInt(1000), 0, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.9005"), newInt(1000), 0, true)
	s.nextBlock()

	s.Require().True(coinsEq(utils.ParseCoins("1000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(utils.ParseCoins("900denom2"), s.getBalances(s.addr(2))))

	s.Require().True(s.getBalances(pair.GetEscrowAddress()).IsZero())
	params := s.keeper.GetParams(s.ctx)
	dustCollectorAddr, _ := sdk.AccAddressFromBech32(params.DustCollectorAddress)
	s.Require().True(coinsEq(utils.ParseCoins("1denom2"), s.getBalances(dustCollectorAddr)))
}

func (s *KeeperTestSuite) TestFitPrice() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	lastPrice := utils.ParseDec("1")
	pair.LastPrice = &lastPrice
	s.keeper.SetPair(s.ctx, pair)

	for _, tc := range []struct {
		name        string
		price       sdk.Dec
		dir         types.OrderDirection
		expectedErr string
	}{
		{
			"",
			utils.ParseDec("1"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			utils.ParseDec("1"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			utils.ParseDec("1.1"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			utils.ParseDec("0.9"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			utils.ParseDec("1.099999999"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			utils.ParseDec("0.900000001"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			utils.ParseDec("1.10000001"),
			types.OrderDirectionBuy,
			"1.100000010000000000 is higher than 1.100000000000000000: price out of range limit",
		},
		{
			"",
			utils.ParseDec("0.8999999"),
			types.OrderDirectionSell,
			"0.899999900000000000 is lower than 0.900000000000000000: price out of range limit",
		},
	} {
		s.Run(tc.name, func() {
			amt := newInt(10000)
			var offerCoin sdk.Coin
			var demandCoinDenom string
			switch tc.dir {
			case types.OrderDirectionBuy:
				offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, tc.price.MulInt(amt).Ceil().TruncateInt())
				demandCoinDenom = pair.BaseCoinDenom
			case types.OrderDirectionSell:
				offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
				demandCoinDenom = pair.QuoteCoinDenom
			}
			s.fundAddr(s.addr(1), sdk.NewCoins(offerCoin))
			msg := types.NewMsgLimitOrder(s.addr(1), pair.Id, tc.dir, offerCoin, demandCoinDenom, tc.price, amt, 0)
			req, err := s.keeper.LimitOrder(s.ctx, msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				switch tc.dir {
				case types.OrderDirectionBuy:
					s.Require().True(req.Price.LTE(tc.price))
				case types.OrderDirectionSell:
					s.Require().True(req.Price.GTE(tc.price))
				}
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetOrdersByOrderer() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pair2 := s.createPair(s.addr(0), "denom2", "denom3", true)

	order1 := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), 0, true)
	order2 := s.sellLimitOrder(s.addr(1), pair2.Id, utils.ParseDec("1.0"), sdk.NewInt(10000), 0, true)

	orders := s.keeper.GetOrdersByOrderer(s.ctx, s.addr(1))
	s.Require().Len(orders, 2)
	s.Require().Equal(order1.PairId, orders[0].PairId)
	s.Require().Equal(order1.Id, orders[0].Id)
	s.Require().Equal(order2.PairId, orders[1].PairId)
	s.Require().Equal(order2.Id, orders[1].Id)
}

func (s *KeeperTestSuite) TestInsufficientRemainingOfferCoin() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.5"), sdk.NewInt(10000), time.Minute, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.5"), sdk.NewInt(1001), 0, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	liquidity.BeginBlocker(s.ctx, s.keeper)

	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.5"), sdk.NewInt(8999), 0, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusExpired, order.Status)
	s.Require().True(intEq(sdk.OneInt(), order.OpenAmount))
}

func (s *KeeperTestSuite) TestNegativeOpenAmount() {
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(utils.ParseTime("2022-03-01T00:00:00Z"))

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.82"), sdk.NewInt(648744), 0, true)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.82"), sdk.NewInt(648745), 0, true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().False(order.OpenAmount.IsNegative())

	genState := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotPanics(func() {
		s.keeper.InitGenesis(s.ctx, *genState)
	})
}

func (s *KeeperTestSuite) TestRejectSmallOrders() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.fundAddr(s.addr(1), utils.ParseCoins("10000000denom1,10000000denom2"))

	// Too small offer coin amount.
	msg := types.NewMsgLimitOrder(
		s.addr(1), pair.Id, types.OrderDirectionBuy, utils.ParseCoin("99denom2"),
		"denom1", utils.ParseDec("0.1"), sdk.NewInt(990), 0)
	s.Require().EqualError(msg.ValidateBasic(), "offer coin is less than minimum coin amount: invalid request")

	// Too small order amount.
	msg = types.NewMsgLimitOrder(
		s.addr(1), pair.Id, types.OrderDirectionBuy, utils.ParseCoin("990denom2"),
		"denom1", utils.ParseDec("10.0"), sdk.NewInt(99), 0)
	s.Require().EqualError(msg.ValidateBasic(), "base coin is less than minimum coin amount: invalid request")

	// Too small orders.
	msg = types.NewMsgLimitOrder(
		s.addr(1), pair.Id, types.OrderDirectionBuy, utils.ParseCoin("101denom2"),
		"denom1", utils.ParseDec("0.00010001"), sdk.NewInt(999999), 0)
	s.Require().NoError(msg.ValidateBasic())
	_, err := s.keeper.LimitOrder(s.ctx, msg)
	s.Require().ErrorIs(err, types.ErrTooSmallOrder)

	msg = types.NewMsgLimitOrder(
		s.addr(1), pair.Id, types.OrderDirectionSell, utils.ParseCoin("999999denom1"),
		"denom2", utils.ParseDec("0.0001"), sdk.NewInt(999999), 0)
	s.Require().NoError(msg.ValidateBasic())
	_, err = s.keeper.LimitOrder(s.ctx, msg)
	s.Require().ErrorIs(err, types.ErrTooSmallOrder)

	// Too small offer coin amount.
	msg2 := types.NewMsgMarketOrder(
		s.addr(1), pair.Id, types.OrderDirectionSell, utils.ParseCoin("99denom1"),
		"denom2", sdk.NewInt(99), 0)
	s.Require().EqualError(msg2.ValidateBasic(), "offer coin is less than minimum coin amount: invalid request")

	// Too small order amount.
	msg2 = types.NewMsgMarketOrder(
		s.addr(1), pair.Id, types.OrderDirectionSell, utils.ParseCoin("100denom1"),
		"denom2", sdk.NewInt(99), 0)
	s.Require().EqualError(msg2.ValidateBasic(), "base coin is less than minimum coin amount: invalid request")

	p := utils.ParseDec("0.0001")
	pair.LastPrice = &p
	s.keeper.SetPair(s.ctx, pair)

	// Too small orders.
	msg2 = types.NewMsgMarketOrder(
		s.addr(1), pair.Id, types.OrderDirectionBuy, utils.ParseCoin("100denom2"),
		"denom1", sdk.NewInt(909090), 0)
	s.Require().NoError(msg2.ValidateBasic())
	_, err = s.keeper.MarketOrder(s.ctx, msg2)
	s.Require().ErrorIs(err, types.ErrTooSmallOrder)

	msg2 = types.NewMsgMarketOrder(
		s.addr(1), pair.Id, types.OrderDirectionSell, utils.ParseCoin("1000denom1"),
		"denom2", sdk.NewInt(1000), 0)
	s.Require().NoError(msg2.ValidateBasic())
	_, err = s.keeper.MarketOrder(s.ctx, msg2)
	s.Require().ErrorIs(err, types.ErrTooSmallOrder)
}

func (s *KeeperTestSuite) TestExpireSmallOrders() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.000018"), sdk.NewInt(10000000), time.Minute, true)
	// This order should have 10000 open amount after matching.
	// If this order would be matched after that, then the orderer will receive
	// floor(10000*0.000018) demand coin, which is zero.
	// So the order must have been expired after matching.
	order := s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.000018"), sdk.NewInt(10010000), time.Minute, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	order, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusExpired, order.Status)
	liquidity.BeginBlocker(s.ctx, s.keeper) // Delete outdated states.

	s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("0.000019"), sdk.NewInt(100000000), time.Minute, true)
	s.sellLimitOrder(s.addr(3), pair.Id, utils.ParseDec("0.000019"), sdk.NewInt(100000000), time.Minute, true)
	liquidity.EndBlocker(s.ctx, s.keeper)
}
