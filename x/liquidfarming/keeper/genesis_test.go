package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.farm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.farm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.farm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	// Trigger AllocateRewards hook to create the first rewards auction
	s.advanceEpochDays()

	s.placeBid(pool.Id, s.addr(4), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("150_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("200_000pool1"), true)

	// Finish the first auction and create the second rewards auction
	s.advanceEpochDays()

	s.farm(pool.Id, s.addr(1), utils.ParseCoin("200_000pool1"), true)
	s.farm(pool.Id, s.addr(2), utils.ParseCoin("200_000pool1"), true)
	s.nextBlock()

	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(4), utils.ParseCoin("150_000pool1"), true)
	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("200_000pool1"), true)

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
