package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmosquad-labs/squad/types"
	"github.com/stretchr/testify/require"
)

func TestGetShareValue(t *testing.T) {
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("0.9")), sdk.NewInt(90))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(110))

	// truncated
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(101), sdk.MustNewDecFromStr("0.9")), sdk.NewInt(90))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(101), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(111))

	require.EqualValues(t, types.GetShareValue(sdk.NewInt(100), sdk.MustNewDecFromStr("0")), sdk.NewInt(0))
	require.EqualValues(t, types.GetShareValue(sdk.NewInt(0), sdk.MustNewDecFromStr("1.1")), sdk.NewInt(0))
}

func TestAddOrInit(t *testing.T) {
	strIntMap := make(types.StrIntMap)

	// Set when the key not existed on the map
	strIntMap.AddOrSet("a", sdk.NewInt(1))
	require.Equal(t, strIntMap["a"], sdk.NewInt(1))

	// Added when the key existed on the map
	strIntMap.AddOrSet("a", sdk.NewInt(1))
	require.Equal(t, strIntMap["a"], sdk.NewInt(2))
}

func TestParseTime(t *testing.T) {
	normalCase := "9999-12-31T00:00:00Z"
	normalRes, err := time.Parse(time.RFC3339, normalCase)
	require.NoError(t, err)
	errorCase := "9999-12-31T00:00:00_ErrorCase"
	_, err = time.Parse(time.RFC3339, errorCase)
	require.PanicsWithError(t, err.Error(), func() { types.ParseTime(errorCase) })
	require.Equal(t, normalRes, types.ParseTime(normalCase))
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
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
			types.ParseTime("2021-12-04T00:00:00Z"),
		},
		{
			"not overlapped on same end time and start time",
			false,
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"not overlapped on same end time and start time 2",
			false,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
		},
		{
			"for the same time, it doesn't seem to overlap",
			false,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
		},
		{
			"end time and start time differs by a little amount",
			true,
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00.01Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"overlap #1",
			true,
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-04T00:00:00Z"),
		},
		{
			"overlap #2 - same ranges",
			true,
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"overlap #3 - one includes another",
			true,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-04T00:00:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedResult, types.DateRangesOverlap(tc.startTimeA, tc.endTimeA, tc.startTimeB, tc.endTimeB))
			require.Equal(t, tc.expectedResult, types.DateRangesOverlap(tc.startTimeB, tc.endTimeB, tc.startTimeA, tc.endTimeA))
		})
	}
}

func TestDateRangeIncludes(t *testing.T) {
	testCases := []struct {
		name           string
		expectedResult bool
		targeTime      time.Time
		startTime      time.Time
		endTime        time.Time
	}{
		{
			"not included, before started",
			false,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:01Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"not included, after ended",
			false,
			types.ParseTime("2021-12-03T00:00:01Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"included on start time",
			true,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-03T00:00:00Z"),
		},
		{
			"not included on end time",
			false,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-01T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
		},
		{
			"not included on same start time and end time",
			false,
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
			types.ParseTime("2021-12-02T00:00:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedResult, types.DateRangeIncludes(tc.startTime, tc.endTime, tc.targeTime))
		})
	}
}
