package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	"github.com/cosmosquad-labs/squad/x/liquidity"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func (s *KeeperTestSuite) TestSingleOrderNoMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req := s.swapBatchBuy(s.addr(1), pair.Id, parseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
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

	req1 := s.swapBatchBuy(s.addr(1), pair.Id, parseDec("1.0"), newInt(10000), time.Hour, true)
	req2 := s.swapBatchSell(s.addr(2), pair.Id, parseDec("1.0"), newInt(10000), time.Hour, true)
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
