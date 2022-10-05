package types_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
	"github.com/crescent-network/crescent/v3/x/liquidity/amm"
)

func TestAddressDerivations(t *testing.T) {
	require.Equal(
		t, "45997F8B613746075AB2BFADE26769D3C747929499FA62D8D831D6A3DF6FCFB0",
		fmt.Sprint(types.DeriveFarmingPoolAddress(1)))
	require.Equal(
		t, "75DD27C3D455E4942C5490321A2828118588C4E7B38DFC8B501F7ED43562D972",
		fmt.Sprint(types.DeriveFarmingReserveAddress("pool1")))
}

func TestRewardsForBlock(t *testing.T) {
	for _, tc := range []struct {
		name          string
		rewardsPerDay sdk.Coins
		blockDuration time.Duration
		expected      sdk.DecCoins
	}{
		{
			"#1",
			utils.ParseCoins("100_000000stake"), 10 * time.Second,
			utils.ParseDecCoins("11574.074074074074074074stake"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rewards := types.RewardsForBlock(tc.rewardsPerDay, tc.blockDuration)
			require.Equal(t, tc.expected, rewards)
		})
	}
}

func TestPoolRewardWeight(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pool     amm.Pool
		expected sdk.Dec
	}{
		{
			"#1",
			amm.NewBasicPool(sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.Int{}),
			utils.ParseDec("1000000000"),
		},
		{
			"#2",
			amm.NewBasicPool(sdk.NewInt(200_000000), sdk.NewInt(8000_000000), sdk.Int{}),
			utils.ParseDec("1264911064.067351732799557418"),
		},
		{
			"#3",
			amm.NewRangedPool(
				sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.Int{},
				utils.ParseDec("0.9"), utils.ParseDec("1.15")),
			utils.ParseDec("16824065823.719412156326941740"),
		},
		{
			"#4",
			amm.NewRangedPool(
				sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.Int{},
				utils.ParseDec("0.99"), utils.ParseDec("1.01")),
			utils.ParseDec("200493749898.277059377702876726"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			weight := types.PoolRewardWeight(tc.pool)
			require.True(sdk.DecEq(t, tc.expected, weight))
		})
	}
}