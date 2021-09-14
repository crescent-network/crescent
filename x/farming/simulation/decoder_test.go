package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/tendermint/farming/x/farming/simulation"
	"github.com/tendermint/farming/x/farming/types"
)

func TestDecodeFarmingStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	basePlan := types.BasePlan{}
	staking := types.Staking{}
	queuedStaking := types.QueuedStaking{}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.PlanKeyPrefix, Value: cdc.MustMarshal(&basePlan)},
			{Key: types.StakingKeyPrefix, Value: cdc.MustMarshal(&staking)},
			{Key: types.QueuedStakingKeyPrefix, Value: cdc.MustMarshal(&queuedStaking)},
			// TODO: f1 structs, indexes
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Plan", fmt.Sprintf("%v\n%v", basePlan, basePlan)},
		{"Staking", fmt.Sprintf("%v\n%v", staking, staking)},
		{"QueuedStaking", fmt.Sprintf("%v\n%v", queuedStaking, queuedStaking)},
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
