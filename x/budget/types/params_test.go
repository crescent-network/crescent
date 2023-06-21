package types_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

func TestDefaultParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()
	require.NoError(t, defaultParams.Validate())

	paramsStr := `epoch_blocks: 1
budgets: []
`
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestValidateEpochBlocks(t *testing.T) {
	err := types.ValidateEpochBlocks(uint32(0))
	require.NoError(t, err)

	err = types.ValidateEpochBlocks(nil)
	require.EqualError(t, err, "invalid parameter type: <nil>")

	err = types.ValidateEpochBlocks(types.DefaultEpochBlocks)
	require.NoError(t, err)

	err = types.ValidateEpochBlocks(10000000000000000)
	require.EqualError(t, err, "invalid parameter type: int")
}

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name        string
		configure   func(*types.Params)
		expectedErr string
	}{
		{
			"valid independent time budgets",
			func(params *types.Params) {
				params.Budgets = []types.Budget{
					{
						Name:               "budget1-2",
						Rate:               sdk.MustNewDecFromStr("1.0"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-01T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-02T00:00:00Z"),
					},
					{
						Name:               "budget3-4",
						Rate:               sdk.MustNewDecFromStr("1.0"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-03T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-04T00:00:00Z"),
					},
				}
			},
			"",
		},
		{
			"valid transition budgets",
			func(params *types.Params) {
				params.Budgets = []types.Budget{
					{
						Name:               "budget1-4",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-01T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-04T00:00:00Z"),
					},
					{
						Name:               "budget1-2",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-01T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-02T00:00:00Z"),
					},
					{
						Name:               "budget3-4",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-03T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-04T00:00:00Z"),
					},
				}
			},
			"",
		},
		{
			"overlapped with over 1 total rate budgets",
			func(params *types.Params) {
				params.Budgets = []types.Budget{
					{
						Name:               "budget1-4",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-01T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-04T00:00:00Z"),
					},
					{
						Name:               "budget1-2",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-01T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-02T00:00:00Z"),
					},
					{
						Name:               "budget3-4",
						Rate:               sdk.MustNewDecFromStr("0.5"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-03T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-04T00:00:00Z"),
					},
					{
						Name:               "budget3-5",
						Rate:               sdk.MustNewDecFromStr("0.001"),
						SourceAddress:      sAddr1.String(),
						DestinationAddress: dAddr1.String(),
						StartTime:          types.MustParseRFC3339("2022-01-03T00:00:00Z"),
						EndTime:            types.MustParseRFC3339("2022-01-05T00:00:00Z"),
					},
				}
			},
			"total rate for source address cosmos1g6umphvhteymdm6n2arju4q2h0d8c78p2t7p4tjadlcw98w6ylrqfpwqex must not exceed 1: 1.001000000000000000: invalid total rate of the budgets with the same source address",
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
