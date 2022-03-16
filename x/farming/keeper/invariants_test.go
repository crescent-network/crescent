package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingkeeper "github.com/crescent-network/crescent/x/farming/keeper"
	"github.com/crescent-network/crescent/x/farming/types"
)

func (suite *KeeperTestSuite) TestPositiveStakingAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	// Normal staking
	k.SetStaking(ctx, denom1, suite.addrs[0], types.Staking{
		Amount:        sdk.NewInt(1000000),
		StartingEpoch: 1,
	})
	_, broken := farmingkeeper.PositiveStakingAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Zero-amount staking
	k.SetStaking(ctx, denom1, suite.addrs[1], types.Staking{
		Amount:        sdk.ZeroInt(),
		StartingEpoch: 1,
	})
	_, broken = farmingkeeper.PositiveStakingAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Negative-amount staking
	k.SetStaking(ctx, denom1, suite.addrs[1], types.Staking{
		Amount:        sdk.NewInt(-1),
		StartingEpoch: 1,
	})
	_, broken = farmingkeeper.PositiveStakingAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestPositiveQueuedStakingAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	// Normal queued staking
	k.SetQueuedStaking(ctx, denom1, suite.addrs[0], types.QueuedStaking{
		Amount: sdk.NewInt(1000000),
	})
	_, broken := farmingkeeper.PositiveQueuedStakingAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Zero-amount queued staking
	k.SetQueuedStaking(ctx, denom1, suite.addrs[1], types.QueuedStaking{
		Amount: sdk.ZeroInt(),
	})
	_, broken = farmingkeeper.PositiveQueuedStakingAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Negative-amount queued staking
	k.SetQueuedStaking(ctx, denom1, suite.addrs[1], types.QueuedStaking{
		Amount: sdk.NewInt(-1),
	})
	_, broken = farmingkeeper.PositiveQueuedStakingAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestStakingReservedAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))

	// Check staked/queued coin amounts.
	suite.Require().True(coinsEq(
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)),
		k.GetAllStakedCoinsByFarmer(ctx, suite.addrs[0]),
	))
	suite.Require().True(coinsEq(
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)),
		k.GetAllQueuedCoinsByFarmer(ctx, suite.addrs[0]),
	))

	// This is normal state, must not be broken.
	_, broken := farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	staking, _ := k.GetStaking(ctx, denom1, suite.addrs[0])

	// Staking amount in the store <= balance of staking reserve acc. This should be OK.
	staking.Amount = sdk.NewInt(999999)
	k.SetStaking(ctx, denom1, suite.addrs[0], staking)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Staking amount in the store > balance of staking reserve acc. This shouldn't be OK.
	staking.Amount = sdk.NewInt(1000001)
	k.SetStaking(ctx, denom1, suite.addrs[0], staking)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Reset to the original state.
	staking.Amount = sdk.NewInt(1000000)
	k.SetStaking(ctx, denom1, suite.addrs[0], staking)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins into the staking reserve acc.
	// Staking amount in the store <= balance of staking reserve acc. This should be OK.
	err := suite.app.BankKeeper.SendCoins(
		ctx, suite.addrs[1], types.StakingReserveAcc(denom1), sdk.NewCoins(sdk.NewInt64Coin(denom1, 1)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins from staking reserve acc to another acc.
	// Staking amount in the store < balance of staking reserve acc. This shouldn't be OK.
	err = suite.app.BankKeeper.SendCoins(
		ctx, types.StakingReserveAcc(denom1), suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 2)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestRemainingRewardsAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.AdvanceEpoch()
	suite.AdvanceEpoch()

	_, broken := farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Withdrawable rewards amount in the store > balance of rewards reserve acc.
	// Should not be OK.
	k.SetHistoricalRewards(ctx, denom1, 2, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 3)),
	})
	_, broken = farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Withdrawable rewards amount in the store <= balance of rewards reserve acc.
	// Should be OK.
	k.SetHistoricalRewards(ctx, denom1, 2, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 1)),
	})
	_, broken = farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Reset.
	k.SetHistoricalRewards(ctx, denom1, 2, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 2)),
	})
	_, broken = farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins into the rewards reserve acc.
	// Should be OK.
	err := suite.app.BankKeeper.SendCoins(
		ctx, suite.addrs[1], types.RewardsReserveAcc, sdk.NewCoins(sdk.NewInt64Coin(denom3, 1)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins from the rewards reserve acc to another acc.
	// Should not be OK.
	err = suite.app.BankKeeper.SendCoins(
		ctx, types.RewardsReserveAcc, suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 2)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.RemainingRewardsAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestNonNegativeOutstandingRewardsInvariant() {
	k, ctx := suite.keeper, suite.ctx

	k.SetOutstandingRewards(ctx, denom1, types.OutstandingRewards{
		Rewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 1000000)),
	})
	_, broken := farmingkeeper.NonNegativeOutstandingRewardsInvariant(k)(ctx)
	suite.Require().False(broken)

	// Zero-amount outstanding rewards
	// It's acceptable, and for the initial epoch, the outstanding rewards is set to 0.
	k.SetOutstandingRewards(ctx, denom2, types.OutstandingRewards{
		Rewards: sdk.DecCoins{},
	})
	_, broken = farmingkeeper.NonNegativeOutstandingRewardsInvariant(k)(ctx)
	suite.Require().False(broken)

	// Delete the zero-amount outstanding rewards.
	k.DeleteOutstandingRewards(ctx, denom2)

	// Negative-amount outstanding rewards
	// This should not be OK.
	k.SetOutstandingRewards(ctx, denom2, types.OutstandingRewards{
		Rewards: sdk.DecCoins{sdk.DecCoin{Denom: denom3, Amount: sdk.NewDec(-1)}},
	})
	_, broken = farmingkeeper.NonNegativeOutstandingRewardsInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestOutstandingRewardsAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.AdvanceEpoch()

	_, broken := farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Outstanding rewards amount > balance of rewards reserve acc.
	// Should not be OK.
	k.SetOutstandingRewards(ctx, denom1, types.OutstandingRewards{
		Rewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 1000001)),
	})
	_, broken = farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Outstanding rewards amount <= balance of rewards reserve acc.
	// Should be OK.
	k.SetOutstandingRewards(ctx, denom1, types.OutstandingRewards{
		Rewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 999999)),
	})
	_, broken = farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Reset.
	k.SetOutstandingRewards(ctx, denom1, types.OutstandingRewards{
		Rewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 1000000)),
	})
	_, broken = farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins into the rewards reserve acc. Should be OK.
	err := suite.app.BankKeeper.SendCoins(
		ctx, suite.addrs[1], types.RewardsReserveAcc, sdk.NewCoins(sdk.NewInt64Coin(denom3, 1)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins from the rewards reserve acc to another acc. Should not be OK.
	err = suite.app.BankKeeper.SendCoins(
		ctx, types.RewardsReserveAcc, suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom3, 2)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.OutstandingRewardsAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestNonNegativeHistoricalRewardsInvariant() {
	k, ctx := suite.keeper, suite.ctx

	// This is normal.
	k.SetHistoricalRewards(ctx, denom1, 1, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 1000000)),
	})
	_, broken := farmingkeeper.NonNegativeHistoricalRewardsInvariant(k)(ctx)
	suite.Require().False(broken)

	// Zero-amount historical rewards
	k.SetHistoricalRewards(ctx, denom2, 1, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.DecCoins{},
	})
	_, broken = farmingkeeper.NonNegativeHistoricalRewardsInvariant(k)(ctx)
	suite.Require().False(broken)

	// Negative-amount historical rewards
	k.SetHistoricalRewards(ctx, denom2, 1, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.DecCoins{sdk.DecCoin{Denom: denom3, Amount: sdk.NewDec(-1)}},
	})
	_, broken = farmingkeeper.NonNegativeHistoricalRewardsInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestPositiveTotalStakingsAmountInvariant() {
	k, ctx := suite.keeper, suite.ctx

	// This is normal.
	k.SetTotalStakings(ctx, denom1, types.TotalStakings{Amount: sdk.NewInt(1000000)})
	_, broken := farmingkeeper.PositiveTotalStakingsAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Zero-amount total stakings.
	k.SetTotalStakings(ctx, denom1, types.TotalStakings{Amount: sdk.ZeroInt()})
	_, broken = farmingkeeper.PositiveTotalStakingsAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// Negative-amount total stakings.
	k.SetTotalStakings(ctx, denom1, types.TotalStakings{Amount: sdk.NewInt(-1)})
	_, broken = farmingkeeper.PositiveTotalStakingsAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}
