package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func TestRewardsForBlock(t *testing.T) {
	for _, tc := range []struct {
		name          string
		rewardsPerDay sdk.Coins
		blockDuration time.Duration
		expected      sdk.DecCoins
	}{
		{
			"#1", utils.ParseCoins("100_000000stake"), 10 * time.Second,
			utils.ParseDecCoins("11574.074074074074074074stake"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rewards := types.RewardsForBlock(tc.rewardsPerDay, tc.blockDuration)
			require.Equal(t, tc.expected, rewards)
		})
	}
}
