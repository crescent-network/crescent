package types_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `mint_denom:"stake" mint_pool_address:"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta" block_time_threshold:<seconds:10 > inflation_schedules:<start_time:<seconds:1640995200 > end_time:<seconds:1672531200 > amount:"300000000000000" > inflation_schedules:<start_time:<seconds:1672531200 > end_time:<seconds:1704067200 > amount:"200000000000000" > `
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	testCases := []struct {
		name        string
		configure   func(*types.Params)
		expectedErr string
	}{
		{
			"valid default params",
			func(params *types.Params) {},
			"",
		},
		{
			"empty mint denom",
			func(params *types.Params) {
				params.MintDenom = ""
			},
			"mint denom cannot be blank",
		},
		{
			"invalid mint denom",
			func(params *types.Params) {
				params.MintDenom = "a"
			},
			"invalid denom: a",
		},
		{
			"negative block time threshold",
			func(params *types.Params) {
				params.BlockTimeThreshold = -1
			},
			"block time threshold must be positive: -1",
		},
		{
			"nil inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = nil
			},
			"",
		},
		{
			"empty inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{}
			},
			"",
		},
		{
			"invalid inflation schedule start, end time",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2023-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2022-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(300000000000000),
					},
				}
			},
			"inflation end time 2022-01-01T00:00:00Z must be greater than start time 2023-01-01T00:00:00Z",
		},
		{
			"negative inflation Amount",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(-1),
					},
				}
			},
			"inflation schedule amount must be positive: -1",
		},
		{
			"too small inflation Amount",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31535999),
					},
				}
			},
			"inflation amount too small, it should be over period duration seconds: 31535999",
		},
		{
			"overlapped inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
					{
						StartTime: utils.ParseTime("2022-12-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2024-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
				}
			},
			"inflation periods cannot be overlapped 2022-01-01T00:00:00Z ~ 2023-01-01T00:00:00Z with 2022-12-01T00:00:00Z ~ 2024-01-01T00:00:00Z",
		},
		{
			"valid inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
					{
						StartTime: utils.ParseTime("2023-01-01T00:00:01Z"),
						EndTime:   utils.ParseTime("2024-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
				}
			},
			"",
		},
		{
			"same start date with end date is allowed on inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
					{
						StartTime: utils.ParseTime("2023-01-01T00:00:00Z"),
						EndTime:   utils.ParseTime("2024-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
				}
			},
			"",
		},
		{
			"empty mint pool",
			func(params *types.Params) {
				params.MintPoolAddress = ""
			},
			"invalid mint pool address: empty address string is not allowed",
		},
		{
			"invalid mint pool",
			func(params *types.Params) {
				params.MintPoolAddress = "abc123"
			},
			"invalid mint pool address: decoding bech32 failed: invalid bech32 string length 6",
		},
		{
			"valid mint pool",
			func(params *types.Params) {
				params.MintPoolAddress = "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta"
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.configure(&params)
			err := params.Validate()

			var err2 error
			for _, p := range params.ParamSetPairs() {
				err := p.ValidatorFn(reflect.ValueOf(p.Value).Elem().Interface())
				if err != nil {
					err2 = err
					break
				}
			}
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.EqualError(t, err2, tc.expectedErr)
			} else {
				require.Nil(t, err)
				require.Nil(t, err2)
			}
		})
	}
}
