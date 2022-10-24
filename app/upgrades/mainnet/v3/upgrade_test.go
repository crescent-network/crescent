package v3_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v3/app"
	v3 "github.com/crescent-network/crescent/v3/app/upgrades/mainnet/v3"
	"github.com/crescent-network/crescent/v3/cmd/crescentd/cmd"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm"
	farmtypes "github.com/crescent-network/crescent/v3/x/farm/types"
	"github.com/crescent-network/crescent/v3/x/farming"
	farmingkeeper "github.com/crescent-network/crescent/v3/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v3/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	marketmakertypes "github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

type UpgradeTestSuite struct {
	suite.Suite
	ctx sdk.Context
	app *chain.App
}

func (s *UpgradeTestSuite) SetupTest() {
	cmd.GetConfig()
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Height: 1,
		Time:   utils.ParseTime("2022-06-01T00:00:00Z"),
	})
	farming.EndBlocker(s.ctx, s.app.FarmingKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const testUpgradeHeight = 10

func (s *UpgradeTestSuite) TestUpgradeV3() {
	testCases := []struct {
		title   string
		before  func()
		after   func()
		expPass bool
	}{
		{
			"v3 upgrade liquidity, market maker params",
			func() {
			},
			func() {
				liquidityParams := s.app.LiquidityKeeper.GetParams(s.ctx)
				marketmakerParams := s.app.MarketMakerKeeper.GetParams(s.ctx)
				s.Require().EqualValues(liquidityParams.MaxNumMarketMakingOrderTicks, liquiditytypes.DefaultMaxNumMarketMakingOrderTicks)
				s.Require().EqualValues(marketmakerParams.IncentivePairs, []marketmakertypes.IncentivePair(nil))
				s.Require().EqualValues(marketmakerParams.DepositAmount, sdk.NewCoins(sdk.NewCoin("ucre", sdk.NewInt(1000000000))))
				s.Require().EqualValues(marketmakerParams.IncentiveBudgetAddress, "cre1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sq3qhhde")
				s.Require().EqualValues(marketmakerParams.Common, marketmakertypes.Common{
					MinOpenRatio:      sdk.MustNewDecFromStr("0.5"),
					MinOpenDepthRatio: sdk.MustNewDecFromStr("0.1"),
					MaxDowntime:       uint32(20),
					MaxTotalDowntime:  uint32(100),
					MinHours:          uint32(16),
					MinDays:           uint32(22),
				})
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			s.SetupTest()

			tc.before()

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: v3.UpgradeName, Height: testUpgradeHeight}
			err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
			s.Require().NoError(err)
			_, exists := s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
			s.Require().True(exists)

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight)
			s.Require().NotPanics(func() {
				s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{})
			})

			tc.after()
		})
	}
}

