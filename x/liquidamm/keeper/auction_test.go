package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (s *KeeperTestSuite) TestPlaceBid() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr := utils.TestAddress(1)
	s.MintShare(
		minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()
	s.AdvanceRewardsAuctions() // Starts a rewards auction

	bidderAddr1 := utils.TestAddress(2)
	s.MintShare(
		bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	ctx := s.Ctx

	// public position not found
	s.Ctx, _ = ctx.CacheContext()
	_, err := s.keeper.PlaceBid(
		s.Ctx, bidderAddr1, 3, 1, utils.ParseCoin("1000000sb2"))
	s.Require().EqualError(err, "public position not found: not found")

	// share denom mismatch
	s.Ctx, _ = ctx.CacheContext()
	_, err = s.keeper.PlaceBid(
		s.Ctx, bidderAddr1, publicPosition.Id, 1, utils.ParseCoin("1000000sb2"))
	s.Require().EqualError(err, "share denom != sb1: invalid request")

	// rewards auction not found
	s.Ctx, _ = ctx.CacheContext()
	_, err = s.keeper.PlaceBid(
		s.Ctx, bidderAddr1, publicPosition.Id, 3, utils.ParseCoin("1000000sb1"))
	s.Require().EqualError(err, "rewards auction not found: not found")

	// Skip the current rewards auction(auction id 1).
	s.Ctx = ctx
	s.AdvanceRewardsAuctions()
	ctx = s.Ctx

	auction, found := s.keeper.GetRewardsAuction(s.Ctx, publicPosition.Id, 1)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusSkipped, auction.Status)

	// invalid auction status
	s.Ctx, _ = ctx.CacheContext()
	_, err = s.keeper.PlaceBid(
		s.Ctx, bidderAddr1, publicPosition.Id, 1, utils.ParseCoin("1000000sb1"))
	s.Require().EqualError(
		err, "invalid rewards auction status: AUCTION_STATUS_SKIPPED: invalid request")

	// successfully bids.
	s.Ctx = ctx
	s.PlaceBid(bidderAddr1, publicPosition.Id, 2, utils.ParseCoin("1000000sb1"))
	auction, found = s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.Require().NotNil(auction.WinningBid)
	s.Require().Equal(bidderAddr1.String(), auction.WinningBid.Bidder)
	s.AssertEqual(utils.ParseCoin("1000000sb1"), auction.WinningBid.Share)
	s.AssertEqual(
		utils.ParseCoins("1000000sb1"), s.GetAllBalances(publicPosition.MustGetBidReserveAddress()))
	ctx = s.Ctx

	bidderAddr2 := utils.TestAddress(3)
	s.MintShare(
		bidderAddr2, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	// bidding amount <= winning bid amount
	s.Ctx, _ = ctx.CacheContext()
	_, err = s.keeper.PlaceBid(
		s.Ctx, bidderAddr2, publicPosition.Id, 2, utils.ParseCoin("1000000sb1"))
	s.Require().EqualError(
		err, "share amount must be greater than winning bid's share 1000000: insufficient bid amount")

	// successfully replaces the winning bid.
	s.Ctx = ctx
	s.PlaceBid(bidderAddr2, publicPosition.Id, 2, utils.ParseCoin("2000000sb1"))
	ctx = s.Ctx

	auction, found = s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.Require().NotNil(auction.WinningBid)
	s.Require().Equal(bidderAddr2.String(), auction.WinningBid.Bidder)
	s.AssertEqual(utils.ParseCoin("2000000sb1"), auction.WinningBid.Share)
	s.AssertEqual(
		utils.ParseCoins("2000000sb1"), s.GetAllBalances(publicPosition.MustGetBidReserveAddress()))
}

func (s *KeeperTestSuite) TestRewardsAuction() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	// Start the first rewards auction.
	s.AdvanceRewardsAuctions()

	position := s.App.LiquidAMMKeeper.MustGetAMMPosition(s.Ctx, publicPosition)
	_, farmingRewards, err := s.App.AMMKeeper.CollectibleCoins(s.Ctx, position.Id)
	s.Require().NoError(err)
	s.Require().Equal("5786uatom", farmingRewards.String())

	bidderAddr1 := utils.TestAddress(2)
	bidderShare1, _, _, _ := s.MintShare(bidderAddr1, publicPosition.Id, utils.ParseCoins("10_00000ucre,50_000000uusd"), true)
	bidderAddr2 := utils.TestAddress(3)
	bidderShare2, _, _, _ := s.MintShare(bidderAddr2, publicPosition.Id, utils.ParseCoins("20_00000ucre,100_000000uusd"), true)

	auction, found := s.App.LiquidAMMKeeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.Require().Nil(auction.WinningBid)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, bidderShare1.SubAmount(sdk.NewInt(1000)))
	auction, _ = s.App.LiquidAMMKeeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().Equal(bidderAddr1.String(), auction.WinningBid.Bidder)

	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, bidderShare1) // Update the bid with the higher amount
	auction, _ = s.App.LiquidAMMKeeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().Equal(bidderShare1, auction.WinningBid.Share)

	s.PlaceBid(bidderAddr2, publicPosition.Id, auction.Id, bidderShare2)
	auction, _ = s.App.LiquidAMMKeeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().Equal(bidderAddr2.String(), auction.WinningBid.Bidder)
	s.Require().Equal(bidderShare2, auction.WinningBid.Share)

	// Finish the current rewards auction.
	s.AdvanceRewardsAuctions()
	s.Require().Equal(sdk.NewInt(5768), s.GetAllBalances(bidderAddr2).AmountOf("uatom"))

	auction, _ = s.App.LiquidAMMKeeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().Equal(types.AuctionStatusFinished, auction.Status)
	s.Require().Equal("5786uatom", auction.Rewards.String()) // Rewards before deducting fees
	s.Require().Equal("18uatom", auction.Fees.String())
}

