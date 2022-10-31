package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	pair := s.createPair(helperAddr, "denom1", "denom2")
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), false)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), false)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), false)
	s.nextBlock()

	// Ensure that the position is created and the amount of farmed coin
	position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, types.LiquidFarmReserveAddress(pool.Id), pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(sdk.NewInt(900_000), position.FarmingAmount) // 100+300+500

	// Move time to auctionTime so that rewards auction is created
	s.nextAuction()

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found := s.keeper.GetRewardsAuction(s.ctx, auctionId, pool.Id)
	s.Require().True(found)

	s.placeBid(auction.PoolId, s.addr(4), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(5), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(6), utils.ParseCoin("150_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(7), utils.ParseCoin("200_000pool1"), true)

	// Finish the first auction and create the second rewards auction
	s.nextAuction()

	// Ensure that the amount of farmed coin is increased due to the auction
	position, found = s.app.LPFarmKeeper.GetPosition(s.ctx, types.LiquidFarmReserveAddress(pool.Id), pool.PoolCoinDenom)
	s.Require().True(found)
	s.Require().Equal(sdk.NewInt(1_100_000), position.FarmingAmount) // 100+300+500+200

	// Ensure that the second rewards auction is created
	auctionId = s.keeper.GetLastRewardsAuctionId(s.ctx, pool.Id)
	auction, found = s.keeper.GetRewardsAuction(s.ctx, auctionId, pool.Id)
	s.Require().True(found)

	s.liquidFarm(auction.PoolId, s.addr(1), utils.ParseCoin("200_000pool1"), true)
	s.liquidFarm(auction.PoolId, s.addr(2), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.placeBid(auction.PoolId, s.addr(6), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(7), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(4), utils.ParseCoin("150_000pool1"), true)
	s.placeBid(auction.PoolId, s.addr(5), utils.ParseCoin("300_000pool1"), true)
	s.nextBlock()

	// Export genesis state and verify
	var genState *types.GenesisState
	s.Require().NotPanics(func() {
		genState = s.keeper.ExportGenesis(s.ctx)
		s.Require().Len(genState.RewardsAuctions, 2)
		s.Require().Len(genState.Bids, 4)
		s.Require().Len(genState.WinningBidRecords, 1)
	})
	s.Require().NoError(genState.Validate())

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)

	var genState3 *types.GenesisState
	s.Require().NotPanics(func() {
		genState3 = s.keeper.ExportGenesis(s.ctx)
		s.Require().Equal(*genState, *genState3)
	})
	s.Require().NoError(genState3.Validate())
}
