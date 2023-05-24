package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
)

// var (
//
//	helperAddr      = utils.TestAddress(10000)
//	sampleStartTime = utils.ParseTime("0001-01-01T00:00:00Z")
//	sampleEndTime   = utils.ParseTime("9999-12-31T23:59:59Z")
//
// )
type KeeperTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supply
}

//
//func (s *KeeperTestSuite) createLiquidFarm(poolId uint64, minFarmAmt, minBidAmt sdk.Int, feeRate sdk.Dec) types.LiquidFarm {
//	s.T().Helper()
//	liquidFarm := types.NewLiquidFarm(poolId, minFarmAmt, minBidAmt, feeRate)
//	params := s.keeper.GetParams(s.ctx)
//	params.LiquidFarms = append(params.LiquidFarms, liquidFarm)
//	s.keeper.SetParams(s.ctx, params)
//	s.keeper.SetLiquidFarm(s.ctx, liquidFarm)
//	return liquidFarm
//}
//
//func (s *KeeperTestSuite) createRewardsAuction(poolId uint64) {
//	s.T().Helper()
//	duration := s.keeper.GetRewardsAuctionDuration(s.ctx)
//	s.keeper.CreateRewardsAuction(s.ctx, poolId, s.ctx.BlockTime().Add(duration*time.Hour))
//}
//
//func (s *KeeperTestSuite) liquidFarm(poolId uint64, farmer sdk.AccAddress, lpCoin sdk.Coin, fund bool) {
//	s.T().Helper()
//	if fund {
//		s.fundAddr(farmer, sdk.NewCoins(lpCoin))
//	}
//	err := s.keeper.LiquidFarm(s.ctx, poolId, farmer, lpCoin)
//	s.Require().NoError(err)
//}
//
//func (s *KeeperTestSuite) liquidUnfarm(poolId uint64, farmer sdk.AccAddress, lfCoin sdk.Coin, fund bool) {
//	s.T().Helper()
//	if fund {
//		s.fundAddr(farmer, sdk.NewCoins(lfCoin))
//	}
//	_, err := s.keeper.LiquidUnfarm(s.ctx, poolId, farmer, lfCoin)
//	s.Require().NoError(err)
//}
//
//func (s *KeeperTestSuite) placeBid(poolId uint64, bidder sdk.AccAddress, biddingCoin sdk.Coin, fund bool) types.Bid {
//	s.T().Helper()
//	if fund {
//		s.fundAddr(bidder, sdk.NewCoins(biddingCoin))
//	}
//
//	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, poolId)
//	bid, err := s.keeper.PlaceBid(s.ctx, auctionId, poolId, bidder, biddingCoin)
//	s.Require().NoError(err)
//
//	return bid
//}