func (s *KeeperTestSuite) TestPlaceBid_Refund() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr1 := utils.TestAddress(1)
	s.MintShare(minterAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	s.AdvanceRewardsAuctions()

	bidderAddr1 := utils.TestAddress(2)
	s.MintShare(bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("100000sb1"))
	s.NextBlock()

	balancesBefore := s.GetAllBalances(bidderAddr1)
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("200000sb1"))
	s.NextBlock()
	balancesAfter := s.GetAllBalances(bidderAddr1)
	s.Require().Equal("100000sb1", balancesBefore.Sub(balancesAfter).String())
}

func (s *KeeperTestSuite) TestAfterRewardsAllocated() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr := utils.TestAddress(1)
	_, _, liquidity, _ := s.MintShare(minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	s.AdvanceRewardsAuctions()

	// Ensure that the rewards auction is created
	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	bidderAddr1 := utils.TestAddress(2)
	bidderAddr2 := utils.TestAddress(3)
	bidderAddr3 := utils.TestAddress(4)
	s.MintShare(bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr2, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr3, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	// Previous share balance
	s.Require().Equal("4357388321sb1", s.GetBalance(bidderAddr1, "sb1").String())
	s.Require().Equal("4357388321sb1", s.GetBalance(bidderAddr2, "sb1").String())
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("100000sb1"))
	s.PlaceBid(bidderAddr2, publicPosition.Id, auction.Id, utils.ParseCoin("200000sb1"))
	s.PlaceBid(bidderAddr3, publicPosition.Id, auction.Id, utils.ParseCoin("300000sb1"))

	s.NextBlock()
	s.AdvanceRewardsAuctions()

	// Ensure that two bidders got their shares back to their balances
	s.Require().Equal("4357388321sb1", s.GetBalance(bidderAddr1, "sb1").String())
	s.Require().Equal("4357388321sb1", s.GetBalance(bidderAddr2, "sb1").String())
	s.Require().True(s.GetBalance(bidderAddr3, "uatom").Amount.GT(sdk.NewInt(1)))

	// One more epoch should be advanced
	s.NextBlock()
	s.AdvanceRewardsAuctions()

	// Ensure liquidity per share increased due to the auction result
	removedLiquidity, _, _ := s.BurnShare(minterAddr, publicPosition.Id, s.GetBalance(minterAddr, "sb1"))
	s.Require().True(removedLiquidity.GT(liquidity))
}

func (s *KeeperTestSuite) TestAuctionSkipped() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	s.NextBlock()
	s.AdvanceRewardsAuctions()

	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusStarted, auction.Status)

	s.AdvanceRewardsAuctions()

	auction, found = s.keeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().True(found)
	s.Require().Equal(types.AuctionStatusSkipped, auction.Status)
}

