package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v4/x/marketmaker/simulation"
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
		{"marketmaker/DepositAmount", "DepositAmount", "[{\"denom\":\"stake\",\"amount\":\"98498081\"}]", "marketmaker"},
		{"marketmaker/IncentiveBudgetAddress", "IncentiveBudgetAddress", "\"cosmos1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sqanrtcm\"", "marketmaker"},
		{"marketmaker/Common", "Common", "{\"min_open_ratio\":\"0.500000000000000000\",\"min_open_depth_ratio\":\"0.100000000000000000\",\"max_downtime\":20,\"max_total_downtime\":100,\"min_hours\":16,\"min_days\":22}", "marketmaker"},
		{"marketmaker/IncentivePairs", "IncentivePairs", "[{\"pair_id\":\"2\",\"update_time\":\"0001-01-01T00:00:00Z\",\"incentive_weight\":\"0.000000000000000000\",\"max_spread\":\"0.000000000000000000\",\"min_width\":\"0.000000000000000000\",\"min_depth\":\"0\"},{\"pair_id\":\"3\",\"update_time\":\"0001-01-01T00:00:00Z\",\"incentive_weight\":\"0.000000000000000000\",\"max_spread\":\"0.000000000000000000\",\"min_width\":\"0.000000000000000000\",\"min_depth\":\"0\"},{\"pair_id\":\"4\",\"update_time\":\"0001-01-01T00:00:00Z\",\"incentive_weight\":\"0.000000000000000000\",\"max_spread\":\"0.000000000000000000\",\"min_width\":\"0.000000000000000000\",\"min_depth\":\"0\"}]", "marketmaker"},
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
