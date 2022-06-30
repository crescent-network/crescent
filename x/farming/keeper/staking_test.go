package keeper_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/types"

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

				queuedStakingAmt := suite.keeper.GetAllQueuedStakingAmountByFarmerAndDenom(suite.ctx, suite.addrs[0], denom1)
				suite.Require().True(intEq(sdk.NewInt(tc.amt), queuedStakingAmt))
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

	suite.advanceEpochDays()

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
	suite.advanceEpochDays()
	suite.advanceEpochDays()

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.advanceEpochDays()
	suite.advanceEpochDays()

	suite.CreateFixedAmountPlan(suite.addrs[4], map[string]string{denom1: "1"}, map[string]int64{denom3: 1000000})
	suite.Require().True(coinsEq(sdk.NewCoins(), suite.AllRewards(suite.addrs[0])))
	suite.advanceEpochDays()
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom3, 1000000)), suite.AllRewards(suite.addrs[0])))
}

func (suite *KeeperTestSuite) TestQueuedStaking() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T09:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	// Stake more after 30 minutes.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T09:30:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	// Stake more just before the day ends.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T23:59:59Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 750000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	queuedCoins := suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 2250000)), queuedCoins))

	// The next day.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	// There shouldn't be any staking yet.
	_, found := suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().False(found)
	queuedCoins = suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 2250000)), queuedCoins))

	// Not yet...
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T08:59:59.999Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	_, found = suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().False(found)
	queuedCoins = suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 2250000)), queuedCoins))

	// The first queued staking has been staked.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T09:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	staking, found := suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1000000), staking.Amount))
	queuedCoins = suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1250000)), queuedCoins))

	// The second queued staking has been staked.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T10:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	staking, _ = suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(intEq(sdk.NewInt(1500000), staking.Amount))
	queuedCoins = suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 750000)), queuedCoins))

	// Finally, the last queued staking has been staked.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-03T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	staking, _ = suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(intEq(sdk.NewInt(2250000), staking.Amount))
	queuedCoins = suite.keeper.GetAllQueuedCoinsByFarmer(suite.ctx, suite.addrs[0])
	suite.Require().True(coinsEq(sdk.Coins{}, queuedCoins))
}

func (suite *KeeperTestSuite) TestStakeTwice() {
	// Stake twice in the same block.
	// Queued stakings are merged into one queued staking which ends
	// one day later.
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)
	amt := suite.keeper.GetAllQueuedStakingAmountByFarmerAndDenom(suite.ctx, suite.addrs[0], denom1)
	suite.Require().True(intEq(sdk.NewInt(1500000), amt))
	cnt := 0
	suite.keeper.IterateQueuedStakingsByFarmer(suite.ctx, suite.addrs[0], func(_ string, _ time.Time, _ types.QueuedStaking) (stop bool) {
		cnt++
		return false
	})
	suite.Require().Equal(1, cnt) // There should be only one queued staking object.

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	amt = suite.keeper.GetAllQueuedStakingAmountByFarmerAndDenom(suite.ctx, suite.addrs[0], denom1)
	suite.Require().True(intEq(sdk.ZeroInt(), amt))
	staking, found := suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1500000), staking.Amount))
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
			suite.advanceEpochDays()

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
	suite.advanceEpochDays()

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
	suite.advanceEpochDays()
	suite.advanceEpochDays() // Now, there are rewards to be withdrawn.

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

	suite.advanceEpochDays()
	suite.advanceEpochDays()

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 250000)))
	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 250000)))
	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, suite.addrs[0], denom3)
	suite.Require().True(intEq(balanceBefore.Amount, balanceAfter.Amount))
}

