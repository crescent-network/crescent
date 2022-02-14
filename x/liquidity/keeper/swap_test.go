package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func (s *KeeperTestSuite) TestLimitOrder() {
	// Create a denom1/denom2 pair and set last price to 1.0
	pair1 := s.createPair(s.addr(0), "denom1", "denom2", true)
	lastPrice := squad.ParseDec("1.0")
	pair1.LastPrice = &lastPrice
	s.keeper.SetPair(s.ctx, pair1)

	// denom2/denom1 pair doesn't have last price
	pair2 := s.createPair(s.addr(0), "denom2", "denom1", true)

	orderer := s.addr(1)
	s.fundAddr(orderer, squad.ParseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLimitOrder
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom2"), "denom1",
				squad.ParseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"wrong offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom1"), "denom2",
				squad.ParseDec("1.0"), newInt(1000000), 0),
			"denom pair (denom2, denom1) != (denom1, denom2): wrong denom pair",
		},
		{
			"correct offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom1"), "denom2",
				squad.ParseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"price not fit in ticks",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionSell, squad.ParseCoin("1000000denom1"), "denom2",
				squad.ParseDec("1.0005"), newInt(1000000), 0),
			"",
		},
		{
			"too long order lifespan",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionSell, squad.ParseCoin("1000000denom1"), "denom2",
				squad.ParseDec("1.0"), newInt(1000000), 48*time.Hour),
			"48h0m0s is longer than 24h0m0s: order lifespan is too long",
		},
		{
			"pair not found",
			types.NewMsgLimitOrder(
				orderer, 3, types.OrderDirectionBuy, squad.ParseCoin("1000000denom1"), "denom2",
				squad.ParseDec("1.0"), newInt(1000000), 0),
			"pair 3 not found: not found",
		},
		{
			"price out of lower limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom2"), "denom1",
				squad.ParseDec("0.8"), newInt(1000000), 0),
			"0.800000000000000000 is lower than 0.900000000000000000: price out of range limit",
		},
		{
			"price out of upper limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.OrderDirectionBuy, squad.ParseCoin("2000000denom2"), "denom1",
				squad.ParseDec("1.2"), newInt(1000000), 0),
			"1.200000000000000000 is higher than 1.100000000000000000: price out of range limit",
		},
		{
			"no price limit without last price",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.OrderDirectionSell, squad.ParseCoin("1000000denom2"), "denom1",
				squad.ParseDec("100.0"), newInt(1000000), 0),
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
	s.fundAddr(orderer, squad.ParseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		msg          *types.MsgLimitOrder
		refundedCoin sdk.Coin
	}{
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom2"), "denom1",
				squad.ParseDec("1.0"), newInt(1000000), 0),
			squad.ParseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, squad.ParseCoin("1000000denom2"), "denom1",
				squad.ParseDec("1.0"), newInt(10000), 0),
			squad.ParseCoin("990000denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, squad.ParseCoin("1000denom2"), "denom1",
				squad.ParseDec("0.9999"), newInt(1000), 0),
			squad.ParseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, squad.ParseCoin("102denom2"), "denom1",
				squad.ParseDec("1.001"), newInt(100), 0),
			squad.ParseCoin("1denom2"),
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

func (s *KeeperTestSuite) TestSingleOrderNoMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req := s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	// Execute matching
	liquidity.EndBlocker(ctx, k)

	req, found := k.GetOrder(ctx, req.PairId, req.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusNotMatched, req.Status)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Second))
	// Expire the order, here BeginBlocker is not called to check
	// the request's changed status
	liquidity.EndBlocker(ctx, k)

	req, _ = k.GetOrder(ctx, req.PairId, req.Id)
	s.Require().Equal(types.OrderStatusExpired, req.Status)

	s.Require().True(coinsEq(squad.ParseCoins("1000000denom2"), s.getBalances(s.addr(1))))
}

func (s *KeeperTestSuite) TestTwoOrderExactMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req1 := s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), newInt(10000), time.Hour, true)
	req2 := s.sellLimitOrder(s.addr(2), pair.Id, squad.ParseDec("1.0"), newInt(10000), time.Hour, true)
	liquidity.EndBlocker(ctx, k)

	req1, _ = k.GetOrder(ctx, req1.PairId, req1.Id)
	s.Require().Equal(types.OrderStatusCompleted, req1.Status)
	req2, _ = k.GetOrder(ctx, req2.PairId, req2.Id)
	s.Require().Equal(types.OrderStatusCompleted, req2.Status)

	s.Require().True(coinsEq(squad.ParseCoins("10000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(squad.ParseCoins("10000denom2"), s.getBalances(s.addr(2))))

	pair, _ = k.GetPair(ctx, pair.Id)
	s.Require().NotNil(pair.LastPrice)
	s.Require().True(decEq(squad.ParseDec("1.0"), *pair.LastPrice))
}

func (s *KeeperTestSuite) TestCancelOrder() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req := s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), newInt(10000), types.DefaultMaxOrderLifespan, true)

	// Cannot cancel an order within a same batch
	err := k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), req.PairId, req.Id))
	s.Require().ErrorIs(err, types.ErrSameBatch)

	s.nextBlock()

	// Now an order can be canceled
	err = k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), req.PairId, req.Id))
	s.Require().NoError(err)

	req, found := k.GetOrder(ctx, req.PairId, req.Id)
	s.Require().True(found)
	s.Require().Equal(types.OrderStatusCanceled, req.Status)

	// Coins are refunded
	s.Require().True(coinsEq(squad.ParseCoins("10000denom2"), s.getBalances(s.addr(1))))

	s.nextBlock()

	// Order is deleted
	_, found = k.GetOrder(ctx, req.PairId, req.Id)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestDustCollector() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("0.9005"), newInt(1000), 0, true)
	s.sellLimitOrder(s.addr(2), pair.Id, squad.ParseDec("0.9005"), newInt(1000), 0, true)
	s.nextBlock()

	s.Require().True(coinsEq(squad.ParseCoins("1000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(squad.ParseCoins("900denom2"), s.getBalances(s.addr(2))))

	s.Require().True(s.getBalances(pair.GetEscrowAddress()).IsZero())
	s.Require().True(coinsEq(squad.ParseCoins("1denom2"), s.getBalances(types.DustCollectorAddress)))
}

func (s *KeeperTestSuite) TestFitPrice() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	lastPrice := squad.ParseDec("1")
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
			squad.ParseDec("1"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			squad.ParseDec("1"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			squad.ParseDec("1.1"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			squad.ParseDec("0.9"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			squad.ParseDec("1.099999999"),
			types.OrderDirectionBuy,
			"",
		},
		{
			"",
			squad.ParseDec("0.900000001"),
			types.OrderDirectionSell,
			"",
		},
		{
			"",
			squad.ParseDec("1.10000001"),
			types.OrderDirectionBuy,
			"1.100000010000000000 is higher than 1.100000000000000000: price out of range limit",
		},
		{
			"",
			squad.ParseDec("0.8999999"),
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
