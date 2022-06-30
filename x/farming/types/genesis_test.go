package types_test

import (
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func TestValidateGenesis(t *testing.T) {
	validAcc := sdk.AccAddress(crypto.AddressHash([]byte("validAcc")))
	validStakingCoinDenom := "denom1"
	validPlan := types.NewRatioPlan(
		types.NewBasePlan(
			1,
			"planA",
			types.PlanTypePublic,
			validAcc.String(),
			validAcc.String(),
			sdk.NewDecCoins(
				sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
			),
			types.ParseTime("0001-01-01T00:00:00Z"),
			types.ParseTime("9999-12-31T00:00:00Z"),
		),
		sdk.NewDecWithPrec(5, 2),
	)
	validStaking := types.Staking{
		Amount:        sdk.NewInt(1000000),
		StartingEpoch: 1,
	}
	validQueuedStaking := types.QueuedStaking{
		Amount: sdk.NewInt(1000000),
	}
	validHistoricalRewards := types.HistoricalRewards{
		CumulativeUnitRewards: sdk.NewDecCoins(sdk.NewInt64DecCoin("denom3", 100000)),
	}
	validOutstandingRewards := types.OutstandingRewards{
		Rewards: sdk.NewDecCoins(sdk.NewInt64DecCoin("denom3", 1000000)),
	}
	validUnharvestedRewards := types.UnharvestedRewards{
		Rewards: utils.ParseCoins("1000000denom3"),
	}

	testCases := []struct {
		name        string
		configure   func(*types.GenesisState)
		expectedErr string
	}{
		{
			"default case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				genState.Params = params
			},
			"",
		},
		{
			"invalid NextEpochDays case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.NextEpochDays = 0
				genState.Params = params
			},
			"next epoch days must be positive: 0",
		},
		{
			"invalid plan",
			func(genState *types.GenesisState) {
				plan := types.NewRatioPlan(
					types.NewBasePlan(
						1,
						"planA",
						types.PlanTypeNil,
						validAcc.String(),
						validAcc.String(),
						sdk.NewDecCoins(
							sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
						),
						types.ParseTime("0001-01-01T00:00:00Z"),
						types.ParseTime("9999-12-31T00:00:00Z"),
					),
					sdk.NewDecWithPrec(5, 2),
				)
				planAny, _ := types.PackPlan(plan)
				genState.PlanRecords = []types.PlanRecord{
					{
						Plan:             *planAny,
						FarmingPoolCoins: sdk.NewCoins(),
					},
				}
				genState.GlobalPlanId = 1
			},
			"unknown plan type: PLAN_TYPE_UNSPECIFIED: invalid plan type",
		},
		{
			"invalid plan records - empty type url",
			func(genState *types.GenesisState) {
				genState.PlanRecords = []types.PlanRecord{
					{
						Plan:             cdctypes.Any{},
						FarmingPoolCoins: sdk.NewCoins(),
					},
				}
			},
			"empty type url: invalid type",
		},
		{
			"invalid plan records - invalid farming pool coins",
			func(genState *types.GenesisState) {
				planAny, _ := types.PackPlan(validPlan)
				genState.PlanRecords = []types.PlanRecord{
					{
						Plan:             *planAny,
						FarmingPoolCoins: sdk.Coins{sdk.NewInt64Coin("denom3", 0)},
					},
				}
			},
			"coin 0denom3 amount is not positive",
		},
		{
			"plan id greater than the global last plan id",
			func(genState *types.GenesisState) {
				planA := types.NewRatioPlan(
					types.NewBasePlan(
						1,
						"planA",
						types.PlanTypePublic,
						validAcc.String(),
						validAcc.String(),
						sdk.NewDecCoins(
							sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
						),
						types.ParseTime("0001-01-01T00:00:00Z"),
						types.ParseTime("9999-12-31T00:00:00Z"),
					),
					sdk.NewDecWithPrec(5, 2),
				)
				planB := types.NewFixedAmountPlan(
					types.NewBasePlan(
						3,
						"planB",
						types.PlanTypePublic,
						validAcc.String(),
						validAcc.String(),
						sdk.NewDecCoins(
							sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
						),
						types.ParseTime("0001-01-01T00:00:00Z"),
						types.ParseTime("9999-12-31T00:00:00Z"),
					),
					sdk.NewCoins(sdk.NewInt64Coin("denom3", 1000000)),
				)
				planAAny, _ := types.PackPlan(planA)
				planBAny, _ := types.PackPlan(planB)
				genState.PlanRecords = []types.PlanRecord{
					{
						Plan:             *planAAny,
						FarmingPoolCoins: sdk.NewCoins(),
					},
					{
						Plan:             *planBAny,
						FarmingPoolCoins: sdk.NewCoins(),
					},
				}
				genState.GlobalPlanId = 2
			},
			"plan id is greater than the global last plan id",
		},
		{
			"invalid plan records - invalid sum of epoch ratio",
			func(genState *types.GenesisState) {
				planA := types.NewRatioPlan(
					types.NewBasePlan(
						1,
						"planA",
						types.PlanTypePublic,
						validAcc.String(),
						validAcc.String(),
						sdk.NewDecCoins(
							sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
						),
						types.ParseTime("0001-01-01T00:00:00Z"),
						types.ParseTime("9999-12-31T00:00:00Z"),
					),
					sdk.OneDec(),
				)
				planB := types.NewRatioPlan(
					types.NewBasePlan(
						2,
						"planB",
						types.PlanTypePublic,
						validAcc.String(),
						validAcc.String(),
						sdk.NewDecCoins(
							sdk.NewInt64DecCoin(validStakingCoinDenom, 1),
						),
						types.ParseTime("0001-01-01T00:00:00Z"),
						types.ParseTime("9999-12-31T00:00:00Z"),
					),
					sdk.OneDec(),
				)
				planAAny, _ := types.PackPlan(planA)
				planBAny, _ := types.PackPlan(planB)
				genState.PlanRecords = []types.PlanRecord{
					{
						Plan:             *planAAny,
						FarmingPoolCoins: sdk.NewCoins(),
					},
					{
						Plan:             *planBAny,
						FarmingPoolCoins: sdk.NewCoins(),
					},
				}
				genState.GlobalPlanId = 2
			},
			"total epoch ratio must be lower than 1: invalid total epoch ratio",
		},
		{
			"invalid staking records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.StakingRecords = []types.StakingRecord{
					{
						StakingCoinDenom: "!",
						Farmer:           validAcc.String(),
						Staking:          validStaking,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid staking records - invalid farmer addr",
			func(genState *types.GenesisState) {
				genState.StakingRecords = []types.StakingRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						Farmer:           "invalid",
						Staking:          validStaking,
					},
				}
			},
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid staking records - invalid staking amount",
			func(genState *types.GenesisState) {
				genState.StakingRecords = []types.StakingRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						Farmer:           validAcc.String(),
						Staking: types.Staking{
							Amount:        sdk.ZeroInt(),
							StartingEpoch: 0,
						},
					},
				}
			},
			"staking amount must be positive: 0",
		},
		{
			"invalid queued staking records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.QueuedStakingRecords = []types.QueuedStakingRecord{
					{
						StakingCoinDenom: "!",
						Farmer:           validAcc.String(),
						QueuedStaking:    validQueuedStaking,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid queued staking records - invalid farmer addr",
			func(genState *types.GenesisState) {
				genState.QueuedStakingRecords = []types.QueuedStakingRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						Farmer:           "invalid",
						QueuedStaking:    validQueuedStaking,
					},
				}
			},
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid queued staking records - invalid queued staking amount",
			func(genState *types.GenesisState) {
				genState.QueuedStakingRecords = []types.QueuedStakingRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						Farmer:           validAcc.String(),
						QueuedStaking: types.QueuedStaking{
							Amount: sdk.ZeroInt(),
						},
					},
				}
			},
			"queued staking amount must be positive: 0",
		},
		{
			"invalid historical rewards records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.HistoricalRewardsRecords = []types.HistoricalRewardsRecord{
					{
						StakingCoinDenom:  "!",
						Epoch:             0,
						HistoricalRewards: validHistoricalRewards,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid total staking records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.TotalStakingsRecords = []types.TotalStakingsRecord{
					{
						StakingCoinDenom: "!",
						Amount:           sdk.OneInt(),
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid total staking records - invalid staking amount",
			func(genState *types.GenesisState) {
				genState.TotalStakingsRecords = []types.TotalStakingsRecord{
					{
						StakingCoinDenom: "uatom",
						Amount:           sdk.ZeroInt(),
					},
				}
			},
			"total staking amount must be positive: 0",
		},
		{
			"invalid historical rewards records - invalid historical rewards",
			func(genState *types.GenesisState) {
				genState.HistoricalRewardsRecords = []types.HistoricalRewardsRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						Epoch:            0,
						HistoricalRewards: types.HistoricalRewards{
							CumulativeUnitRewards: sdk.DecCoins{sdk.NewInt64DecCoin("denom3", 0)},
						},
					},
				}
			},
			"coin 0.000000000000000000denom3 amount is not positive",
		},
		{
			"invalid outstanding rewards records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.OutstandingRewardsRecords = []types.OutstandingRewardsRecord{
					{
						StakingCoinDenom:   "!",
						OutstandingRewards: validOutstandingRewards,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid outstanding rewards records - invalid outstanding rewards",
			func(genState *types.GenesisState) {
				genState.OutstandingRewardsRecords = []types.OutstandingRewardsRecord{
					{
						StakingCoinDenom: validStakingCoinDenom,
						OutstandingRewards: types.OutstandingRewards{
							Rewards: sdk.DecCoins{sdk.NewInt64DecCoin("denom3", 0)},
						},
					},
				}
			},
			"coin 0.000000000000000000denom3 amount is not positive",
		},
		{
			"invalid unharvested rewards records - invalid farmer address",
			func(genState *types.GenesisState) {
				genState.UnharvestedRewardsRecords = []types.UnharvestedRewardsRecord{
					{
						Farmer:             "invalid",
						StakingCoinDenom:   validStakingCoinDenom,
						UnharvestedRewards: validUnharvestedRewards,
					},
				}
			},
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid unharvested rewards records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.UnharvestedRewardsRecords = []types.UnharvestedRewardsRecord{
					{
						Farmer:             validAcc.String(),
						StakingCoinDenom:   "!",
						UnharvestedRewards: validUnharvestedRewards,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid unharvested rewards records - invalid rewards",
			func(genState *types.GenesisState) {
				genState.UnharvestedRewardsRecords = []types.UnharvestedRewardsRecord{
					{
						Farmer:           validAcc.String(),
						StakingCoinDenom: validStakingCoinDenom,
						UnharvestedRewards: types.UnharvestedRewards{
							Rewards: sdk.Coins{sdk.Coin{Denom: "denom3", Amount: sdk.ZeroInt()}},
						},
					},
				}
			},
			"coin 0denom3 amount is not positive",
		},
		{
			"invalid current epoch records - invalid staking coin denom",
			func(genState *types.GenesisState) {
				genState.CurrentEpochRecords = []types.CurrentEpochRecord{
					{
						StakingCoinDenom: "!",
						CurrentEpoch:     0,
					},
				}
			},
			"invalid denom: !",
		},
		{
			"invalid reward pool coins",
			func(genState *types.GenesisState) {
				genState.RewardPoolCoins = sdk.Coins{sdk.NewInt64Coin(validStakingCoinDenom, 0)}
			},
			"coin 0denom1 amount is not positive",
		},
		{
			"invalid current epoch days",
			func(genState *types.GenesisState) {
				genState.CurrentEpochDays = 0
			},
			"current epoch days must be positive",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			tc.configure(genState)

			err := types.ValidateGenesis(*genState)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