func (suite *KeeperTestSuite) TestUnstakeFromLatestQueuedStaking() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T09:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T15:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T23:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	staking, found := suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1000000), staking.Amount))
	amt := suite.keeper.GetAllQueuedStakingAmountByFarmerAndDenom(suite.ctx, suite.addrs[0], denom1)
	suite.Require().True(intEq(sdk.NewInt(2000000), amt))

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-02T03:00:00Z"))
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 700000)))
	_, found = suite.keeper.GetQueuedStaking(suite.ctx, types.ParseTime("2022-01-02T23:00:00Z"), denom1, suite.addrs[0])
	suite.Require().False(found)
	queuedStaking, found := suite.keeper.GetQueuedStaking(suite.ctx, types.ParseTime("2022-01-02T15:00:00Z"), denom1, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(300000), queuedStaking.Amount))
	queuedStaking, found = suite.keeper.GetQueuedStaking(suite.ctx, types.ParseTime("2022-01-02T09:00:00Z"), denom1, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1000000), queuedStaking.Amount))

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1500000)))
	cnt := 0
	suite.keeper.IterateQueuedStakingsByFarmer(suite.ctx, suite.addrs[0], func(_ string, _ time.Time, _ types.QueuedStaking) (stop bool) {
		cnt++
		return false
	})
	suite.Require().Equal(0, cnt)
	staking, _ = suite.keeper.GetStaking(suite.ctx, denom1, suite.addrs[0])
	suite.Require().True(intEq(sdk.NewInt(800000), staking.Amount))
}

func (suite *KeeperTestSuite) TestUnstakeInsufficientFunds() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2022-01-01T00:00:00Z"))

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	cacheCtx, _ := suite.ctx.CacheContext()
	err := suite.keeper.Unstake(cacheCtx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1100000)))
	suite.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	suite.Require().EqualError(err, "not enough staked coins, 1000000denom1 is less than 1100000denom1: insufficient funds")

	suite.advanceEpochDays()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))

	err = suite.keeper.Unstake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1600000)))
	suite.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	suite.Require().EqualError(err, "not enough staked coins, 1500000denom1 is less than 1600000denom1: insufficient funds")
}

func (suite *KeeperTestSuite) TestTotalStakings() {
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	_, found := suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().False(found)

	suite.advanceEpochDays()
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500000)))
	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 300000)))
	totalStakings, found := suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1000000), totalStakings.Amount))

	suite.advanceEpochDays()
	totalStakings, _ = suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(found)
	suite.Require().True(intEq(sdk.NewInt(1200000), totalStakings.Amount))

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	totalStakings, _ = suite.keeper.GetTotalStakings(suite.ctx, denom1)
	suite.Require().True(intEq(sdk.NewInt(200000), totalStakings.Amount))

	suite.Unstake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 200000)))
	farming.EndBlocker(suite.ctx, suite.keeper)
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

			suite.advanceEpochDays()
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

	suite.advanceEpochDays()

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
			err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, tc.farmerAcc, sdk.NewCoins(sdk.Coin{
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

func (suite *KeeperTestSuite) TestPreserveCurrentEpoch() {
	_, err := suite.createPublicFixedAmountPlan(
		suite.addrs[0], suite.addrs[0], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, utils.ParseCoins("1000000denom2"))
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-01T23:00:00Z"))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-02T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // next epoch
	suite.Require().Equal(uint64(0), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-02T23:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // queued -> staked
	suite.Require().Equal(uint64(1), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-03T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // rewards distribution
	suite.Require().Equal(uint64(2), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))
	suite.Require().True(coinsEq(utils.ParseCoins("1000000denom2"), suite.AllRewards(suite.addrs[1])))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-03T01:00:00Z"))
	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[1])
	suite.Unstake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)
	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[1])
	suite.Require().Equal(uint64(2), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))
	suite.Require().True(coinsEq(
		sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)),
		balancesAfter.Sub(balancesBefore)))

	// Few days later...
	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-04T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-05T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-06T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-07T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-07T23:00:00Z"))
	suite.Stake(suite.addrs[2], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.Require().Equal(uint64(2), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-08T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.Require().Equal(uint64(2), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-08T23:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.Require().Equal(uint64(2), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-09T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.Require().Equal(uint64(3), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))
	suite.Require().True(coinsEq(utils.ParseCoins("1000000denom2"), suite.AllRewards(suite.addrs[2])))

	balancesBefore = suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[2])
	suite.Unstake(suite.addrs[2], utils.ParseCoins("1000000denom1"))
	farming.EndBlocker(suite.ctx, suite.keeper)
	balancesAfter = suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[2])
	suite.Require().Equal(uint64(3), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))
	suite.Require().True(coinsEq(utils.ParseCoins("1000000denom1,1000000denom2"), balancesAfter.Sub(balancesBefore)))

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2022-04-10T00:00:00Z"))
	suite.Require().Equal(uint64(3), suite.keeper.GetCurrentEpoch(suite.ctx, denom1))
}

