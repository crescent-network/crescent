package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/farming/types"
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
	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))

	suite.Require().True(coinsEq(
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000000)),
		suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])))
	suite.Require().True(coinsEq(sdk.Coins{}, suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[0])))

	suite.AdvanceEpoch()

	suite.Require().True(coinsEq(
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000000)),
		suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[0])))
	suite.Require().True(coinsEq(sdk.Coins{}, suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])))

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Require().True(intEq(balanceBefore.Amount, balanceAfter.Amount))
}

func (suite *KeeperTestSuite) TestStakeInAdvance() {
	// Staking in advance must not affect the total rewards.

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.AdvanceEpoch()

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.AdvanceEpoch()

	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
	suite.Require().True(coinsEq(sdk.NewCoins(), suite.AllRewards(suite.addrs[0])))
	suite.AdvanceEpoch()
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
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

func (suite *KeeperTestSuite) TestUnstakePriority() {
	// Unstake must withdraw coins from queued staking coins first,
	// not from already staked coins.

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()

	check := func(staked, queued int64) {
		suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, staked)), suite.keeper.GetAllStakedCoinsByFarmer(suite.ctx, suite.addrs[0])))
		suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, queued)), suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])))
	}

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	check(1000000, 500000)

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 300000)))
	check(1000000, 200000)

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	check(1000000, 700000)

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 300000)))
	check(1000000, 400000)
}

func (suite *KeeperTestSuite) TestUnstakeNotAlwaysWithdraw() {
	// Unstaking from queued staking coins should not trigger
	// reward withdrawal.

	suite.CreateRatioPlan(suite.addrs[4], map[string]string{denom1: "1"}, "0.1")

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.AdvanceEpoch()
	suite.AdvanceEpoch() // Now, there are rewards to be withdrawn.

	rewards := suite.AllRewards(suite.addrs[0])

	// Stake and immediately unstake.
	// This will not affect the amount of staked coins.
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))

	rewards2 := suite.AllRewards(suite.addrs[0])

	suite.Require().True(coinsEq(rewards, rewards2))
}

func (suite *KeeperTestSuite) TestMultipleUnstake() {
	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))

	suite.AdvanceEpoch()
	suite.AdvanceEpoch()

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 250000)))
	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 250000)))
	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Require().True(intEq(balanceBefore.Amount, balanceAfter.Amount))
}

func (suite *KeeperTestSuite) TestTotalStakings() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	_, found := suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().False(found)

	suite.AdvanceEpoch()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 300000)))
	totalStakings, found := suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1000000), totalStakings.Amount))

	suite.AdvanceEpoch()
	totalStakings, _ = suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1200000), totalStakings.Amount))

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	totalStakings, _ = suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(intEq(sdk.NewInt(200000), totalStakings.Amount))

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 200000)))
	_, found = suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().False(found)
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

func (suite *KeeperTestSuite) TestReserveAndReleaseStakingCoins() {
	denom9 := "denom9"
	denom10 := "denom10"
	denom11 := "denom11"
	denom12 := "denom12"
	for _, tc := range []struct {
		name             string
		farmerAcc        sdk.AccAddress
		stakingCoins     sdk.Coins
		releaseCoins     sdk.Coins
		expectErrStaking bool
		expectErrRelease bool
	}{
		{
			"normal",
			suite.addrs[0],
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}),
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}),
			false,
			false,
		},
		{
			"bad farmer's balance",
			suite.addrs[0],
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(1_100_000_000),
			}),
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(1_100_000_000),
			}),
			true,
			true,
		},
		{
			"multi coins",
			suite.addrs[0],
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom10,
				Amount: sdk.NewInt(100000),
			}),
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom10,
				Amount: sdk.NewInt(100000),
			}),
			false,
			false,
		},
		{
			"over release",
			suite.addrs[0],
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom10,
				Amount: sdk.NewInt(100000),
			}),
			sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom10,
				Amount: sdk.NewInt(110000),
			}),
			false,
			true,
		},
		{
			"partial release",
			suite.addrs[0],
			sdk.NewCoins(sdk.Coin{
				Denom:  denom11,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom12,
				Amount: sdk.NewInt(100000),
			}),
			sdk.NewCoins(sdk.Coin{
				Denom:  denom11,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom12,
				Amount: sdk.NewInt(90000),
			}),
			false,
			false,
		},
	} {
		suite.Run(tc.name, func() {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, tc.farmerAcc, sdk.NewCoins(sdk.Coin{
				Denom:  denom9,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom10,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom11,
				Amount: sdk.NewInt(100000),
			}, sdk.Coin{
				Denom:  denom12,
				Amount: sdk.NewInt(100000),
			}))
			suite.Require().NoError(err)

			errStaking := suite.keeper.ReserveStakingCoins(suite.ctx, tc.farmerAcc, tc.stakingCoins)
			if tc.expectErrStaking {
				suite.Require().Error(errStaking)
			} else {
				suite.Require().NoError(errStaking)
				errRelease := suite.keeper.ReleaseStakingCoins(suite.ctx, tc.farmerAcc, tc.releaseCoins)
				if tc.expectErrRelease {
					suite.Require().Error(errRelease)
				} else {
					suite.Require().NoError(errRelease)
					for _, coin := range tc.stakingCoins {
						reserveAcc := types.StakingReserveAcc(coin.Denom)
						suite.Require().True(suite.app.BankKeeper.BlockedAddr(suite.ctx, reserveAcc))
						reservedBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, types.StakingReserveAcc(coin.Denom))
						suite.Require().Equal(tc.stakingCoins.Sub(tc.releaseCoins).AmountOf(coin.Denom), reservedBalance.AmountOf(coin.Denom))
					}
				}
			}
		})

	}
}
