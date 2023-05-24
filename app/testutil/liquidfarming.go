package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	liquidfarmingtypes "github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (s *TestSuite) CreateLiquidFarm(poolId uint64, lowerPrice, upperPrice sdk.Dec, minBidAmt sdk.Int, feeRate sdk.Dec) (liquidFarm liquidfarmingtypes.LiquidFarm) {
	s.T().Helper()
	var err error
	liquidFarm, err = s.App.LiquidFarmingKeeper.CreateLiquidFarm(s.Ctx, poolId, lowerPrice, upperPrice, minBidAmt, feeRate)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) MintShare(senderAddr sdk.AccAddress, liquidFarmId uint64, desiredAmt sdk.Coins) (mintedShare sdk.Coin, position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins) {
	s.T().Helper()
	var err error
	mintedShare, position, liquidity, amt, err = s.App.LiquidFarmingKeeper.MintShare(s.Ctx, senderAddr, liquidFarmId, desiredAmt)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) BurnShare(senderAddr sdk.AccAddress, liquidFarmId uint64, share sdk.Coin) (removedLiquidity sdk.Int, amt sdk.Coins) {
	s.T().Helper()
	var err error
	removedLiquidity, amt, err = s.App.LiquidFarmingKeeper.BurnShare(s.Ctx, senderAddr, liquidFarmId, share)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceBid(bidderAddr sdk.AccAddress, liquidFarmId, auctionId uint64, share sdk.Coin) (bid liquidfarmingtypes.Bid) {
	s.T().Helper()
	var err error
	bid, err = s.App.LiquidFarmingKeeper.PlaceBid(s.Ctx, bidderAddr, liquidFarmId, auctionId, share)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) CancelBid(bidderAddr sdk.AccAddress, liquidFarmId, auctionId uint64) (bid liquidfarmingtypes.Bid) {
	s.T().Helper()
	var err error
	bid, err = s.App.LiquidFarmingKeeper.CancelBid(s.Ctx, bidderAddr, liquidFarmId, auctionId)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) AdvanceRewardsAuctions() {
	s.T().Helper()
	nextEndTime := s.Ctx.BlockTime().Add(s.App.LiquidFarmingKeeper.GetRewardsAuctionDuration(s.Ctx))
	s.Require().NoError(s.App.LiquidFarmingKeeper.AdvanceRewardsAuctions(s.Ctx, nextEndTime))
}
