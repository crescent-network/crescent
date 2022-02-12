package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
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
	} {
		t.Run("", func(t *testing.T) {
			poolCoinDenom := types.PoolCoinDenom(tc.poolId)
			poolId := types.ParsePoolCoinDenom(poolCoinDenom)
			require.Equal(t, tc.expected, poolCoinDenom)
			require.Equal(t, tc.poolId, poolId)
		})
	}
}
