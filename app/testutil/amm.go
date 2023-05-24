package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *TestSuite) CreatePool(creatorAddr sdk.AccAddress, marketId uint64, price sdk.Dec, fundFee bool) ammtypes.Pool {
	s.T().Helper()
	if fundFee {
		s.FundAccount(creatorAddr, s.App.AMMKeeper.GetPoolCreationFee(s.Ctx))
	}
	pool, err := s.App.AMMKeeper.CreatePool(s.Ctx, creatorAddr, marketId, price)
	s.Require().NoError(err)
	return pool
}

func (s *TestSuite) AddLiquidity(ownerAddr sdk.AccAddress, poolId uint64, lowerPrice, upperPrice sdk.Dec, desiredAmt sdk.Coins) (position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins) {
	s.T().Helper()
	var err error
	position, liquidity, amt, err = s.App.AMMKeeper.AddLiquidity(s.Ctx, ownerAddr, poolId, lowerPrice, upperPrice, desiredAmt)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) RemoveLiquidity(ownerAddr sdk.AccAddress, positionId uint64, liquidity sdk.Int) (position ammtypes.Position, amt sdk.Coins) {
	s.T().Helper()
	var err error
	position, amt, err = s.App.AMMKeeper.RemoveLiquidity(s.Ctx, ownerAddr, positionId, liquidity)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) Collect(ownerAddr sdk.AccAddress, positionId uint64, amt sdk.Coins) {
	s.T().Helper()
	s.Require().NoError(s.App.AMMKeeper.Collect(s.Ctx, ownerAddr, positionId, amt))
}

func (s *TestSuite) CreatePrivateFarmingPlan(creatorAddr sdk.AccAddress, description string, termAddr sdk.AccAddress, rewardAllocs []ammtypes.FarmingRewardAllocation, startTime, endTime time.Time, initialFunds sdk.Coins, fundFee bool) (plan ammtypes.FarmingPlan) {
	s.T().Helper()
	if fundFee {
		s.FundAccount(creatorAddr, s.App.AMMKeeper.GetPrivateFarmingPlanCreationFee(s.Ctx))
	}
	var err error
	plan, err = s.App.AMMKeeper.CreatePrivateFarmingPlan(
		s.Ctx, creatorAddr, description, termAddr, rewardAllocs, startTime, endTime)
	s.Require().NoError(err)
	s.FundAccount(sdk.MustAccAddressFromBech32(plan.FarmingPoolAddress), initialFunds)
	return plan
}