func (s *UpgradeTestSuite) TestFarmingMigration() {
	creatorAddr := utils.TestAddress(0)
	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, creatorAddr,
			utils.ParseCoins("10000_000000stake,10000_000000denom1,10000_000000denom2,10000_000000denom3")))

	// Create two pairs and pools
	_, err := s.app.LiquidityKeeper.CreatePair(
		s.ctx, liquiditytypes.NewMsgCreatePair(
			creatorAddr, "denom1", "denom2"))
	s.Require().NoError(err)
	_, err = s.app.LiquidityKeeper.CreatePair(
		s.ctx, liquiditytypes.NewMsgCreatePair(
			creatorAddr, "denom2", "denom3"))
	s.Require().NoError(err)
	_, err = s.app.LiquidityKeeper.CreatePool(
		s.ctx, liquiditytypes.NewMsgCreatePool(
			creatorAddr, 1, utils.ParseCoins("100_000000denom1,100_000000denom2")))
	s.Require().NoError(err)
	_, err = s.app.LiquidityKeeper.CreatePool(
		s.ctx, liquiditytypes.NewMsgCreatePool(
			creatorAddr, 2, utils.ParseCoins("100_000000denom2,100_000000denom3")))
	s.Require().NoError(err)

	// Create a private farming plan and a public farming plan
	var (
		planStartTime = utils.ParseTime("2022-01-01T00:00:00Z")
		planEndTime   = utils.ParseTime("2023-01-01T00:00:00Z")
	)
	planCreationFee := s.app.FarmingKeeper.GetParams(s.ctx).PrivatePlanCreationFee
	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, creatorAddr, planCreationFee))

	msg := farmingtypes.NewMsgCreateFixedAmountPlan(
		"Private Farming Plan",
		creatorAddr, utils.ParseDecCoins("0.3pool1,0.7pool2"),
		planStartTime, planEndTime, utils.ParseCoins("100_000000stake"),
	)
	farmingPoolAddr1, err := s.app.FarmingKeeper.DerivePrivatePlanFarmingPoolAcc(s.ctx, msg.Name)
	s.Require().NoError(err)
	_, err = s.app.FarmingKeeper.CreateFixedAmountPlan(s.ctx, msg, farmingPoolAddr1, creatorAddr, farmingtypes.PlanTypePrivate)
	s.Require().NoError(err)
	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, farmingPoolAddr1,
			utils.ParseCoins("10000_000000stake")))

	farmingPoolAddr2 := utils.TestAddress(1)
	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, farmingPoolAddr2,
			utils.ParseCoins("10000_000000stake")))
	proposal := farmingtypes.NewPublicPlanProposal(
		"Title", "Description", []farmingtypes.AddPlanRequest{
			farmingtypes.NewAddPlanRequest(
				"Public Farming Plan", farmingPoolAddr2.String(), farmingPoolAddr2.String(),
				utils.ParseDecCoins("0.6pool1,0.4pool2"), planStartTime, planEndTime,
				utils.ParseCoins("100_000000stake"), sdk.Dec{}),
		}, nil, nil)
	s.Require().NoError(
		farmingkeeper.HandlePublicPlanProposal(s.ctx, s.app.FarmingKeeper, proposal))

	farmerAddr1 := utils.TestAddress(100)
	farmerAddr2 := utils.TestAddress(101)
	farmerAddr3 := utils.TestAddress(102)
	for _, addr := range []sdk.AccAddress{farmerAddr1, farmerAddr2, farmerAddr3} {
		s.Require().NoError(
			chain.FundAccount(
				s.app.BankKeeper, s.ctx, addr,
				utils.ParseCoins("100_000000pool1,100_000000pool2")))
	}
	s.Require().NoError(
		s.app.FarmingKeeper.Stake(s.ctx, farmerAddr1, utils.ParseCoins("1000000pool1,500000pool2")))
	s.Require().NoError(
		s.app.FarmingKeeper.Stake(s.ctx, farmerAddr2, utils.ParseCoins("1000000pool1")))

	s.ctx = s.ctx.
		WithBlockTime(s.ctx.BlockTime().Add(farmingtypes.Day)).
		WithBlockHeight(s.ctx.BlockHeight() + 1)
	farming.EndBlocker(s.ctx, s.app.FarmingKeeper)

	s.Require().NoError(
		s.app.FarmingKeeper.Stake(s.ctx, farmerAddr1, utils.ParseCoins("500000pool2")))

	s.ctx = s.ctx.
		WithBlockTime(s.ctx.BlockTime().Add(farmingtypes.Day)).
		WithBlockHeight(s.ctx.BlockHeight() + 1)
	farming.EndBlocker(s.ctx, s.app.FarmingKeeper)

	s.Require().NoError(
		s.app.FarmingKeeper.Stake(s.ctx, farmerAddr2, utils.ParseCoins("1000000pool2")))
	s.Require().NoError(
		s.app.FarmingKeeper.Stake(s.ctx, farmerAddr3, utils.ParseCoins("1000000pool1,1000000pool2")))

	// Execute the upgrade.
	s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight - 1)
	upgradePlan := upgradetypes.Plan{Name: v3.UpgradeName, Height: testUpgradeHeight}
	s.Require().NoError(s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, upgradePlan))
	s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight)
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{})

	// Check if rewards has been withdrawn.
	s.Require().Equal(
		"310000000stake", s.app.BankKeeper.GetBalance(s.ctx, farmerAddr1, "stake").String())
	s.Require().Equal(
		"90000000stake", s.app.BankKeeper.GetBalance(s.ctx, farmerAddr2, "stake").String())
	s.Require().Equal(
		"0stake", s.app.BankKeeper.GetBalance(s.ctx, farmerAddr3, "stake").String())

	// Check if keys have been deleted.
	s.Require().Zero(s.app.FarmingKeeper.GetGlobalPlanId(s.ctx))
	_, found := s.app.FarmingKeeper.GetLastEpochTime(s.ctx)
	s.Require().False(found)
	s.app.FarmingKeeper.IterateCurrentEpochs(s.ctx, func(string, uint64) (stop bool) {
		s.Require().FailNow("current epoch must not exist")
		return false
	})
	s.app.FarmingKeeper.IteratePlans(s.ctx, func(farmingtypes.PlanI) (stop bool) {
		s.Require().FailNow("plan must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateTotalStakings(s.ctx, func(string, farmingtypes.TotalStakings) (stop bool) {
		s.Require().FailNow("total staking must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateStakings(s.ctx, func(string, sdk.AccAddress, farmingtypes.Staking) (stop bool) {
		s.Require().FailNow("staking must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateQueuedStakings(s.ctx, func(time.Time, string, sdk.AccAddress, farmingtypes.QueuedStaking) (stop bool) {
		s.Require().FailNow("queued staking must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateAllUnharvestedRewards(s.ctx, func(sdk.AccAddress, string, farmingtypes.UnharvestedRewards) (stop bool) {
		s.Require().FailNow("unharvested rewards must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateHistoricalRewards(s.ctx, func(string, uint64, farmingtypes.HistoricalRewards) (stop bool) {
		s.Require().FailNow("historical rewards must not exist")
		return false
	})
	s.app.FarmingKeeper.IterateOutstandingRewards(s.ctx, func(string, farmingtypes.OutstandingRewards) (stop bool) {
		s.Require().FailNow("outstanding rewards must not exist")
		return false
	})

	s.Require().EqualValues(1, s.app.FarmKeeper.GetNumPrivatePlans(s.ctx))
	lastPlanId, found := s.app.FarmKeeper.GetLastPlanId(s.ctx)
	s.Require().True(found)
	s.Require().EqualValues(2, lastPlanId)

	// Check the private plan.
	plan, found := s.app.FarmKeeper.GetPlan(s.ctx, 1)
	s.Require().True(found)
	s.Require().Equal("Private Farming Plan", plan.Description)
	s.Require().EqualValues(farmingPoolAddr1.String(), plan.FarmingPoolAddress)
	s.Require().EqualValues(creatorAddr.String(), plan.TerminationAddress)
	s.Require().EqualValues(
		[]farmtypes.RewardAllocation{
			farmtypes.NewDenomRewardAllocation("pool1", utils.ParseCoins("30_000000stake")),
			farmtypes.NewDenomRewardAllocation("pool2", utils.ParseCoins("70_000000stake")),
		},
		plan.RewardAllocations,
	)
	s.Require().Equal(planStartTime, plan.StartTime)
	s.Require().Equal(planEndTime, plan.EndTime)
	s.Require().True(plan.IsPrivate)
	s.Require().False(plan.IsTerminated)

	// Check the public plan.
	plan, found = s.app.FarmKeeper.GetPlan(s.ctx, 2)
	s.Require().True(found)
	s.Require().Equal("Public Farming Plan", plan.Description)
	s.Require().EqualValues(farmingPoolAddr2.String(), plan.FarmingPoolAddress)
	s.Require().EqualValues(farmingPoolAddr2.String(), plan.TerminationAddress)
	s.Require().EqualValues(
		[]farmtypes.RewardAllocation{
			farmtypes.NewDenomRewardAllocation("pool1", utils.ParseCoins("60_000000stake")),
			farmtypes.NewDenomRewardAllocation("pool2", utils.ParseCoins("40_000000stake")),
		},
		plan.RewardAllocations,
	)
	s.Require().Equal(planStartTime, plan.StartTime)
	s.Require().Equal(planEndTime, plan.EndTime)
	s.Require().False(plan.IsPrivate)
	s.Require().False(plan.IsTerminated)

	// Check farms.
	f, found := s.app.FarmKeeper.GetFarm(s.ctx, "pool1")
	s.Require().True(found)
	s.Require().Equal("3000000", f.TotalFarmingAmount.String())
	s.Require().Equal("", f.CurrentRewards.String())
	s.Require().Equal("", f.OutstandingRewards.String())
	f, found = s.app.FarmKeeper.GetFarm(s.ctx, "pool2")
	s.Require().True(found)
	s.Require().Equal("3000000", f.TotalFarmingAmount.String())
	s.Require().Equal("", f.CurrentRewards.String())
	s.Require().Equal("", f.OutstandingRewards.String())

	// Check positions.
	position, found := s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr1, "pool1")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())
	position, found = s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr1, "pool2")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())
	position, found = s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr2, "pool1")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())
	position, found = s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr2, "pool2")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())
	position, found = s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr3, "pool1")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())
	position, found = s.app.FarmKeeper.GetPosition(s.ctx, farmerAddr3, "pool2")
	s.Require().True(found)
	s.Require().Equal("1000000", position.FarmingAmount.String())

	farm.BeginBlocker(s.ctx, s.app.FarmKeeper)
	s.ctx = s.ctx.
		WithBlockTime(s.ctx.BlockTime().Add(5 * time.Second)).
		WithBlockHeight(s.ctx.BlockHeight() + 1)
	farm.BeginBlocker(s.ctx, s.app.FarmKeeper)

	// Now all farmers receives same rewards per block.
	for _, addr := range []sdk.AccAddress{farmerAddr1, farmerAddr2, farmerAddr3} {
		s.Require().Equal(
			"1736.000000000000000000stake",
			s.app.FarmKeeper.Rewards(s.ctx, addr, "pool1").String())
		s.Require().Equal(
			"2121.333333333333000000stake",
			s.app.FarmKeeper.Rewards(s.ctx, addr, "pool2").String())
	}
}
