package amm

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
)

func Test_findFirstTrueCondition(t *testing.T) {
	arr := []int{1, 5, 9, 10, 14, 20, 25}
	i, found := findFirstTrueCondition(0, len(arr)-1, func(i int) bool {
		return arr[i] >= 9
	})
	require.True(t, found)
	require.Equal(t, 2, i)
	i, found = findFirstTrueCondition(len(arr)-1, 0, func(i int) bool {
		return arr[i] < 9
	})
	require.True(t, found)
	require.Equal(t, 1, i)
	_, found = findFirstTrueCondition(0, len(arr)-1, func(i int) bool {
		return arr[i] > 25
	})
	require.False(t, found)
	_, found = findFirstTrueCondition(len(arr)-1, 0, func(i int) bool {
		return arr[i] < 1
	})
	require.False(t, found)
}

func Test_poolOrderPriceGapRatio(t *testing.T) {
	for _, tc := range []struct {
		poolPrice    sdk.Dec
		currentPrice sdk.Dec
		expected     sdk.Dec
	}{
		{utils.ParseDec("1"), utils.ParseDec("1"), utils.ParseDec("0.00003")},
		{utils.ParseDec("1"), utils.ParseDec("1.005"), utils.ParseDec("0.000065")},
		{utils.ParseDec("1"), utils.ParseDec("1.01"), utils.ParseDec("0.0001")},
		{utils.ParseDec("1"), utils.ParseDec("1.015"), utils.ParseDec("0.00055")},
		{utils.ParseDec("1"), utils.ParseDec("1.02"), utils.ParseDec("0.001")},
		{utils.ParseDec("1"), utils.ParseDec("1.05"), utils.ParseDec("0.0025")},
		{utils.ParseDec("1"), utils.ParseDec("1.1"), utils.ParseDec("0.005")},
		{utils.ParseDec("1"), utils.ParseDec("10"), utils.ParseDec("0.005")},
		{utils.ParseDec("1"), utils.ParseDec("0.99"), utils.ParseDec("0.0001")},
		{utils.ParseDec("1"), utils.ParseDec("0.98"), utils.ParseDec("0.001")},
		{utils.ParseDec("1"), utils.ParseDec("0.9"), utils.ParseDec("0.005")},
		{utils.ParseDec("1"), utils.ParseDec("0.1"), utils.ParseDec("0.005")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, poolOrderPriceGapRatio(tc.poolPrice, tc.currentPrice)))
		})
	}
}
