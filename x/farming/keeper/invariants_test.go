package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingkeeper "github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
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

	// Delete the zero-amount staking
	k.DeleteStaking(ctx, denom1, suite.addrs[1])
	_, broken = farmingkeeper.PositiveStakingAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Negative-amount staking
	k.SetStaking(ctx, denom1, suite.addrs[2], types.Staking{
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

	// Delete the zero-amount queued staking
	k.DeleteQueuedStaking(ctx, denom1, suite.addrs[1])
	_, broken = farmingkeeper.PositiveQueuedStakingAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Negative-amount queued staking
	k.SetQueuedStaking(ctx, denom1, suite.addrs[2], types.QueuedStaking{
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
		ctx, suite.addrs[1], k.GetStakingReservePoolAcc(ctx), sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// Send coins from staking reserve acc to another acc.
	// Staking amount in the store < balance of staking reserve acc. This shouldn't be OK.
	err = suite.app.BankKeeper.SendCoins(
		ctx, k.GetStakingReservePoolAcc(ctx), suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000001)))
	suite.Require().NoError(err)
	_, broken = farmingkeeper.StakingReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)
}

func (suite *KeeperTestSuite) TestRemainingRewardsAmountInvariant() {
	// TODO: implement
}
