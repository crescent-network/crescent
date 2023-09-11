package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/simulation"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	pool := types.NewPool(1, 2, "ucre", "uusd", 10, sdk.NewDec(1), sdk.NewDec(1))
	poolState := types.NewPoolState(123, utils.ParseDec("1.01231111"))
	position := types.NewPosition(1, 2, utils.TestAddress(1), -500, 500)
	tickInfo := types.NewTickInfo(sdk.NewInt(1000_000000), sdk.NewInt(1000_000000))
	farmingPlan := types.NewFarmingPlan(
		1, "Farming plan", utils.TestAddress(1), utils.TestAddress(2),
		[]types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2023-07-01T00:00:00Z"),
		false)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.LastPoolIdKey, Value: sdk.Uint64ToBigEndian(10)},
			{Key: types.LastPositionIdKey, Value: sdk.Uint64ToBigEndian(100)},
			{Key: types.GetPoolKey(1), Value: cdc.MustMarshal(&pool)},
			{Key: types.GetPoolStateKey(10), Value: cdc.MustMarshal(&poolState)},
			{Key: types.GetPositionKey(100), Value: cdc.MustMarshal(&position)},
			{Key: types.GetTickInfoKey(1, -100), Value: cdc.MustMarshal(&tickInfo)},
			{Key: types.LastFarmingPlanIdKey, Value: sdk.Uint64ToBigEndian(100)},
			{Key: types.GetFarmingPlanKey(5), Value: cdc.MustMarshal(&farmingPlan)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LastPoolId", fmt.Sprintf("%d\n%d", 10, 10)},
		{"LastPositionId", fmt.Sprintf("%d\n%d", 100, 100)},
		{"Pool", fmt.Sprintf("%v\n%v", pool, pool)},
		{"PoolState", fmt.Sprintf("%v\n%v", poolState, poolState)},
		{"Position", fmt.Sprintf("%v\n%v", position, position)},
		{"TickInfo", fmt.Sprintf("%v\n%v", tickInfo, tickInfo)},
		{"LastFarmingPlanId", fmt.Sprintf("%d\n%d", 100, 100)},
		{"FarmingPlan", fmt.Sprintf("%v\n%v", farmingPlan, farmingPlan)},
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
