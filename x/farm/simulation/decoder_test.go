package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	chain "github.com/crescent-network/crescent/v3/app"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/simulation"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	plan := types.NewPlan(
		1, "Farming Plan",
		utils.TestAddress(0), utils.TestAddress(1),
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(1, utils.ParseCoins("100_000000stake")),
		},
		utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T23:59:59Z"), true)
	farm := types.Farm{
		TotalFarmingAmount: sdk.NewInt(100_000000),
		CurrentRewards:     utils.ParseDecCoins("1_000000stake"),
		OutstandingRewards: utils.ParseDecCoins("1_000000stake"),
		Period:             2,
	}
	farmerAddr := utils.TestAddress(2)
	position := types.Position{
		Farmer:              farmerAddr.String(),
		Denom:               "pool1",
		FarmingAmount:       sdk.NewInt(100_000000),
		PreviousPeriod:      1,
		StartingBlockHeight: 10,
	}
	hist := types.HistoricalRewards{
		CumulativeUnitRewards: utils.ParseDecCoins("1.5stake"),
		ReferenceCount:        1,
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.GetPlanKey(plan.Id), Value: cdc.MustMarshal(&plan)},
			{Key: types.GetFarmKey("pool1"), Value: cdc.MustMarshal(&farm)},
			{Key: types.GetPositionKey(farmerAddr, "pool1"), Value: cdc.MustMarshal(&position)},
			{Key: types.GetHistoricalRewardsKey("pool1", 1), Value: cdc.MustMarshal(&hist)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Plan", fmt.Sprintf("%v\n%v", plan, plan)},
		{"Farm", fmt.Sprintf("%v\n%v", farm, farm)},
		{"Position", fmt.Sprintf("%v\n%v", position, position)},
		{"HistoricalRewards", fmt.Sprintf("%v\n%v", hist, hist)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
