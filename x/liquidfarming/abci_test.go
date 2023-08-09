package liquidfarming_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *ModuleTestSuite) TestRewardsAuctionEndTimeTest() {
	pair := s.createPairWithLastPrice(s.addr(0), "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	params := s.keeper.GetParams(s.ctx)
	params.RewardsAuctionDuration = 8 * time.Hour
	params.LiquidFarms = []types.LiquidFarm{
		types.NewLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec()),
	}
	s.keeper.SetParams(s.ctx, params)

	_, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().False(found)

	// Set chain launch time
	currTime := utils.ParseTime("2022-10-01T23:00:00Z")
	s.ctx = s.ctx.WithBlockTime(currTime)
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	// Ensure that the end time is initialized
	endTime, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T00:00:00Z"), endTime)

	for i := 0; i < 1; i++ {
		t := s.ctx.BlockTime().Add(1 * time.Hour)
		s.ctx = s.ctx.WithBlockTime(t)
		liquidfarming.BeginBlocker(s.ctx, s.keeper)
	}

	// Ensure the length of auction
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.ctx), 1)

	// Ensure that the end time
	endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T08:00:00Z"), endTime)

	for i := 0; i < 9; i++ {
		t := s.ctx.BlockTime().Add(1 * time.Hour)
		s.ctx = s.ctx.WithBlockTime(t)
		liquidfarming.BeginBlocker(s.ctx, s.keeper)
	}

	// Ensure the length of auction
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.ctx), 2)

	// Ensure that the end time
	endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T16:00:00Z"), endTime)
}

// TestChainHaltedForShortTime covers a case when chain is halted less than one duration.
func (s *ModuleTestSuite) TestChainHaltedForShortTime() {
	pair := s.createPairWithLastPrice(s.addr(0), "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	params := s.keeper.GetParams(s.ctx)
	params.RewardsAuctionDuration = 8 * time.Hour
	params.LiquidFarms = []types.LiquidFarm{
		types.NewLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec()),
	}
	s.keeper.SetParams(s.ctx, params)

	// Set chain launch time
	currTime := utils.ParseTime("2022-10-01T23:00:00Z")
	s.ctx = s.ctx.WithBlockTime(currTime)
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	endTime, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T00:00:00Z"), endTime)

	// First RewardsAuction is created
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(1 * time.Hour))
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.ctx), 1)

	endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T08:00:00Z"), endTime)

	// 9 hours passed after chain is halted
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(9 * time.Hour))
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T16:00:00Z"), endTime)
}

// TestChainHaltedForLongTime covers a case when chain is halted more than at least one duration.
func (s *ModuleTestSuite) TestChainHaltedForLongTime() {
	params := s.keeper.GetParams(s.ctx)
	params.RewardsAuctionDuration = 8 * time.Hour
	s.keeper.SetParams(s.ctx, params)

	// Set chain launch time
	currTime := utils.ParseTime("2022-10-01T23:00:00Z")
	s.ctx = s.ctx.WithBlockTime(currTime)
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	endTime, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-02T00:00:00Z"), endTime)

	// 48 hours passed after chain is halted
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(48 * time.Hour))
	liquidfarming.BeginBlocker(s.ctx, s.keeper)

	// Ensure that new end time is set to 00:00UTC next day
	endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	s.Require().True(found)
	s.Require().Equal(utils.ParseTime("2022-10-04T00:00:00Z"), endTime)
}

func (s *ModuleTestSuite) TestChangingRewardsAuctionDurationTest() {
	durationTest := func(before, after time.Duration) {
		s.SetupTest()

		params := s.keeper.GetParams(s.ctx)
		params.RewardsAuctionDuration = before
		s.keeper.SetParams(s.ctx, params)

		currTime := utils.ParseTime("2022-10-01T23:00:00Z")
		s.ctx = s.ctx.WithBlockTime(currTime)
		liquidfarming.BeginBlocker(s.ctx, s.keeper)

		endTime, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
		s.Require().True(found)
		s.Require().Equal(utils.ParseTime("2022-10-02T00:00:00Z"), endTime)

		// 2 days passed
		for i := 0; i < 48; i++ {
			t := s.ctx.BlockTime().Add(1 * time.Hour)
			s.ctx = s.ctx.WithBlockTime(t)
			liquidfarming.BeginBlocker(s.ctx, s.keeper)
		}

		params = s.keeper.GetParams(s.ctx)
		params.RewardsAuctionDuration = after
		s.keeper.SetParams(s.ctx, params)

		endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
		s.Require().True(found)

		s.ctx = s.ctx.WithBlockTime(endTime)
		liquidfarming.BeginBlocker(s.ctx, s.keeper)

		expEndTime := endTime.Add(after)

		endTime, found = s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
		s.Require().True(found)
		s.Require().Equal(expEndTime, endTime)
	}

	// Change from 1 to 8 hours
	durationTest(1*time.Hour, 8*time.Hour)

	// Change from 8 to 1 hours
	durationTest(8*time.Hour, 1*time.Hour)

	// Change from 1 hour 15 minutes to 6 hours
	durationTest(75*time.Minute, 6*time.Hour)

	// Change from 50 seconds to 80 seconds
	durationTest(50*time.Second, 80*time.Second)
}
