package keeper_test

import (
	"fmt"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesisState()

	suite.keeper.InitGenesis(suite.ctx, genState)
	got := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().Equal(genState, *got)
}

func (suite *KeeperTestSuite) TestImportExportGenesisEmpty() {
	k, ctx := suite.keeper, suite.ctx
	genState := k.ExportGenesis(ctx)

	var genState2 types.GenesisState
	bz := suite.app.AppCodec().MustMarshalJSON(genState)
	suite.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)

	genState3 := k.ExportGenesis(ctx)
	suite.Require().Equal(*genState, genState2, *genState3)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	plans := []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"name1",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1))),
				types.ParseTime("2021-07-30T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000))),
		types.NewRatioPlan(
			types.NewBasePlan(
				2,
				"name2",
				types.PlanTypePublic,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1))),
				types.ParseTime("2021-07-30T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
			),
			sdk.MustNewDecFromStr("0.01"),
		),
	}
	//for _, plan := range plans {
	//	suite.keeper.SetPlan(suite.ctx, plan)
	//}
	suite.keeper.SetPlan(suite.ctx, plans[1])
	suite.keeper.SetPlan(suite.ctx, plans[0])
	suite.keeper.SetGlobalPlanId(suite.ctx, 2)

	suite.Stake(suite.addrs[1], sdk.NewCoins(
		sdk.NewInt64Coin(denom1, 1_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000)))
	suite.advanceEpochDays()

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-07-31T00:00:00Z"))

	// Advance 2 epochs
	suite.advanceEpochDays()
	suite.advanceEpochDays()

	var genState *types.GenesisState
	suite.Require().NotPanics(func() {
		genState = suite.keeper.ExportGenesis(suite.ctx)
	})

	err := types.ValidateGenesis(*genState)
	suite.Require().NoError(err)

	suite.Require().NotPanics(func() {
		suite.keeper.InitGenesis(suite.ctx, *genState)
	})
	suite.Require().Equal(genState, suite.keeper.ExportGenesis(suite.ctx))
	suite.Require().Equal(1, suite.keeper.GetNumActivePrivatePlans(suite.ctx))
}

func (suite *KeeperTestSuite) TestInitGenesisPanics() {
	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-06T00:00:00Z"))

	for _, plan := range suite.samplePlans {
		suite.keeper.SetPlan(suite.ctx, plan)
	}
	suite.keeper.SetGlobalPlanId(suite.ctx, 4)

	err := suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	suite.Require().NoError(err)
	err = suite.keeper.Stake(suite.ctx, suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 700000), sdk.NewInt64Coin(denom2, 500000)))
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(types.Day))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(types.Day))
	farming.EndBlocker(suite.ctx, suite.keeper)

	err = suite.keeper.Stake(suite.ctx, suite.addrs[0], utils.ParseCoins("100000denom1,800000denom2"))
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(types.Day))
	farming.EndBlocker(suite.ctx, suite.keeper)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(types.Day))
	farming.EndBlocker(suite.ctx, suite.keeper)

	err = suite.keeper.Stake(suite.ctx, suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom2, 800000)))
	suite.Require().NoError(err)

	for _, tc := range []struct {
		name        string
		malleate    func(state *types.GenesisState)
		expectPanic bool
	}{
		{
			"normal",
			func(genState *types.GenesisState) {},
			false,
		},
		{
			"invalid staking records",
			func(genState *types.GenesisState) {
				genState.StakingRecords[0].Staking.Amount = sdk.NewInt(10000000)
			},
			true,
		},
		{
			"invalid queued staking records",
			func(genState *types.GenesisState) {
				genState.QueuedStakingRecords[0].QueuedStaking.Amount = sdk.NewInt(10000000)
			},
			true,
		},
		{
			"invalid remaining rewards",
			func(genState *types.GenesisState) {
				genState.HistoricalRewardsRecords[0].HistoricalRewards.CumulativeUnitRewards = genState.HistoricalRewardsRecords[0].HistoricalRewards.CumulativeUnitRewards.Add(
					sdk.NewInt64DecCoin(denom3, 1000000))
			},
			true,
		},
		{
			"invalid reward pool coins",
			func(genState *types.GenesisState) {
				genState.RewardPoolCoins = sdk.NewCoins(sdk.NewInt64Coin(denom3, 100))
			},
			true,
		},
		{
			"invalid outstanding rewards records",
			func(genState *types.GenesisState) {
				genState.OutstandingRewardsRecords[0].OutstandingRewards.Rewards = genState.OutstandingRewardsRecords[0].OutstandingRewards.Rewards.Add(
					sdk.NewInt64DecCoin(denom3, 1000000))
			},
			true,
		},
		{
			"invalid unharvested rewards records",
			func(genState *types.GenesisState) {
				genState.UnharvestedRewardsRecords[0].UnharvestedRewards.Rewards = genState.UnharvestedRewardsRecords[0].UnharvestedRewards.Rewards.Add(
					utils.ParseCoin("1000000denom3"))
			},
			true,
		},
		{
			"invalid current epoch days",
			func(genState *types.GenesisState) {
				genState.CurrentEpochDays = 0
			},
			true,
		},
	} {
		suite.Run(tc.name, func() {
			genState := suite.keeper.ExportGenesis(suite.ctx)
			tc.malleate(genState)

			cacheCtx, _ := suite.ctx.CacheContext()

			fn := suite.Require().NotPanics
			if tc.expectPanic {
				fn = suite.Require().Panics
			}
			fn(func() {
				suite.keeper.InitGenesis(cacheCtx, *genState)
			})
		})
	}
}

