package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/budget/simulation"
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
		{"budget/EpochBlocks", "EpochBlocks", "6", "budget"},
		{"budget/Budgets", "Budgets", `[{"name":"simulation-test-FpXzp","rate":"0.300000000000000000","source_address":"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta","destination_address":"cosmos1ke7rn6vl3vmeasmcrxdm3pfrt37fsg5jfrex80pp3hvhwgu4h4usxgvk3e","start_time":"2000-01-01T00:00:00Z","end_time":"9999-12-31T00:00:00Z"}]`, "budget"},
	}

	paramChanges := simulation.ParamChanges(r)

	require.Len(t, paramChanges, 2)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
