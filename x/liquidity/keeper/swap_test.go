package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	"github.com/cosmosquad-labs/squad/x/liquidity"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func (s *KeeperTestSuite) TestLimitOrder() {
	// Create a denom1/denom2 pair and set last price to 1.0
	pair1 := s.createPair(s.addr(0), "denom1", "denom2", true)
	lastPrice := parseDec("1.0")
	pair1.LastPrice = &lastPrice
	s.keeper.SetPair(s.ctx, pair1)

	// denom2/denom1 pair doesn't have last price
	pair2 := s.createPair(s.addr(0), "denom2", "denom1", true)

	orderer := s.addr(1)
	s.fundAddr(orderer, parseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLimitOrder
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionBuy, parseCoin("1000000denom2"), "denom1",
				parseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"wrong offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionBuy, parseCoin("1000000denom1"), "denom2",
				parseDec("1.0"), newInt(1000000), 0),
			"denom pair (denom2, denom1) != (denom1, denom2): wrong denom pair",
		},
		{
			"correct offer coin and demand coin denom",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.SwapDirectionBuy, parseCoin("1000000denom1"), "denom2",
				parseDec("1.0"), newInt(1000000), 0),
			"",
		},
		{
			"price not fit in ticks",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionSell, parseCoin("1000000denom1"), "denom2",
				parseDec("1.0005"), newInt(1000000), 0),
			"price not fit into ticks",
		},
		{
			"too long order lifespan",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionSell, parseCoin("1000000denom1"), "denom2",
				parseDec("1.0"), newInt(1000000), 48*time.Hour),
			"order lifespan is too long",
		},
		{
			"pair not found",
			types.NewMsgLimitOrder(
				orderer, 3, types.SwapDirectionBuy, parseCoin("1000000denom1"), "denom2",
				parseDec("1.0"), newInt(1000000), 0),
			"pair not found: not found",
		},
		{
			"price out of lower limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionBuy, parseCoin("1000000denom2"), "denom1",
				parseDec("0.8"), newInt(1000000), 0),
			"price out of range limit",
		},
		{
			"price out of upper limit",
			types.NewMsgLimitOrder(
				orderer, pair1.Id, types.SwapDirectionBuy, parseCoin("2000000denom2"), "denom1",
				parseDec("1.2"), newInt(1000000), 0),
			"price out of range limit",
		},
		{
			"no price limit without last price",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.SwapDirectionSell, parseCoin("1000000denom2"), "denom1",
				parseDec("100.0"), newInt(1000000), 0),
			"",
		},
		{

			"insufficient offer coin",
			types.NewMsgLimitOrder(
				orderer, pair2.Id, types.SwapDirectionBuy, parseCoin("1000000denom1"), "denom2",
				parseDec("10.0"), newInt(1000000), 0),
			"insufficient offer coin",
		},
	} {
		s.Run(tc.name, func() {
			// The msg is valid, but may cause an error when it's being handled in the msg server.
			s.Require().NoError(tc.msg.ValidateBasic())
			_, err := s.keeper.LimitOrder(s.ctx, tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestLimitOrderRefund() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	orderer := s.addr(1)
	s.fundAddr(orderer, parseCoins("1000000000denom1,1000000000denom2"))

	for _, tc := range []struct {
		msg          *types.MsgLimitOrder
		refundedCoin sdk.Coin
	}{
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.SwapDirectionBuy, parseCoin("1000000denom2"), "denom1",
				parseDec("1.0"), newInt(1000000), 0),
			parseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.SwapDirectionBuy, parseCoin("1000000denom2"), "denom1",
				parseDec("1.0"), newInt(10000), 0),
			parseCoin("990000denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.SwapDirectionBuy, parseCoin("1000denom2"), "denom1",
				parseDec("0.9999"), newInt(1000), 0),
			parseCoin("0denom2"),
		},
		{
			types.NewMsgLimitOrder(
				orderer, pair.Id, types.SwapDirectionBuy, parseCoin("102denom2"), "denom1",
				parseDec("1.001"), newInt(100), 0),
			parseCoin("1denom2"),
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

	req := s.buyLimitOrder(s.addr(1), pair.Id, parseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	// Execute matching
	liquidity.EndBlocker(ctx, k)

	req, found := k.GetSwapRequest(ctx, req.PairId, req.Id)
	s.Require().True(found)
	s.Require().Equal(types.SwapRequestStatusNotMatched, req.Status)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(10 * time.Second))
	// Expire the swap request, here BeginBlocker is not called to check
	// the request's changed status
	liquidity.EndBlocker(ctx, k)

	req, _ = k.GetSwapRequest(ctx, req.PairId, req.Id)
	s.Require().Equal(types.SwapRequestStatusExpired, req.Status)

	s.Require().True(coinsEq(parseCoins("1000000denom2"), s.getBalances(s.addr(1))))
}

func (s *KeeperTestSuite) TestTwoOrderExactMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req1 := s.buyLimitOrder(s.addr(1), pair.Id, parseDec("1.0"), newInt(10000), time.Hour, true)
	req2 := s.sellLimitOrder(s.addr(2), pair.Id, parseDec("1.0"), newInt(10000), time.Hour, true)
	liquidity.EndBlocker(ctx, k)

	req1, _ = k.GetSwapRequest(ctx, req1.PairId, req1.Id)
	s.Require().Equal(types.SwapRequestStatusCompleted, req1.Status)
	req2, _ = k.GetSwapRequest(ctx, req2.PairId, req2.Id)
	s.Require().Equal(types.SwapRequestStatusCompleted, req2.Status)

	s.Require().True(coinsEq(parseCoins("10000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(parseCoins("10000denom2"), s.getBalances(s.addr(2))))

	pair, _ = k.GetPair(ctx, pair.Id)
	s.Require().NotNil(pair.LastPrice)
	s.Require().True(decEq(parseDec("1.0"), *pair.LastPrice))
}

func (s *KeeperTestSuite) TestCancelOrder() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req := s.buyLimitOrder(s.addr(1), pair.Id, parseDec("1.0"), newInt(10000), types.DefaultMaxOrderLifespan, true)

	// Cannot cancel an order within a same batch
	err := k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), req.PairId, req.Id))
	s.Require().ErrorIs(err, types.ErrSameBatch)

	s.nextBlock()

	// Now an order can be canceled
	err = k.CancelOrder(ctx, types.NewMsgCancelOrder(s.addr(1), req.PairId, req.Id))
	s.Require().NoError(err)

	req, found := k.GetSwapRequest(ctx, req.PairId, req.Id)
	s.Require().True(found)
	s.Require().Equal(types.SwapRequestStatusCanceled, req.Status)

	// Coins are refunded
	s.Require().True(coinsEq(parseCoins("10000denom2"), s.getBalances(s.addr(1))))

	s.nextBlock()

	// Swap request is deleted
	_, found = k.GetSwapRequest(ctx, req.PairId, req.Id)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestDustCollector() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	s.buyLimitOrder(s.addr(1), pair.Id, parseDec("0.9005"), newInt(1000), 0, true)
	s.sellLimitOrder(s.addr(2), pair.Id, parseDec("0.9005"), newInt(1000), 0, true)
	s.nextBlock()

	s.Require().True(coinsEq(parseCoins("1000denom1"), s.getBalances(s.addr(1))))
	s.Require().True(coinsEq(parseCoins("900denom2"), s.getBalances(s.addr(2))))

	s.Require().True(s.getBalances(pair.GetEscrowAddress()).IsZero())
	s.Require().True(coinsEq(parseCoins("1denom2"), s.getBalances(types.DustCollectorAddress)))
}