func (suite *KeeperTestSuite) TestMarshalUnmarshalDefaultGenesis() {
	genState := suite.keeper.ExportGenesis(suite.ctx)
	bz, err := suite.app.AppCodec().MarshalJSON(genState)
	suite.Require().NoError(err)
	var genState2 types.GenesisState
	err = suite.app.AppCodec().UnmarshalJSON(bz, &genState2)
	suite.Require().NoError(err)
	suite.Require().Equal(*genState, genState2)

	app2 := chain.Setup(false)
	ctx2 := app2.BaseApp.NewContext(false, tmproto.Header{})
	keeper2 := app2.FarmingKeeper
	keeper2.InitGenesis(ctx2, genState2)

	genState3 := keeper2.ExportGenesis(ctx2)
	suite.Require().Equal(genState2, *genState3)
}

func (suite *KeeperTestSuite) TestExportGenesis() {
	for i := range suite.sampleFixedAmtPlans {
		plan := suite.sampleFixedAmtPlans[len(suite.sampleFixedAmtPlans)-i-1]
		suite.keeper.SetPlan(suite.ctx, plan)
	}

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-04T23:00:00Z"))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 500000)))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-05T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // next epoch

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-05T23:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // queued coins => staked coins

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-06T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // allocate rewards

	suite.ctx = suite.ctx.WithBlockTime(utils.ParseTime("2021-08-06T03:00:00Z"))
	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000000), sdk.NewInt64Coin(denom2, 1500000)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 500000)))
	farming.EndBlocker(suite.ctx, suite.keeper)

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-07T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // allocated rewards

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-07T03:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // queued coins => staked coins

	suite.ctx = suite.ctx.WithBlockTime(types.ParseTime("2021-08-08T00:00:00Z"))
	farming.EndBlocker(suite.ctx, suite.keeper) // allocate rewards

	suite.Stake(suite.addrs[0], sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)))
	suite.Stake(suite.addrs[1], sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000000), sdk.NewInt64Coin(denom2, 1000000)))

	genState := suite.keeper.ExportGenesis(suite.ctx)
	bz, err := suite.app.AppCodec().MarshalJSON(genState)
	suite.Require().NoError(err)
	*genState = types.GenesisState{}
	err = suite.app.AppCodec().UnmarshalJSON(bz, genState)
	suite.Require().NoError(err)

	for _, tc := range []struct {
		name  string
		check func()
	}{
		{
			"Params",
			func() {
				err := genState.Params.Validate()
				suite.Require().NoError(err)
				suite.Require().Equal(suite.keeper.GetParams(suite.ctx), genState.Params)
			},
		},
		{
			"PlanRecords",
			func() {
				suite.Require().Len(genState.PlanRecords, len(suite.sampleFixedAmtPlans))
				for _, record := range genState.PlanRecords {
					plan, err := types.UnpackPlan(&record.Plan)
					suite.Require().NoError(err)
					err = plan.Validate()
					suite.Require().NoError(err)
				}
			},
		},
		{
			"StakingRecords",
			func() {
				suite.Require().Len(genState.StakingRecords, 4)
				for _, record := range genState.StakingRecords {
					switch record.Farmer {
					case suite.addrs[0].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(intEq(sdk.NewInt(3000000), record.Staking.Amount))
						case denom2:
							suite.Require().True(intEq(sdk.NewInt(3000000), record.Staking.Amount))
						}
					case suite.addrs[1].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(intEq(sdk.NewInt(2000000), record.Staking.Amount))
						case denom2:
							suite.Require().True(intEq(sdk.NewInt(1000000), record.Staking.Amount))
						}
					}
				}
			},
		},
		{
			"QueuedStakingRecords",
			func() {
				suite.Require().Len(genState.QueuedStakingRecords, 4)
				for _, record := range genState.QueuedStakingRecords {
					switch record.Farmer {
					case suite.addrs[0].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(intEq(sdk.NewInt(1000000), record.QueuedStaking.Amount))
						case denom2:
							suite.Require().True(intEq(sdk.NewInt(1000000), record.QueuedStaking.Amount))
						}
					case suite.addrs[1].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(intEq(sdk.NewInt(2000000), record.QueuedStaking.Amount))
						case denom2:
							suite.Require().True(intEq(sdk.NewInt(1000000), record.QueuedStaking.Amount))
						}
					}
				}
			},
		},
		{
			"TotalStakingsRecords",
			func() {
				suite.Require().Len(genState.TotalStakingsRecords, 2)
				for _, record := range genState.TotalStakingsRecords {
					switch record.StakingCoinDenom {
					case denom1:
						suite.Require().True(intEq(sdk.NewInt(5000000), record.Amount))
						suite.Require().True(coinsEq(
							sdk.NewCoins(sdk.NewInt64Coin(denom1, 8000000)),
							record.StakingReserveCoins))
					case denom2:
						suite.Require().True(intEq(sdk.NewInt(4000000), record.Amount))
						suite.Require().True(coinsEq(
							sdk.NewCoins(sdk.NewInt64Coin(denom2, 6000000)),
							record.StakingReserveCoins))
					}
				}
			},
		},
		{
			"HistoricalRewards",
			func() {
				suite.Require().Len(genState.HistoricalRewardsRecords, 8)
				for _, record := range genState.HistoricalRewardsRecords {
					suite.Require().Contains([]string{denom1, denom2}, record.StakingCoinDenom)
					switch record.Epoch {
					case 0:
						suite.Require().True(record.HistoricalRewards.CumulativeUnitRewards.IsZero())
					case 1, 2, 3:
						suite.Require().False(record.HistoricalRewards.CumulativeUnitRewards.IsZero())
					default:
						panic(fmt.Sprintf("unexpected epoch %d", record.Epoch))
					}
				}
			},
		},
		{
			"OutstandingRewards",
			func() {
				suite.Require().Len(genState.OutstandingRewardsRecords, 2)
				for _, record := range genState.OutstandingRewardsRecords {
					switch record.StakingCoinDenom {
					case denom1:
						suite.Require().True(decCoinsEq(
							sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 2300000)),
							record.OutstandingRewards.Rewards))
					case denom2:
						suite.Require().True(decCoinsEq(
							sdk.NewDecCoins(sdk.NewInt64DecCoin(denom3, 700000)),
							record.OutstandingRewards.Rewards))
					}
				}
			},
		},
		{
			"UnharvestedRewards",
			func() {
				suite.Require().Len(genState.UnharvestedRewardsRecords, 4)
				for _, record := range genState.UnharvestedRewardsRecords {
					switch record.Farmer {
					case suite.addrs[0].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(coinsEq(utils.ParseCoins("2300000denom3"), record.UnharvestedRewards.Rewards))
						case denom2:
							suite.Require().True(coinsEq(utils.ParseCoins("1050000denom3"), record.UnharvestedRewards.Rewards))
						}
					case suite.addrs[1].String():
						switch record.StakingCoinDenom {
						case denom1:
							suite.Require().True(coinsEq(utils.ParseCoins("2300000denom3"), record.UnharvestedRewards.Rewards))
						case denom2:
							suite.Require().True(coinsEq(utils.ParseCoins("350000denom3"), record.UnharvestedRewards.Rewards))
						}
					}
				}
			},
		},
		{
			"CurrentEpochRecords",
			func() {
				suite.Require().Len(genState.CurrentEpochRecords, 2)
				for _, record := range genState.CurrentEpochRecords {
					suite.Require().Equal(uint64(4), record.CurrentEpoch)
				}
			},
		},
		{
			"RewardPoolCoins",
			func() {
				suite.Require().True(coinsEq(
					sdk.NewCoins(sdk.NewInt64Coin(denom3, 3000000)),
					genState.RewardPoolCoins))
			},
		},
		{
			"LastEpochTime",
			func() {
				suite.Require().NotNil(genState.LastEpochTime)
				suite.Require().Equal(types.ParseTime("2021-08-08T00:00:00Z"), *genState.LastEpochTime)
			},
		},
		{
			"CurrentEpochDays",
			func() {
				suite.Require().Equal(uint32(1), genState.CurrentEpochDays)
			},
		},
	} {
		suite.Run(tc.name, tc.check)
	}
}
