package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v2/x/liquidity/simulation"
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
		{"liquidity/TickPrecision", "TickPrecision", "2", "liquidity"},
		{"liquidity/MaxPriceLimitRatio", "MaxPriceLimitRatio", "\"0.151488470335453781\"", "liquidity"},
		{"liquidity/WithdrawFeeRate", "WithdrawFeeRate", "\"0.001083067024517151\"", "liquidity"},
		{"liquidity/MaxOrderLifespan", "MaxOrderLifespan", "\"229244578894929\"", "liquidity"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
