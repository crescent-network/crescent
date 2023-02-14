package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

func TestGenesisState_Validate(t *testing.T) {
	// Valid structs.
	validPlan := types.NewPlan(
		1, "Farming Plan", utils.TestAddress(0), utils.TestAddress(1),
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
			types.NewPairRewardAllocation(2, utils.ParseCoins("200_000000stake")),
		},
		utils.ParseTime("2022-01-01T00:00:00Z"),
		utils.ParseTime("2023-01-01T00:00:00Z"), true)
	validFarm := types.Farm{
		TotalFarmingAmount: sdk.NewInt(100_000000),
		CurrentRewards:     utils.ParseDecCoins("10000stake"),
		OutstandingRewards: utils.ParseDecCoins("20000stake"),
		Period:             1,
	}
	validPosition := types.Position{
		Farmer:              utils.TestAddress(2).String(),
		Denom:               "pool1",
		FarmingAmount:       sdk.NewInt(100_000000),
		PreviousPeriod:      3,
		StartingBlockHeight: 10,
	}
	validHist := types.HistoricalRewards{
		CumulativeUnitRewards: utils.ParseDecCoins("5.5stake"),
		ReferenceCount:        1,
	}

	for _, tc := range []struct {
		name        string
		malleate    func(genState *types.GenesisState)
		expectedErr string
	}{
		{
			"default is valid",
			func(genState *types.GenesisState) {},
			"",
		},
		{
			"invalid params",
			func(genState *types.GenesisState) {
				genState.Params.FeeCollector = "invalidaddr"
			},
			"invalid fee collector address: invalidaddr",
		},
		{
			"duplicate plan",
			func(genState *types.GenesisState) {
				genState.Plans = []types.Plan{
					validPlan, validPlan,
				}
			},
			"duplicate plan: 1",
		},
		{
			"invalid farm: invalid denom",
			func(genState *types.GenesisState) {
				genState.Farms = []types.FarmRecord{{Denom: "invalid!", Farm: validFarm}}
			},
			"invalid farm denom: invalid denom: invalid!",
		},
		{
			"total farming amount can be zero",
			func(genState *types.GenesisState) {
				farm := validFarm
				farm.TotalFarmingAmount = sdk.ZeroInt()
				genState.Farms = []types.FarmRecord{{Denom: "pool1", Farm: farm}}
			},
			"",
		},
		{
			"invalid farm: negative total farming amount",
			func(genState *types.GenesisState) {
				farm := validFarm
				farm.TotalFarmingAmount = sdk.NewInt(-1)
				genState.Farms = []types.FarmRecord{{Denom: "pool1", Farm: farm}}
			},
			"total farming amount must not be negative: -1",
		},
		{
			"invalid farm: invalid current rewards",
			func(genState *types.GenesisState) {
				farm := validFarm
				farm.CurrentRewards = sdk.DecCoins{utils.ParseDecCoin("0stake")}
				genState.Farms = []types.FarmRecord{{Denom: "pool1", Farm: farm}}
			},
			"invalid current rewards: coin 0.000000000000000000stake amount is not positive",
		},
		{
			"invalid farm: invalid outstanding rewards",
			func(genState *types.GenesisState) {
				farm := validFarm
				farm.OutstandingRewards = sdk.DecCoins{utils.ParseDecCoin("0stake")}
				genState.Farms = []types.FarmRecord{{Denom: "pool1", Farm: farm}}
			},
			"invalid outstanding rewards: coin 0.000000000000000000stake amount is not positive",
		},
		{
			"invalid farm: invalid period",
			func(genState *types.GenesisState) {
				farm := validFarm
				farm.Period = 0
				genState.Farms = []types.FarmRecord{{Denom: "pool1", Farm: farm}}
			},
			"period must be positive",
		},
		{
			"duplicate farm",
			func(genState *types.GenesisState) {
				genState.Farms = []types.FarmRecord{
					{
						Denom: "pool1",
						Farm:  validFarm,
					},
					{
						Denom: "pool1",
						Farm:  validFarm,
					},
				}
			},
			"duplicate farm: pool1",
		},
		{
			"invalid position: invalid farmer",
			func(genState *types.GenesisState) {
				position := validPosition
				position.Farmer = "invalidaddr"
				genState.Positions = []types.Position{position}
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid position: invalid denom",
			func(genState *types.GenesisState) {
				position := validPosition
				position.Denom = "invalid!"
				genState.Positions = []types.Position{position}
			},
			"invalid position denom: invalid denom: invalid!",
		},
		{
			"invalid position: invalid farming amount",
			func(genState *types.GenesisState) {
				position := validPosition
				position.FarmingAmount = sdk.ZeroInt()
				genState.Positions = []types.Position{position}
			},
			"farming amount must be positive: 0",
		},
		{
			"invalid position: invalid starting block height",
			func(genState *types.GenesisState) {
				position := validPosition
				position.StartingBlockHeight = 0
				genState.Positions = []types.Position{position}
			},
			"starting block height must be positive: 0",
		},
		{
			"duplicate position",
			func(genState *types.GenesisState) {
				genState.Positions = []types.Position{
					validPosition, validPosition,
				}
			},
			"duplicate position: cosmos1qsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv4uhu3, pool1",
		},
		{
			"invalid historical rewards: invalid denom",
			func(genState *types.GenesisState) {
				genState.HistoricalRewards = []types.HistoricalRewardsRecord{
					{Denom: "invalid!", Period: 2, HistoricalRewards: validHist},
				}
			},
			"invalid historical rewards denom: invalid denom: invalid!",
		},
		{
			"invalid historical rewards: zero reference count",
			func(genState *types.GenesisState) {
				hist := validHist
				hist.ReferenceCount = 0
				genState.HistoricalRewards = []types.HistoricalRewardsRecord{
					{Denom: "pool1", Period: 2, HistoricalRewards: hist},
				}
			},
			"reference count must be positive",
		},
		{
			"invalid historical rewards: invalid reference count",
			func(genState *types.GenesisState) {
				hist := validHist
				hist.ReferenceCount = 3
				genState.HistoricalRewards = []types.HistoricalRewardsRecord{
					{Denom: "pool1", Period: 2, HistoricalRewards: hist},
				}
			},
			"reference count must not exceed 2",
		},
		{
			"duplicate historical rewards",
			func(genState *types.GenesisState) {
				genState.HistoricalRewards = []types.HistoricalRewardsRecord{
					{
						Denom:             "pool1",
						Period:            1,
						HistoricalRewards: validHist,
					},
					{
						Denom:             "pool1",
						Period:            1,
						HistoricalRewards: validHist,
					},
				}
			},
			"duplicate historical rewards: pool1, 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lastBlockTime := utils.ParseTime("2022-01-01T00:00:00Z")
			genState := types.GenesisState{
				Params:        types.DefaultParams(),
				LastBlockTime: &lastBlockTime,
				LastPlanId:    1,
				Plans:         []types.Plan{validPlan},
				Farms:         []types.FarmRecord{{Denom: "pool1", Farm: validFarm}},
				Positions:     []types.Position{validPosition},
				HistoricalRewards: []types.HistoricalRewardsRecord{
					{Denom: "pool1", Period: 0, HistoricalRewards: validHist},
				},
			}
			require.NoError(t, genState.Validate())
			tc.malleate(&genState)
			err := genState.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
