package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v2/x/farming/simulation"
)

func TestParamChanges(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"farming/PrivatePlanCreationFee", "PrivatePlanCreationFee", "[{\"denom\":\"stake\",\"amount\":\"98498081\"}]", "farming"},
		{"farming/NextEpochDays", "NextEpochDays", "1", "farming"},
		{"farming/FarmingFeeCollector", "FarmingFeeCollector", "\"cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x\"", "farming"},
		{"farming/MaxNumPrivatePlans", "MaxNumPrivatePlans", "4575", "farming"},
	}

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 4)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
