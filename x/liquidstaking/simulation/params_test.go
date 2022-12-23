package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v4/x/liquidstaking/simulation"
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
		{"liquidstaking/WhitelistedValidators", "WhitelistedValidators", "[]", "liquidstaking"},
		{"liquidstaking/LiquidBondDenom", "LiquidBondDenom", "\"bstake\"", "liquidstaking"},
		{"liquidstaking/UnstakeFeeRate", "UnstakeFeeRate", "\"0.010000000000000000\"", "liquidstaking"},
		{"liquidstaking/MinLiquidStakingAmount", "MinLiquidStakingAmount", "\"9727887\"", "liquidstaking"},
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
