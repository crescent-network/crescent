package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/budget/types"
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
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:01Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"not included, after ended",
			false,
			types.MustParseRFC3339("2021-12-03T00:00:01Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"included on start time",
			true,
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-03T00:00:00Z"),
		},
		{
			"not included on end time",
			false,
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-01T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
		},
		{
			"not included on same start time and end time",
			false,
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
			types.MustParseRFC3339("2021-12-02T00:00:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedResult, types.DateRangeIncludes(tc.startTime, tc.endTime, tc.targeTime))
		})
	}
}

func TestDeriveAddress(t *testing.T) {
	testCases := []struct {
		addressType     types.AddressType
		moduleName      string
		name            string
		expectedAddress string
	}{
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/fee_collector?type=1
			types.AddressType20Bytes,
			"",
			"fee_collector",
			"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
		},
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/GravityDEXFarmingBudget?module_name=farming
			types.AddressType32Bytes,
			"farming",
			"GravityDEXFarmingBudget",
			"cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
		},
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/?module_name=budget&type=1
			types.AddressType20Bytes,
			types.ModuleName,
			"",
			"cosmos1ptuk4rkky23ef69j5gujsnhyd6d857cwr02kz6",
		},
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/test1?module_name=budget&type=1
			types.AddressType20Bytes,
			types.ModuleName,
			"test1",
			"cosmos1j6y8plh9yyurax3srcw87z7vu3gr3uluhmsk96",
		},
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/test1?module_name=budget&type=0
			types.AddressType32Bytes,
			types.ModuleName,
			"test1",
			"cosmos1tfpll5msf3nz3ud2ey29hk9wczhtg7fg0cttp9q3082qrtkurvdsyh32gh",
		},
		{
			// http://127.0.0.1:1317/cosmos/budget/v1beta1/addresses/?module_name=test2
			types.AddressType32Bytes,
			"test2",
			"",
			"cosmos1v9ejakp386det8xftkvvazvqud43v3p5mmjdpnuzy3gw84h4dwxsfn6dly",
		},
		{
			types.AddressType32Bytes,
			"test2",
			"test2",
			"cosmos1qmsgyd6yu06uryqtw7t6lg7ua5ll7s3ej828fcqfakrphppug4xqcx7w45",
		},
		{
			types.AddressType20Bytes,
			"",
			"test2",
			"cosmos1vqcr4c3tnxyxr08rk28n8mkphe6c5gfuk5eh34",
		},
		{
			types.AddressType20Bytes,
			"test2",
			"",
			"cosmos1vqcr4c3tnxyxr08rk28n8mkphe6c5gfuk5eh34",
		},
		{
			types.AddressType20Bytes,
			"test2",
			"test2",
			"cosmos15642je7gk5lxugnqx3evj3jgfjdjv3q0nx6wn7",
		},
		{
			3,
			"test2",
			"invalidAddressType",
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedAddress, types.DeriveAddress(tc.addressType, tc.moduleName, tc.name).String())
		})
	}
}
