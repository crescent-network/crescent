package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

var _ types.FarmingHooks = (*MockFarmingHooksReceiver)(nil)

// MockFarmingHooksReceiver event hooks for farming object (noalias)
type MockFarmingHooksReceiver struct {
	AfterAllocateRewardsValid bool
}

func (h *MockFarmingHooksReceiver) AfterAllocateRewards(ctx sdk.Context) {
	h.AfterAllocateRewardsValid = true
}

func (s *KeeperTestSuite) TestHooks() {
	farmingHooksReceiver := MockFarmingHooksReceiver{}

	// Set hooks
	keeper.UnsafeSetHooks(
		&s.keeper, types.NewMultiFarmingHooks(&farmingHooksReceiver),
	)

	// Default must be false
	s.Require().False(farmingHooksReceiver.AfterAllocateRewardsValid)

	// Create sample farming plan
	s.CreateFixedAmountPlan(s.addrs[5], map[string]string{denom1: "1"}, map[string]int64{denom3: 700000})

	// Stake
	s.Stake(s.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))

	// Advanced epoch twice to trigger AllocateRewards function
	s.advanceEpochDays()
	s.advanceEpochDays()

	// Must be true
	s.Require().True(farmingHooksReceiver.AfterAllocateRewardsValid)
}
