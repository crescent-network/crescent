package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v3/x/liquidity/simulation"
)

func TestParamChanges(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 5)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"liquidity/BatchSize", "BatchSize", "5", "liquidity"},
		{"liquidity/TickPrecision", "TickPrecision", "3", "liquidity"},
		{"liquidity/MaxPriceLimitRatio", "MaxPriceLimitRatio", "\"0.187775771919630322\"", "liquidity"},
		{"liquidity/WithdrawFeeRate", "WithdrawFeeRate", "\"0.001029277270893505\"", "liquidity"},
		{"liquidity/MaxOrderLifespan", "MaxOrderLifespan", "\"87434677422456\"", "liquidity"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
