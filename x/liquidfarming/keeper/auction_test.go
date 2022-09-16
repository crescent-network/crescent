package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestPlaceBid_Validation() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.NewInt(1_000_000), sdk.ZeroDec())
	s.farm(pool.Id, s.addr(0), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	s.createRewardsAuction(pool.Id)

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	_, found := s.keeper.GetRewardsAuction(s.ctx, pool.Id, auctionId)
	s.Require().True(found)

	s.placeBid(pool.Id, s.addr(1), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(2), utils.ParseCoin("15_000_000pool1"), true)
	s.nextBlock()

	bids := s.keeper.GetBidsByPoolId(s.ctx, pool.Id)
	s.Require().Len(bids, 2)

	s.fundAddr(s.addr(5), utils.ParseCoins("100pool1"))

	for _, tc := range []struct {
		name        string
		msg         *types.MsgPlaceBid
		postRun     func(ctx sdk.Context, bid types.Bid)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgPlaceBid(
				pool.Id,
				s.addr(0).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000),
			),
			func(ctx sdk.Context, bid types.Bid) {
				s.Require().Equal(pool.Id, bid.PoolId)
				s.Require().Equal(s.addr(0), bid.GetBidder())
				s.Require().Equal(sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000), bid.Amount)
			},
			"",
		},
		{
			"insufficient funds",
			types.NewMsgPlaceBid(
				pool.Id,
				s.addr(5).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000),
			),
			nil,
			"100pool1 is smaller than 30000000pool1: insufficient funds",
		},
		{
			"invalid bidding coin denom",
			types.NewMsgPlaceBid(
				pool.Id,
				s.addr(5).String(),
				sdk.NewInt64Coin("denom1", 30_000_000),
			),
			nil,
			"expected denom pool1, but got denom1: invalid request",
		},
		{
			"minimum bid amount",
			types.NewMsgPlaceBid(
				pool.Id,
				s.addr(5).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 100),
			),
			nil,
			"100 is smaller than 1000000: smaller than minimum amount",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()

			bid, err := s.keeper.PlaceBid(cacheCtx, tc.msg.PoolId, tc.msg.GetBidder(), tc.msg.BiddingCoin)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				bid, found := s.keeper.GetBid(cacheCtx, bid.PoolId, bid.GetBidder())
				s.Require().True(found)
				tc.postRun(cacheCtx, bid)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPlaceBid() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	_, err := s.keeper.PlaceBid(s.ctx, pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.NewInt(10_000_000), sdk.NewInt(10_000_000), sdk.ZeroDec())
	s.farm(pool.Id, s.addr(0), utils.ParseCoin("10_000_000pool1"), true)

	s.advanceEpochDays() // advance epoch to move from QueuedCoins to StakedCoins

	_, err = s.keeper.PlaceBid(s.ctx, pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().EqualError(err, "auction by pool 1 not found: not found")

	s.advanceEpochDays() // trigger AfterAllocateRewards hook to create rewards auction

	var (
		bidderAddr1 = s.addr(1)
		bidderAddr2 = s.addr(2)
	)
	s.fundAddr(bidderAddr1, utils.ParseCoins("250_000_000pool1"))
	s.fundAddr(bidderAddr2, utils.ParseCoins("100_000_000pool1"))

	// Place a bid successfully
	_, err = s.keeper.PlaceBid(s.ctx, pool.Id, bidderAddr1, sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().NoError(err)

	// Place with higher bidding amount
	_, err = s.keeper.PlaceBid(s.ctx, pool.Id, bidderAddr1, sdk.NewInt64Coin(pool.PoolCoinDenom, 150_000_000))
	s.Require().NoError(err)

	// Ensure the refunded amount
	s.Require().Equal(sdk.NewInt(100_000_000), s.getBalance(bidderAddr1, pool.PoolCoinDenom).Amount)

	// Place a bid with less than the winning bid amount
	_, err = s.keeper.PlaceBid(s.ctx, pool.Id, bidderAddr2, sdk.NewInt64Coin(pool.PoolCoinDenom, 90_000_000))
	s.Require().EqualError(err, "90000000 is smaller than 150000000: smaller than winning bid  amount")
}

func (s *KeeperTestSuite) TestRefundBid() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.createLiquidFarm(pool.Id, sdk.NewInt(10_000_000), sdk.NewInt(10_000_000), sdk.ZeroDec())
	s.createRewardsAuction(pool.Id)
	s.placeBid(pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 500_000_000), true)
	s.placeBid(pool.Id, s.addr(1), sdk.NewInt64Coin(pool.PoolCoinDenom, 600_000_000), true)

	for _, tc := range []struct {
		name        string
		msg         *types.MsgRefundBid
		postRun     func(ctx sdk.Context, bid types.Bid)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgRefundBid(
				pool.Id,
				s.addr(0).String(),
			),
			func(ctx sdk.Context, bid types.Bid) {
				s.Require().Equal(pool.Id, bid.PoolId)
				s.Require().Equal(s.addr(0), bid.GetBidder())
			},
			"",
		},
		{
			"auction not found",
			types.NewMsgRefundBid(
				5,
				s.addr(1).String(),
			),
			nil,
			"auction by pool 5 not found: not found",
		},
		{
			"refund winning bid",
			types.NewMsgRefundBid(
				pool.Id,
				s.addr(1).String(),
			),
			nil,
			"winning bid can't be refunded: invalid request",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()
			err := s.keeper.RefundBid(cacheCtx, tc.msg.PoolId, tc.msg.GetBidder())
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				_, found := s.keeper.GetBid(cacheCtx, tc.msg.PoolId, s.addr(0))
				s.Require().False(found)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAfterAllocateRewards() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.createPrivateFixedAmountPlan(s.addr(0), map[string]string{pool.PoolCoinDenom: "1"}, map[string]int64{"denom3": 1_000_000}, true)
	s.farm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(50_000_000)), true)

	s.advanceEpochDays() // advance epoch to move from QueuedCoins to StakedCoins
	s.advanceEpochDays() // trigger AllocateRewards hook to create the first RewardsAuction

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, pool.Id, auctionId)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.placeBid(auction.PoolId, s.addr(1), sdk.NewInt64Coin(pool.PoolCoinDenom, 10_000_000), true)
	s.placeBid(auction.PoolId, s.addr(2), sdk.NewInt64Coin(pool.PoolCoinDenom, 20_000_000), true)
	s.placeBid(auction.PoolId, s.addr(3), sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000), true)
	s.advanceEpochDays() // farming rewards are expected to be accumulated

	// Ensure the state of previous auction is updated
	auction, found = s.keeper.GetRewardsAuction(s.ctx, auction.PoolId, auction.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusFinished, auction.Status)

	// Ensure refunded pool coin amount
	s.Require().Equal(sdk.NewInt(10_000_000), s.getBalance(s.addr(1), pool.PoolCoinDenom).Amount)
	s.Require().Equal(sdk.NewInt(20_000_000), s.getBalance(s.addr(2), pool.PoolCoinDenom).Amount)
	s.Require().Equal(sdk.ZeroInt(), s.getBalance(s.addr(3), pool.PoolCoinDenom).Amount)

	// Ensure winner's balance to see if they received farming rewards
	s.Require().Equal(sdk.NewInt(1_000_000), s.getBalance(s.addr(3), "denom3").Amount)

	// Ensure newly staked amount by the liquid farm reserve account
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	queuedCoins := s.app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().Equal(sdk.NewInt(30_000_000), queuedCoins)
}

func (s *KeeperTestSuite) TestAfterAllocateRewards_NoBid() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.createPrivateFixedAmountPlan(s.addr(0), map[string]string{pool.PoolCoinDenom: "1"}, map[string]int64{"denom3": 1_000_000}, true)
	s.farm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(50_000_000)), true)

	s.advanceEpochDays() // advance epoch to move from QueuedCoins to StakedCoins
	s.advanceEpochDays() // trigger AllocateRewards hook to create the first RewardsAuction

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, pool.Id, auctionId)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.advanceEpochDays() // finish the ongoing rewards auction

	auction, found = s.keeper.GetRewardsAuction(s.ctx, auction.PoolId, auction.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusSkipped, auction.Status)
}
