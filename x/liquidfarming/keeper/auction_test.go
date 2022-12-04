package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	lpfarmtypes "github.com/crescent-network/crescent/v3/x/lpfarm/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestPlaceBid_Validation() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.NewInt(1_000_000), sdk.ZeroDec())

	s.liquidFarm(pool.Id, s.addr(0), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	s.createRewardsAuction(pool.Id)

	auction, found := s.keeper.GetLastRewardsAuction(s.ctx, pool.Id)
	s.Require().True(found)

	s.placeBid(pool.Id, s.addr(1), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(2), utils.ParseCoin("15_000_000pool1"), true)
	s.nextBlock()

	s.Require().Len(s.keeper.GetBidsByPoolId(s.ctx, pool.Id), 2)

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
				auction.Id,
				auction.PoolId,
				helperAddr.String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000),
			),
			func(ctx sdk.Context, bid types.Bid) {
				s.Require().Equal(pool.Id, bid.PoolId)
				s.Require().Equal(helperAddr, bid.GetBidder())
				s.Require().Equal(sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000), bid.Amount)
			},
			"",
		},
		{
			"insufficient funds",
			types.NewMsgPlaceBid(
				auction.Id,
				auction.PoolId,
				s.addr(5).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000),
			),
			nil,
			"100pool1 is smaller than 30000000pool1: insufficient funds",
		},
		{
			"minimum bid amount",
			types.NewMsgPlaceBid(
				auction.Id,
				auction.PoolId,
				s.addr(5).String(),
				sdk.NewInt64Coin(pool.PoolCoinDenom, 100),
			),
			nil,
			"must be greater than the minimum bid amount 1000000: invalid request",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()

			bid, err := s.keeper.PlaceBid(cacheCtx, tc.msg.AuctionId, tc.msg.PoolId, tc.msg.GetBidder(), tc.msg.BiddingCoin)
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
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	_, err := s.keeper.PlaceBid(s.ctx, 1, pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.NewInt(10_000_000), sdk.NewInt(10_000_000), sdk.ZeroDec())

	s.liquidFarm(pool.Id, s.addr(0), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	_, err = s.keeper.PlaceBid(s.ctx, 1, pool.Id, s.addr(0), sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().EqualError(err, "auction by pool 1 not found: not found")

	// Create the first auction
	s.nextAuction()

	var (
		bidderAddr1 = s.addr(1)
		bidderAddr2 = s.addr(2)
	)

	s.fundAddr(bidderAddr1, utils.ParseCoins("250_000_000pool1"))
	s.fundAddr(bidderAddr2, utils.ParseCoins("100_000_000pool1"))

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)

	// Place a bid successfully
	_, err = s.keeper.PlaceBid(s.ctx, auctionId, pool.Id, bidderAddr1, sdk.NewInt64Coin(pool.PoolCoinDenom, 100_000_000))
	s.Require().NoError(err)

	// Place with higher bidding amount
	_, err = s.keeper.PlaceBid(s.ctx, auctionId, pool.Id, bidderAddr1, sdk.NewInt64Coin(pool.PoolCoinDenom, 150_000_000))
	s.Require().NoError(err)

	// Ensure the refunded amount
	s.Require().Equal(sdk.NewInt(100_000_000), s.getBalance(bidderAddr1, pool.PoolCoinDenom).Amount)

	// Place a bid with less than the winning bid amount
	_, err = s.keeper.PlaceBid(s.ctx, auctionId, pool.Id, bidderAddr2, sdk.NewInt64Coin(pool.PoolCoinDenom, 90_000_000))
	s.Require().EqualError(err, "must be greater than the winning bid amount 150000000: invalid request")
}

func (s *KeeperTestSuite) TestPlaceBid_AuctionStatus() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	lf1 := s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.liquidFarm(lf1.PoolId, helperAddr, utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	// Create the first auction
	s.nextAuction()

	s.fundAddr(s.addr(0), utils.ParseCoins("300_000_000pool1"))

	_, err := s.keeper.PlaceBid(s.ctx, 1, pool.Id, s.addr(0), utils.ParseCoin("100_000_000pool1"))
	s.Require().NoError(err)

	// Finish the first auction and create next one
	s.nextAuction()

	// Place a bid for the old auction
	_, err = s.keeper.PlaceBid(s.ctx, 1, pool.Id, s.addr(0), utils.ParseCoin("200_000_000pool1"))
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, lf1.PoolId)
	_, err = s.keeper.PlaceBid(s.ctx, auctionId, pool.Id, s.addr(0), utils.ParseCoin("200_000_000pool1"))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestRefundBid() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.NewInt(10_000_000), sdk.NewInt(10_000_000), sdk.ZeroDec())
	s.createRewardsAuction(pool.Id)

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)

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
				auctionId,
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
				auctionId,
				5,
				s.addr(1).String(),
			),
			nil,
			"auction by pool 5 not found: not found",
		},
		{
			"refund winning bid",
			types.NewMsgRefundBid(
				auctionId,
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
			err := s.keeper.RefundBid(cacheCtx, tc.msg.AuctionId, tc.msg.PoolId, tc.msg.GetBidder())
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
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("100_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.liquidFarm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(50_000_000)), true)
	s.nextBlock()

	s.nextAuction()

	// Ensure that the rewards auction is created
	auction, found := s.keeper.GetLastRewardsAuction(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.placeBid(auction.PoolId, s.addr(1), sdk.NewInt64Coin(pool.PoolCoinDenom, 10_000_000), true)
	s.placeBid(auction.PoolId, s.addr(2), sdk.NewInt64Coin(pool.PoolCoinDenom, 20_000_000), true)
	s.placeBid(auction.PoolId, s.addr(3), sdk.NewInt64Coin(pool.PoolCoinDenom, 30_000_000), true)
	s.nextBlock()

	s.nextAuction()

	// Ensure the state of previous auction is updated
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.ctx), 2)

	// Ensure that two bidders got their pool coin back to their balances
	s.Require().Equal(sdk.NewInt(10_000_000), s.getBalance(s.addr(1), pool.PoolCoinDenom).Amount)
	s.Require().Equal(sdk.NewInt(20_000_000), s.getBalance(s.addr(2), pool.PoolCoinDenom).Amount)
	s.Require().Equal(sdk.ZeroInt(), s.getBalance(s.addr(3), pool.PoolCoinDenom).Amount)

	// Ensure winner's balance to see if they received farming rewards
	s.Require().True(s.getBalance(s.addr(3), "stake").Amount.GT(sdk.NewInt(1)))

	// Ensure newly staked amount by the liquid farm reserve account
	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(sdk.NewInt(80_000_000), position.FarmingAmount)
}

func (s *KeeperTestSuite) TestAfterAllocateRewards_NoBid() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("100_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.liquidFarm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(50_000_000)), true)
	s.nextBlock()

	s.nextAuction()

	auction, found := s.keeper.GetLastRewardsAuction(s.ctx, pool.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.nextAuction()

	auction, found = s.keeper.GetRewardsAuction(s.ctx, auction.PoolId, auction.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusSkipped, auction.Status)
}

// [scenario]
// Chain is started and liquid farm is registered.
// There is no one liquid farmed coin; therefore there is no farming rewards
// Bidders either mistakenly or purposely place their bids
//
// [expected results]
// 1. Winning bid amount is still going to be auto compounded
// 2. Minting amount for LiquidFarm is still 1:1 although winning bid amount is auto compounded
// 3. Bids other than the winning bid must be refunded
func (s *KeeperTestSuite) TestFinishRewardsAuction_NoOneFarmed() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("100_000_000stake"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("2_000_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("3_000_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("4_000_000pool1"), true)
	s.nextBlock()

	// Ensure that there is no farming rewards accumulated
	liquidFarmReserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	farmingRewards := s.app.LPFarmKeeper.Rewards(s.ctx, liquidFarmReserveAddr, pool.PoolCoinDenom)
	s.Require().True(farmingRewards.IsZero())

	s.nextAuction()

	// Ensure that winning bid amount is auto compounded
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, liquidFarmReserveAddr, pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().True(position.FarmingAmount.Equal(sdk.NewInt(4_000_000)))

	// Ensure that bidders other than the winning bid is refunded
	s.Require().True(s.getBalance(s.addr(5), "pool1").Amount.Equal(sdk.NewInt(2_000_000)))
	s.Require().True(s.getBalance(s.addr(6), "pool1").Amount.Equal(sdk.NewInt(3_000_000)))

	// Ensure the auction status
	auction, found := s.keeper.GetRewardsAuction(s.ctx, 1, pool.Id)
	s.Require().True(found)
	s.Require().True(auction.Rewards.IsZero())
	s.Require().Equal(types.AuctionStatusFinished, auction.Status)
	s.Require().Equal(s.addr(7).String(), auction.Winner)
	s.Require().Equal(sdk.NewInt(4_000_000), auction.WinningAmount.Amount)

	s.liquidFarm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(50_000_000)), true)
	s.nextBlock()

	// Ensure that minting amount is still 1:1
	s.Require().True(s.getBalance(s.addr(1), types.LiquidFarmCoinDenom(pool.Id)).Amount.Equal(sdk.NewInt(50_000_000)))

	s.liquidUnfarm(pool.Id, s.addr(1), sdk.NewCoin(types.LiquidFarmCoinDenom(pool.Id), sdk.NewInt(50_000_000)), false)
	s.nextBlock()

	// Ensure that received pool coin amount is greater than the original liquid farm amount
	s.Require().True(s.getBalance(s.addr(1), pool.PoolCoinDenom).Amount.GT(sdk.NewInt(50_000_000)))
}

