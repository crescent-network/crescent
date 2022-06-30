package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func TestPoolReserveAddress(t *testing.T) {
	for _, tc := range []struct {
		poolId   uint64
		expected string
	}{
		{1, "cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy"},
		{2, "cosmos1a8a5ktagpr35z3s3nkrkyjvjje5ktsyuh4qssf9jymej6nh58dwq9wng4g"},
	} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.expected, types.PoolReserveAddress(tc.poolId).String())
		})
	}
}

func TestPoolCoinDenom(t *testing.T) {
	for _, tc := range []struct {
		poolId   uint64
		expected string
	}{
		{1, "pool1"},
		{10, "pool10"},
		{18446744073709551615, "pool18446744073709551615"},
	} {
		t.Run("", func(t *testing.T) {
			poolCoinDenom := types.PoolCoinDenom(tc.poolId)
			require.Equal(t, tc.expected, poolCoinDenom)
		})
	}
}

func TestParsePoolCoinDenomFailure(t *testing.T) {
	for _, tc := range []struct {
		denom      string
		expectsErr bool
	}{
		{"pool1", false},
		{"pool10", false},
		{"pool18446744073709551615", false},
		{"pool18446744073709551616", true},
		{"pool01", true},
		{"pool-10", true},
		{"pool+10", true},
		{"ucre", true},
		{"denom1", true},
	} {
		t.Run("", func(t *testing.T) {
			poolId, err := types.ParsePoolCoinDenom(tc.denom)
			if tc.expectsErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.denom, types.PoolCoinDenom(poolId))
			}
		})
	}
}
