package simulation_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/mint/simulation"
	"github.com/cosmosquad-labs/squad/x/mint/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeLastBlockTimeStore(t *testing.T) {

	cdc := simapp.MakeTestEncodingConfig()
	dec := simulation.NewDecodeStore(cdc.Marshaler)

	tc := squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z")

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.LastBlockTimeKey, Value: sdk.FormatTimeBytes(tc)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LastBlockTime", fmt.Sprintf("%v\n%v", tc, tc)},
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
