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

func (s *TestSuite) AddLiquidity(ownerAddr, fromAddr sdk.AccAddress, poolId uint64, lowerPrice, upperPrice sdk.Dec, desiredAmt sdk.Coins) (position ammtypes.Position, liquidity sdk.Int, amt sdk.Coins) {
	s.T().Helper()
	var err error
	position, liquidity, amt, err = s.App.AMMKeeper.AddLiquidity(s.Ctx, ownerAddr, fromAddr, poolId, lowerPrice, upperPrice, desiredAmt)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) RemoveLiquidity(ownerAddr, toAddr sdk.AccAddress, positionId uint64, liquidity sdk.Int) (position ammtypes.Position, amt sdk.Coins) {
	s.T().Helper()
	var err error
	position, amt, err = s.App.AMMKeeper.RemoveLiquidity(s.Ctx, ownerAddr, toAddr, positionId, liquidity)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) Collect(ownerAddr, toAddr sdk.AccAddress, positionId uint64, amt sdk.Coins) {
	s.T().Helper()
	s.Require().NoError(s.App.AMMKeeper.Collect(s.Ctx, ownerAddr, toAddr, positionId, amt))
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
	if initialFunds.IsAllPositive() {
		s.FundAccount(sdk.MustAccAddressFromBech32(plan.FarmingPoolAddress), initialFunds)
	}
	return plan
}

func (s *TestSuite) CreatePublicFarmingPlan(description string, farmingPoolAddr sdk.AccAddress, termAddr sdk.AccAddress, rewardAllocs []ammtypes.FarmingRewardAllocation, startTime, endTime time.Time) (plan ammtypes.FarmingPlan) {
	s.T().Helper()
	var err error
	plan, err = s.App.AMMKeeper.CreatePublicFarmingPlan(
		s.Ctx, description, farmingPoolAddr, termAddr, rewardAllocs, startTime, endTime)
	s.Require().NoError(err)
	return plan
}

func (s *TestSuite) CollectibleCoins(positionId uint64) (fee, farmingRewards sdk.Coins) {
	s.T().Helper()
	var err error
	fee, farmingRewards, err = s.App.AMMKeeper.CollectibleCoins(s.Ctx, positionId)
	s.Require().NoError(err)
	return
}