func (suite *KeeperTestSuite) TestCurrentEpoch() {
	_, err := suite.createPublicFixedAmountPlan(
		suite.addrs[4], suite.addrs[4], parseDecCoins("1denom1"),
		sampleStartTime, sampleEndTime, utils.ParseCoins("1000000denom3"))
	suite.Require().NoError(err)

	suite.executeBlock(utils.ParseTime("2022-04-01T12:00:00Z"), func() {
		suite.Stake(suite.addrs[0], utils.ParseCoins("1000000denom1"))
	})

	suite.executeBlock(utils.ParseTime("2022-04-01T14:00:00Z"), func() {
		suite.Stake(suite.addrs[1], utils.ParseCoins("1000000denom1"))
	})

	suite.executeBlock(utils.ParseTime("2022-04-01T16:00:00Z"), func() {
		suite.Stake(suite.addrs[2], utils.ParseCoins("1000000denom1"))
	})

	suite.executeBlock(utils.ParseTime("2022-04-02T00:00:00Z"), nil)

	suite.Require().Equal(uint64(0), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))
	suite.executeBlock(utils.ParseTime("2022-04-02T12:00:00Z"), nil)
	suite.Require().Equal(uint64(1), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-02T12:30:00Z"), func() {
		suite.Unstake(suite.addrs[0], utils.ParseCoins("1000000denom1"))
	})
	suite.Require().Equal(uint64(1), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-02T14:00:00Z"), nil)
	suite.Require().Equal(uint64(1), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))
	suite.executeBlock(utils.ParseTime("2022-04-02T16:00:00Z"), nil)
	suite.Require().Equal(uint64(1), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-03T00:00:00Z"), nil)
	suite.Require().True(coinsEq(utils.ParseCoins("500000denom3"), suite.AllRewards(suite.addrs[1])))
	suite.Require().True(coinsEq(utils.ParseCoins("500000denom3"), suite.AllRewards(suite.addrs[2])))

	suite.executeBlock(utils.ParseTime("2022-04-04T00:00:00Z"), nil)
	suite.executeBlock(utils.ParseTime("2022-04-05T00:00:00Z"), nil)
	suite.executeBlock(utils.ParseTime("2022-04-06T00:00:00Z"), nil)
	suite.Require().Equal(uint64(5), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-06T23:00:00Z"), func() {
		suite.Stake(suite.addrs[0], utils.ParseCoins("1000000denom1"))
	})
	suite.executeBlock(utils.ParseTime("2022-04-07T00:00:00Z"), nil)
	suite.Require().True(coinsEq(utils.ParseCoins("2500000denom3"), suite.AllRewards(suite.addrs[1])))
	suite.Require().True(coinsEq(utils.ParseCoins("2500000denom3"), suite.AllRewards(suite.addrs[2])))

	balancesBefore := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[1])
	suite.executeBlock(utils.ParseTime("2022-04-07T01:00:00Z"), func() {
		suite.Unstake(suite.addrs[1], utils.ParseCoins("1000000denom1"))
		suite.Unstake(suite.addrs[2], utils.ParseCoins("1000000denom1"))
	})
	balancesAfter := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.addrs[1])
	suite.Require().True(coinsEq(utils.ParseCoins("1000000denom1,2500000denom3"), balancesAfter.Sub(balancesBefore)))
	suite.Require().Equal(uint64(6), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-07T23:00:00Z"), nil)
	suite.Require().Equal(uint64(6), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.executeBlock(utils.ParseTime("2022-04-08T00:00:00Z"), nil)
	suite.Require().Equal(uint64(7), suite.keeper.GetCurrentEpoch(suite.ctx, "denom1"))

	suite.Require().True(coinsEq(utils.ParseCoins("1000000denom3"), suite.AllRewards(suite.addrs[0])))
}
