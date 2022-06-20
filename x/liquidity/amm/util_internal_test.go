package amm

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	i, found = findFirstTrueCondition(0, len(arr)-1, func(i int) bool {
		return arr[i] > 25
	})
	require.False(t, found)
	i, found = findFirstTrueCondition(len(arr)-1, 0, func(i int) bool {
		return arr[i] < 1
	})
	require.False(t, found)
}