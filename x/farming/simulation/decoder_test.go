package simulation_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/tendermint/farming/x/farming/simulation"
)

func TestDecodeFarmingStore(t *testing.T) {

	cdc := simapp.MakeTestEncodingConfig()
	_ = simulation.NewDecodeStore(cdc.Marshaler)

	// TODO: not implemented yet

	// liquidityPool := types.Pool{}
	// liquidityPool.Id = 1
	// liquidityPoolBatch := types.NewPoolBatch(1, 1)

	// kvPairs := kv.Pairs{
	// 	Pairs: []kv.Pair{
	// 		{Key: types.PoolKeyPrefix, Value: cdc.MustMarshalBinaryBare(&liquidityPool)},
	// 		{Key: types.PoolBatchKeyPrefix, Value: cdc.MustMarshalBinaryBare(&liquidityPoolBatch)},
	// 		{Key: []byte{0x99}, Value: []byte{0x99}},
	// 	},
	// }

	// tests := []struct {
	// 	name        string
	// 	expectedLog string
	// }{
	// 	{"Pool", fmt.Sprintf("%v\n%v", liquidityPool, liquidityPool)},
	// 	{"PoolBatch", fmt.Sprintf("%v\n%v", liquidityPoolBatch, liquidityPoolBatch)},
	// 	{"other", ""},
	// }
	// for i, tt := range tests {
	// 	i, tt := i, tt
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		switch i {
	// 		case len(tests) - 1:
	// 			require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
	// 		default:
	// 			require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
	// 		}
	// 	})
	// }
}
