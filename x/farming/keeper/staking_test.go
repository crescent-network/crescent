package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//func (suite *KeeperTestSuite) TestGetStaking() {
//	_, found := suite.keeper.GetStaking(suite.ctx, 1)
//	suite.False(found, "staking should not be present")
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
//
//	_, found = suite.keeper.GetStaking(suite.ctx, 1)
//	suite.True(found, "staking should be present")
//}
//
//func (suite *KeeperTestSuite) TestStake() {
//	for _, tc := range []struct {
//		name            string
//		amt             int64
//		remainingStaked int64
//		remainingQueued int64
//		expectErr       bool
//	}{
//		{
//			"normal",
//			1000,
//			0,
//			1000,
//			false,
//		},
//		{
//			"more than balance",
//			10_000_000_000,
//			0,
//			0,
//			true,
//		},
//	} {
//		suite.Run(tc.name, func() {
//			suite.SetupTest()
//
//			_, found := suite.keeper.GetStakingByFarmer(suite.ctx, suite.addrs[0])
//			suite.Require().False(found, "staking should not be present")
//
//			staking, err := suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, tc.amt)))
//			if tc.expectErr {
//				suite.Error(err)
//			} else {
//				suite.NoError(err)
//				staking2, found := suite.keeper.GetStakingByFarmer(suite.ctx, suite.addrs[0])
//				suite.True(found, "staking should be present")
//				suite.True(staking2.StakedCoins.IsEqual(staking.StakedCoins))
//				suite.True(staking2.QueuedCoins.IsEqual(staking2.QueuedCoins))
//
//				suite.True(intEq(sdk.NewInt(tc.remainingStaked), staking.StakedCoins.AmountOf(denom1)))
//				suite.True(intEq(sdk.NewInt(tc.remainingQueued), staking.QueuedCoins.AmountOf(denom1)))
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestUnstake() {
//	for _, tc := range []struct {
//		name            string
//		addrIdx         int
//		amt             sdk.Coins
//		remainingStaked sdk.Coins
//		remainingQueued sdk.Coins
//		expectErr       bool
//	}{
//		{
//			"from queued coins",
//			0,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 500_000), sdk.NewInt64Coin(denom2, 1_000_000)),
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 495_000)),
//			false,
//		},
//		{
//			"from staked coins",
//			0,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 700_000), sdk.NewInt64Coin(denom2, 100_000)),
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 300_000), sdk.NewInt64Coin(denom2, 900_000)),
//			sdk.NewCoins(),
//			false,
//		},
//		{
//			"one coin",
//			0,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000)),
//			sdk.NewCoins(sdk.NewInt64Coin(denom2, 1_000_000)),
//			sdk.NewCoins(),
//			false,
//		},
//		{
//			"unstake all",
//			0,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000), sdk.NewInt64Coin(denom2, 1_000_000)),
//			sdk.NewCoins(),
//			sdk.NewCoins(),
//			false,
//		},
//		{
//			"more than staked",
//			0,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_100_000), sdk.NewInt64Coin(denom2, 1_100_000)),
//			// We can use nil since there will be an error and we don't use these fields
//			nil,
//			nil,
//			true,
//		},
//		{
//			"no staking",
//			1,
//			sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
//			nil,
//			nil,
//			true,
//		},
//	} {
//		suite.Run(tc.name, func() {
//			suite.SetupTest()
//			suite.Stake(suite.addrs[0], sdk.NewCoins(
//				sdk.NewInt64Coin(denom1, 500_000),
//				sdk.NewInt64Coin(denom2, 1_000_000)))
//
//			// Make queued coins be staked.
//			suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-07-23T05:00:00Z"))
//			farming.EndBlocker(suite.ctx, suite.keeper)
//			suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-07-24T00:05:00Z"))
//			farming.EndBlocker(suite.ctx, suite.keeper)
//
//			suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500_000)))
//
//			// At this moment, we have 500,000denom1,1,000,000denom2 staked and
//			// 500000denom1 queued.
//
//			staking, err := suite.keeper.Unstake(suite.ctx, suite.addrs[tc.addrIdx], tc.amt)
//			if tc.expectErr {
//				suite.Error(err)
//			} else {
//				if suite.NoError(err) {
//					suite.True(coinsEq(tc.remainingStaked, staking.StakedCoins))
//					suite.True(coinsEq(tc.remainingQueued, staking.QueuedCoins))
//				}
//			}
//		})
//	}
//}
//

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

			suite.Require().True(coinsEq(queuedCoins, suite.QueuedCoins(suite.addrs[0])))
			suite.Require().True(coinsEq(stakedCoins, suite.StakedCoins(suite.addrs[0])))

			suite.keeper.ProcessQueuedCoins(suite.ctx)
			stakedCoins = stakedCoins.Add(queuedCoins...)
			queuedCoins = sdk.NewCoins()

			suite.Require().True(coinsEq(queuedCoins, suite.QueuedCoins(suite.addrs[0])))
			suite.Require().True(coinsEq(stakedCoins, suite.StakedCoins(suite.addrs[0])))
		}
	}
}

//
//func (suite *KeeperTestSuite) TestEndBlockerProcessQueuedCoins() {
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)))
//
//	ctx := suite.ctx.WithBlockTime(mustParseRFC3339("2021-07-23T05:00:00Z"))
//	farming.EndBlocker(ctx, suite.keeper)
//
//	staking, _ := suite.keeper.GetStakingByFarmer(ctx, suite.addrs[0])
//	suite.Require().True(intEq(sdk.NewInt(1000), staking.QueuedCoins.AmountOf(denom1)))
//	suite.Require().True(staking.StakedCoins.IsZero(), "staked coins must be empty")
//
//	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)))
//
//	ctx = ctx.WithBlockTime(mustParseRFC3339("2021-07-23T23:59:59Z"))
//	farming.EndBlocker(ctx, suite.keeper)
//
//	staking, _ = suite.keeper.GetStakingByFarmer(ctx, suite.addrs[0])
//	suite.Require().True(intEq(sdk.NewInt(1500), staking.QueuedCoins.AmountOf(denom1)))
//	suite.Require().True(staking.StakedCoins.IsZero(), "staked coins must be empty")
//
//	ctx = ctx.WithBlockTime(mustParseRFC3339("2021-07-24T00:00:01Z"))
//	farming.EndBlocker(ctx, suite.keeper)
//
//	staking, _ = suite.keeper.GetStakingByFarmer(ctx, suite.addrs[0])
//	suite.Require().True(staking.QueuedCoins.IsZero(), "queued coins must be empty")
//	suite.Require().True(intEq(sdk.NewInt(1500), staking.StakedCoins.AmountOf(denom1)))
//}
