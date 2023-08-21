package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/exchange/simulation"
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
		{"exchange/Fees", "Fees", `{"market_creation_fee":[{"denom":"stake","amount":"1000000"}],"default_maker_fee_rate":"-0.004074449554266347","default_taker_fee_rate":"0.007709506529800694","default_order_source_fee_ratio":"0.090000000000000000"}`, "exchange"},
		{"exchange/MaxOrderPriceRatio", "MaxOrderPriceRatio", `"0.167441417343774193"`, "exchange"},
	}

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
