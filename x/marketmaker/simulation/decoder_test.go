package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v5/x/marketmaker/simulation"
	"github.com/crescent-network/crescent/v5/x/marketmaker/types"
)

func TestDecodeMarketMakerStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	mm := types.MarketMaker{}
	deposit := types.Deposit{}
	incentive := types.Incentive{}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.MarketMakerKeyPrefix, Value: cdc.MustMarshal(&mm)},
			{Key: types.DepositKeyPrefix, Value: cdc.MustMarshal(&deposit)},
			{Key: types.IncentiveKeyPrefix, Value: cdc.MustMarshal(&incentive)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"MarketMaker", fmt.Sprintf("%v\n%v", mm, mm)},
		{"Deposit", fmt.Sprintf("%v\n%v", deposit, deposit)},
		{"Incentive", fmt.Sprintf("%v\n%v", incentive, incentive)},
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
