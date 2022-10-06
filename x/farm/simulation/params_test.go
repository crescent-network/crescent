package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v3/x/farm/simulation"
)

func TestParamChanges(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 2)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"farm/FeeCollector", "FeeCollector", `"cosmos1stgwet7cl6tleugpqqqqqqqqqqqqqqqq9dhdq9"`, "farm"},
		{"farm/MaxNumPrivatePlans", "MaxNumPrivatePlans", "19", "farm"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
