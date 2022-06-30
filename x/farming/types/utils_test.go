package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func TestParseTime(t *testing.T) {
	normalCase := "9999-12-31T00:00:00Z"
	normalRes, err := time.Parse(time.RFC3339, normalCase)
	require.NoError(t, err)
	errorCase := "9999-12-31T00:00:00_ErrorCase"
	_, err = time.Parse(time.RFC3339, errorCase)
	require.PanicsWithError(t, err.Error(), func() { types.ParseTime(errorCase) })
	require.Equal(t, normalRes, types.ParseTime(normalCase))
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

func TestDeriveAddress(t *testing.T) {
	testCases := []struct {
		addressType     types.AddressType
		moduleName      string
		name            string
		expectedAddress string
	}{
		{
			types.ReserveAddressType,
			types.ModuleName,
			"StakingReserveAcc|uatom",
			"cosmos1qxs9gxctmd637l7ckpc99kw6ax6thgxx5kshpgzc8kup675xp9dsank7up",
		},
		{
			types.ReserveAddressType,
			types.ModuleName,
			"StakingReserveAcc|stake",
			"cosmos1jn5vt4c3xg38ud89xjl8aumlf3akgdpllmt986w5tj9lureh65dsvk5z3t",
		},
		{
			types.AddressType20Bytes,
			"",
			"fee_collector",
			"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
		},
		{
			types.AddressType32Bytes,
			"farming",
			"GravityDEXFarmingBudget",
			"cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
		},
		{
			types.AddressType20Bytes,
			types.ModuleName,
			"",
			"cosmos1g8n25wpvvs38dec43jtt5a8w2td3nkmz6d2qfh",
		},
		{
			types.AddressType20Bytes,
			types.ModuleName,
			"test1",
			"cosmos19jjdxeykth523wg4xetyf2gr07pykjn60egn0y",
		},
		{
			types.AddressType32Bytes,
			types.ModuleName,
			"test1",
			"cosmos1tveg5at4u8tzulwrq4qq4gnxln729t8r72aphsx9euwsw0cmeq7qudxdv8",
		},
		{
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