func (s *KeeperTestSuite) TestRewardsAuction_RewardsAndFees() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(s.addr(0), []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("100_000_000stake"))

	// Fee rate is 10%
	liquidFarm := s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), utils.ParseDec("0.1"))
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(0), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("10_000_000pool1"), true)
	s.nextBlock()

	withdrawnRewardsReserveAddr := types.WithdrawnRewardsReserveAddress(pool.Id)
	s.Require().False(s.getBalances(withdrawnRewardsReserveAddr).IsZero())

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("2_000_000pool1"), true)
	s.nextBlock()

	liquidFarmReserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	farmingRewards := s.app.LPFarmKeeper.Rewards(s.ctx, liquidFarmReserveAddr, pool.PoolCoinDenom)
	truncatedRewards, _ := farmingRewards.TruncateDecimal()
	spendable := s.app.BankKeeper.SpendableCoins(s.ctx, withdrawnRewardsReserveAddr)
	totalRewards := truncatedRewards.Add(spendable...)

	deducted, fees := types.DeductFees(totalRewards, liquidFarm.FeeRate)

	s.nextAuction()

	auction, found := s.keeper.GetRewardsAuction(s.ctx, 1, pool.Id)
	s.Require().True(found)

	s.Require().True(auction.Rewards.IsEqual(deducted.Add(fees...)))
	s.Require().True(auction.Fees.IsEqual(fees))
}
