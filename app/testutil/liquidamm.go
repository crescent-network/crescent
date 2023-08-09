package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	liquidammtypes "github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (s *TestSuite) CreatePublicPosition(poolId uint64, lowerPrice, upperPrice sdk.Dec, minBidAmt sdk.Int, feeRate sdk.Dec) (publicPosition liquidammtypes.PublicPosition) {
	s.T().Helper()
	var err error
	publicPosition, err = s.App.LiquidAMMKeeper.CreatePublicPosition(s.Ctx, poolId, lowerPrice, upperPrice, minBidAmt, feeRate)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) MintShare(senderAddr sdk.AccAddress, publicPositionId uint64, desiredAmt sdk.Coins, fund bool) (mintedShare sdk.Coin, position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins) {
	s.T().Helper()
	if fund {
		s.FundAccount(senderAddr, desiredAmt)
	}
	var err error
	mintedShare, position, liquidity, amt, err = s.App.LiquidAMMKeeper.MintShare(s.Ctx, senderAddr, publicPositionId, desiredAmt)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) BurnShare(senderAddr sdk.AccAddress, publicPositionId uint64, share sdk.Coin) (removedLiquidity sdk.Int, position ammtypes.Position, amt sdk.Coins) {
	s.T().Helper()
	var err error
	removedLiquidity, position, amt, err = s.App.LiquidAMMKeeper.BurnShare(s.Ctx, senderAddr, publicPositionId, share)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceBid(bidderAddr sdk.AccAddress, positionId, auctionId uint64, share sdk.Coin) (bid liquidammtypes.Bid) {
	s.T().Helper()
	var err error
	bid, err = s.App.LiquidAMMKeeper.PlaceBid(s.Ctx, bidderAddr, positionId, auctionId, share)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) AdvanceRewardsAuctions() {
	s.T().Helper()
	nextEndTime := s.Ctx.BlockTime().Add(s.App.LiquidAMMKeeper.GetRewardsAuctionDuration(s.Ctx))
	s.Require().NoError(s.App.LiquidAMMKeeper.AdvanceRewardsAuctions(s.Ctx, nextEndTime))
}
