package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestStake() {
	for _, tc := range []struct {
		name      string
		amt       int64
		expectErr bool
	}{
		{
			"normal",
			1000,
			false,
		},
		{
			"more than balance",
			1_000_000_000_000,
			true,
		},
	} {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			err := suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, tc.amt)))
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				_, found := suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
				suite.Require().False(found)

				queuedStaking, found := suite.keeper.GetQueuedStaking(suite.ctx, denom1, suite.addrs[0])
				suite.Require().True(found, "queued staking should be present")
				suite.Require().True(intEq(sdk.NewInt(tc.amt), queuedStaking.Amount))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMultipleStake() {
	// TODO: implement
}

func (suite *KeeperTestSuite) TestUnstake() {
	for _, tc := range []struct {
		name            string
		addrIdx         int
		amt             sdk.Coins
		remainingStaked sdk.Coins
		remainingQueued sdk.Coins
		expectErr       bool
	}{
		{
			"from queued coins",
			0,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 500_000), sdk.NewInt64Coin(denom2, 1_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 495_000)),
			false,
		},
		{
			"from staked coins",
			0,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 700_000), sdk.NewInt64Coin(denom2, 100_000)),
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 300_000), sdk.NewInt64Coin(denom2, 900_000)),
			sdk.NewCoins(),
			false,
		},
		{
			"one coin",
			0,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(denom2, 1_000_000)),
			sdk.NewCoins(),
			false,
		},
		{
			"unstake all",
			0,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000), sdk.NewInt64Coin(denom2, 1_000_000)),
			sdk.NewCoins(),
			sdk.NewCoins(),
			false,
		},
		{
			"more than staked",
			0,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_100_000), sdk.NewInt64Coin(denom2, 1_100_000)),
			// We can use nil since there will be an error, and we don't use these fields
			nil,
			nil,
			true,
		},
		{
			"no staking",
			1,
			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
			nil,
			nil,
			true,
		},
	} {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.Stake(suite.addrs[0], sdk.NewCoins(
				sdk.NewInt64Coin(denom1, 500_000),
				sdk.NewInt64Coin(denom2, 1_000_000)))

			// Make queued coins be staked.
			suite.keeper.ProcessQueuedCoins(suite.ctx)

			suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500_000)))

			// At this moment, we have 500000denom1,1000000denom2 staked and
			// 500000denom1 queued.

			err := suite.keeper.Unstake(suite.ctx, suite.addrs[tc.addrIdx], tc.amt)
			if tc.expectErr {
				suite.Error(err)
			} else {
				if suite.NoError(err) {
					suite.True(coinsEq(tc.remainingStaked, suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[tc.addrIdx])))
					suite.True(coinsEq(tc.remainingQueued, suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[tc.addrIdx])))
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMultipleUnstake() {
	// TODO: implement
}

func (suite *KeeperTestSuite) TestTotalStaking() {
	// TODO: implement
}

func (suite *KeeperTestSuite) TestProcessQueuedCoins() {
	for seed := int64(0); seed < 10; seed++ {
		suite.SetupTest()

		r := rand.New(rand.NewSource(seed))

		stakedCoins := sdk.NewCoins()
		queuedCoins := sdk.NewCoins()

		iterations := 100
		for i := 0; i < iterations; i++ {
			if r.Intn(2) == 0 { // Stake with a 50% chance
				// Construct random amount of coins to stake
				stakingCoins := sdk.NewCoins()
				for _, denom := range []string{denom1, denom2} {
					balance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom)
					amt := r.Int63n(balance.Amount.ToDec().QuoTruncate(sdk.NewDec(int64(iterations))).TruncateInt64())
					stakingCoins = stakingCoins.Add(sdk.NewInt64Coin(denom, amt))
				}

				if !stakingCoins.IsZero() {
					suite.Stake(suite.addrs[0], stakingCoins)
					queuedCoins = queuedCoins.Add(stakingCoins...)
				}
			}

			suite.Require().True(coinsEq(queuedCoins, suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])))
			suite.Require().True(coinsEq(stakedCoins, suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[0])))

			suite.keeper.ProcessQueuedCoins(suite.ctx)
			stakedCoins = stakedCoins.Add(queuedCoins...)
			queuedCoins = sdk.NewCoins()

			suite.Require().True(coinsEq(queuedCoins, suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])))
			suite.Require().True(coinsEq(stakedCoins, suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[0])))
		}
	}
}

func (suite *KeeperTestSuite) TestDelayedStakingGasFee() {
	suite.ctx = suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	err := suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000000)))
	suite.Require().NoError(err)
	gasConsumedNormal := suite.ctx.GasMeter().GasConsumed()

	suite.AdvanceEpoch()

	suite.ctx = suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	err = suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000000)))
	suite.Require().NoError(err)
	gasConsumedWithStaking := suite.ctx.GasMeter().GasConsumed()

	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().GreaterOrEqual(gasConsumedWithStaking, params.DelayedStakingGasFee)
	suite.Require().Greater(gasConsumedWithStaking, gasConsumedNormal)
}
