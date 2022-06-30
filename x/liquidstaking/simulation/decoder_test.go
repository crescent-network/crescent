package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/simulation"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

func TestDecodeLiquidStakingStore(t *testing.T) {

	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	tc := types.LiquidValidator{
		OperatorAddress: "cosmosvaloper13w4ueuk80d3kmwk7ntlhp84fk0arlm3m9ammr5",
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.LiquidValidatorsKey, Value: cdc.Marshaler.MustMarshal(&tc)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LiquidValidator", fmt.Sprintf("%v\n%v", tc, tc)},
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