func (s *KeeperTestSuite) TestRewardsAuction_RewardsAndFees() {
	publicPosition := s.CreateSamplePublicPosition()
	s.NextBlock()

	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	s.AdvanceRewardsAuctions()

	bidderAddr1 := utils.TestAddress(2)
	s.MintShare(bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	auction, _ := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("100000sb1"))
	s.NextBlock()

	position := s.keeper.MustGetAMMPosition(s.Ctx, publicPosition)
	_, farmingRewards, err := s.App.AMMKeeper.CollectibleCoins(s.Ctx, position.Id)
	s.Require().NoError(err)

	deducted, fees := types.DeductFees(farmingRewards, publicPosition.FeeRate)

	s.AdvanceRewardsAuctions()

	auction, _ = s.keeper.GetRewardsAuction(s.Ctx, publicPosition.Id, auction.Id)
	s.Require().True(auction.Rewards.IsEqual(deducted.Add(fees...)))
	s.Require().True(auction.Fees.IsEqual(fees))
}

func (s *KeeperTestSuite) TestMaxNumRecentRewardsAuctions() {
	s.keeper.SetMaxNumRecentRewardsAuctions(s.Ctx, 5)

	market1 := s.CreateMarket("ucre", "uusd")
	pool1 := s.CreatePool(market1.Id, utils.ParseDec("5"))
	publicPosition1 := s.CreatePublicPosition(
		pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseDec("0.003"))
	market2 := s.CreateMarket("uusd", "ucre")
	pool2 := s.CreatePool(market2.Id, utils.ParseDec("0.2"))
	publicPosition2 := s.CreatePublicPosition(
		pool2.Id, utils.ParseDec("0.1"), utils.ParseDec("0.3"),
		utils.ParseDec("0.003"))

	s.NextBlock()

	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, publicPosition1.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(minterAddr, publicPosition2.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	for i := 0; i < 10; i++ {
		s.AdvanceRewardsAuctions()
	}

	cnt := 0
	s.keeper.IterateRewardsAuctionsByPublicPosition(s.Ctx, publicPosition1.Id, func(auction types.RewardsAuction) (stop bool) {
		cnt++
		return false
	})
	s.Require().Equal(5+1, cnt) // 5 recent auctions + 1 current auction
	cnt = 0
	s.keeper.IterateRewardsAuctionsByPublicPosition(s.Ctx, publicPosition2.Id, func(auction types.RewardsAuction) (stop bool) {
		cnt++
		return false
	})
	s.Require().Equal(5+1, cnt)

	// Check again
	for i := 0; i < 10; i++ {
		s.AdvanceRewardsAuctions()
	}
	cnt = 0
	s.keeper.IterateRewardsAuctionsByPublicPosition(s.Ctx, publicPosition1.Id, func(auction types.RewardsAuction) (stop bool) {
		cnt++
		return false
	})
	s.Require().Equal(5+1, cnt) // 5 recent auctions + 1 current auction
	cnt = 0
	s.keeper.IterateRewardsAuctionsByPublicPosition(s.Ctx, publicPosition2.Id, func(auction types.RewardsAuction) (stop bool) {
		cnt++
		return false
	})
	s.Require().Equal(5+1, cnt)
}

func (s *KeeperTestSuite) TestPlaceBid_RefundPreviousWinningBid() {
	publicPosition := s.CreateSamplePublicPosition()
	s.AdvanceRewardsAuctions()
	s.NextBlock()
	s.NextBlock()

	bidderAddr1 := utils.TestAddress(1)
	bidderAddr2 := utils.TestAddress(2)
	s.MintShare(
		bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(
		bidderAddr2, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	s.AssertEqual(utils.ParseCoin("4357388321sb1"), s.GetBalance(bidderAddr1, "sb1"))

	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)

	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("1000000sb1"))
	s.AssertEqual(utils.ParseCoin("4356388321sb1"), s.GetBalance(bidderAddr1, "sb1")) // 1000000sb1 locked
	s.Require().Len(s.keeper.GetAllBids(s.Ctx), 1)

	// A bidder modifies its bid
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("1500000sb1"))
	// Previous winning bid refunded.
	s.AssertEqual(utils.ParseCoin("4355888321sb1"), s.GetBalance(bidderAddr1, "sb1")) // 1500000sb1 locked
	s.Require().Len(s.keeper.GetAllBids(s.Ctx), 1)

	// Another bidder places new winning bid
	s.PlaceBid(bidderAddr2, publicPosition.Id, auction.Id, utils.ParseCoin("2000000sb1"))
	// Previous winning bid refunded.
	s.AssertEqual(utils.ParseCoin("4357388321sb1"), s.GetBalance(bidderAddr1, "sb1")) // all refunded
	s.AssertEqual(utils.ParseCoin("4355388321sb1"), s.GetBalance(bidderAddr2, "sb1")) // 2000000sb1 locked
	s.Require().Len(s.keeper.GetAllBids(s.Ctx), 1)
}
