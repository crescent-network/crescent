package liquidfarming_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/keeper"
)

type ModuleTestSuite struct {
	testutil.TestSuite

	keeper  keeper.Keeper
	querier keeper.Querier
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (s *ModuleTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supply
	s.keeper = s.App.LiquidFarmingKeeper
	s.querier = keeper.Querier{Keeper: s.App.LiquidFarmingKeeper}
}

func (s *ModuleTestSuite) TestRewardsAuctionEndTime() {
	s.keeper.SetRewardsAuctionDuration(s.Ctx, 8*time.Hour)
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	s.CreateLiquidFarm(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))

	// Ensure that the end time is initialized
	endTime, found := s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T00:00:00Z"), endTime)

	s.Ctx = s.Ctx.WithBlockTime(utils.ParseTime("2023-01-02T00:00:00Z"))
	s.NextBlock()

	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 1)

	// Ensure the end time
	endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T08:00:00Z"), endTime)

	for i := 0; i < 8; i++ {
		t := s.Ctx.BlockTime().Add(1 * time.Hour)
		s.Ctx = s.Ctx.WithBlockTime(t)
		s.NextBlock()
	}

	// Ensure the length of auction
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 2)

	// Ensure the end time
	endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T16:00:00Z"), endTime)
}

// TestChainHaltedForShortTime covers a case when chain is halted less than one duration.
func (s *ModuleTestSuite) TestChainHaltedForShortTime() {
	s.keeper.SetRewardsAuctionDuration(s.Ctx, 8*time.Hour)
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	s.CreateLiquidFarm(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))

	endTime, found := s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T00:00:00Z"), endTime)

	s.Ctx = s.Ctx.WithBlockTime(utils.ParseTime("2023-01-02T00:00:00Z"))
	s.NextBlock()

	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 1)

	endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T08:00:00Z"), endTime)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(9 * time.Hour))
	s.NextBlock()

	endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T16:00:00Z"), endTime)

	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 2)
}

// TestChainHaltedForLongTime covers a case when chain is halted more than at least one duration.
func (s *ModuleTestSuite) TestChainHaltedForLongTime() {
	s.keeper.SetRewardsAuctionDuration(s.Ctx, 8*time.Hour)

	endTime, found := s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-02T00:00:00Z"), endTime)

	// 48 hours passed after chain is halted
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(48 * time.Hour))
	s.NextBlock()

	// Ensure that new end time is set to 00:00UTC next day
	endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2023-01-04T00:00:00Z"), endTime)
}

func (s *ModuleTestSuite) TestChangingRewardsAuctionDurationTest() {
	durationTest := func(before, after time.Duration) {
		s.SetupTest()

		s.keeper.SetRewardsAuctionDuration(s.Ctx, before)

		endTime, found := s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
		s.Require().True(found)
		s.Require().Equal(utils.ParseTime("2023-01-02T00:00:00Z"), endTime)

		// 2 days passed
		for i := 0; i < 48; i++ {
			t := s.Ctx.BlockTime().Add(1 * time.Hour)
			s.Ctx = s.Ctx.WithBlockTime(t)
			s.EndBlock()
			s.BeginBlock(0)
		}

		s.keeper.SetRewardsAuctionDuration(s.Ctx, after)

		endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
		s.Require().True(found)

		s.Ctx = s.Ctx.WithBlockTime(endTime)
		s.EndBlock()
		s.BeginBlock(0)

		expEndTime := endTime.Add(after)

		endTime, found = s.keeper.GetNextRewardsAuctionEndTime(s.Ctx)
		s.Require().True(found)
		s.Require().Equal(expEndTime, endTime)
	}

	// Change from 1 to 8 hours
	durationTest(1*time.Hour, 8*time.Hour)

	// Change from 8 to 1 hour
	durationTest(8*time.Hour, 1*time.Hour)

	// Change from 1 hour 15 minutes to 6 hours
	durationTest(75*time.Minute, 6*time.Hour)

	// Change from 50 seconds to 80 seconds
	durationTest(50*time.Second, 80*time.Second)
}
