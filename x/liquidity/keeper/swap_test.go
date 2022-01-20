package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (s *KeeperTestSuite) TestSingleOrderNoMatch() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	req := s.swapBatchBuy(s.addr(1), pair.Id, parseCoin("1000000denom2"), parseDec("1.0"), sdk.NewInt(1), 10 * time.Second, true)
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
