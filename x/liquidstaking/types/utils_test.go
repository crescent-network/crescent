package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func TestParseTime(t *testing.T) {
	normalCase := "9999-12-31T00:00:00Z"
	normalRes, err := time.Parse(time.RFC3339, normalCase)
	require.NoError(t, err)
	errorCase := "9999-12-31T00:00:00_ErrorCase"
	_, err = time.Parse(time.RFC3339, errorCase)
	require.PanicsWithError(t, err.Error(), func() { types.MustParseRFC3339(errorCase) })
	require.Equal(t, normalRes, types.MustParseRFC3339(normalCase))
}

func TestDateRangesOverlap(t *testing.T) {
	testCases := []struct {
		name           string
		expectedResult bool
		startTimeA     time.Time
		endTimeA       time.Time
		startTimeB     time.Time
		endTimeB       time.Time
	}{
		{
			"not overlapping",
			false,
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
			types.MustParseRFC3339("2021-12-04T00:00:00Z"),
		},
		{
			"same end time and start time",
			false,
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"end time and start time differs by a little amount",
			true,
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00.001Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"overlap #1",
			true,
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-04T00:00:00Z"),
		},
		{
			"overlap #2 - same ranges",
			true,
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"overlap #3 - one includes another",
			true,
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-04T00:00:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedResult, types.DateRangesOverlap(tc.startTimeA, tc.endTimeA, tc.startTimeB, tc.endTimeB))
			require.Equal(t, tc.expectedResult, types.DateRangesOverlap(tc.startTimeB, tc.endTimeB, tc.startTimeA, tc.endTimeA))
		})
	}
}
