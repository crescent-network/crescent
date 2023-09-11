package types_test

import (
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

var (
	dAddr1  = sdk.AccAddress(address.Module(types.ModuleName, []byte("destinationAddr1")))
	dAddr2  = sdk.AccAddress(address.Module(types.ModuleName, []byte("destinationAddr2")))
	sAddr1  = sdk.AccAddress(address.Module(types.ModuleName, []byte("sourceAddr1")))
	sAddr2  = sdk.AccAddress(address.Module(types.ModuleName, []byte("sourceAddr2")))
	budgets = []types.Budget{
		{
			Name:               "budget0",
			Rate:               sdk.OneDec(),
			SourceAddress:      sAddr1.String(),
			DestinationAddress: dAddr1.String(),
			StartTime:          types.MustParseRFC3339("2021-08-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-08-03T00:00:00Z"),
		},
		{
			Name:               "budget1",
			Rate:               sdk.OneDec(),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-07-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-07-10T00:00:00Z"),
		},
		{
			Name:               "budget2",
			Rate:               sdk.MustNewDecFromStr("0.1"),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-07-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-07-10T00:00:00Z"),
		},
		{
			Name:               "budget3",
			Rate:               sdk.MustNewDecFromStr("0.1"),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-08-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-08-10T00:00:00Z"),
		},
		{
			Name:               "budget4",
			Rate:               sdk.OneDec(),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-08-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-08-20T00:00:00Z"),
		},
		{
			Name:               "budget5",
			Rate:               sdk.MustNewDecFromStr("0.1"),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
		},
		{
			Name:               "budget6",
			Rate:               sdk.MustNewDecFromStr("1.0"),
			SourceAddress:      sAddr2.String(),
			DestinationAddress: dAddr2.String(),
			StartTime:          types.MustParseRFC3339("2021-08-25T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("2021-08-26T00:00:00Z"),
		},
	}
)

func TestValidateBudgets(t *testing.T) {
	testCases := []struct {
		budgets       []types.Budget
		expectedError error
	}{
		{
			[]types.Budget{budgets[0], budgets[1]},
			nil,
		},
		{
			[]types.Budget{budgets[0], budgets[1], budgets[2]},
			types.ErrInvalidTotalBudgetRate,
		},
		{
			[]types.Budget{budgets[1], budgets[4]},
			nil,
		},
		{
			[]types.Budget{budgets[4], budgets[5]},
			types.ErrInvalidTotalBudgetRate,
		},
		{
			[]types.Budget{budgets[3], budgets[3]},
			types.ErrDuplicateBudgetName,
		},
		{
			[]types.Budget{
				{
					Name:               "same-source-destination-addr",
					Rate:               sdk.MustNewDecFromStr("0.1"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: sAddr2.String(),
					StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
				},
			},
			types.ErrSameSourceDestinationAddr,
		},
		{
			[]types.Budget{
				{
					Name:               "over-1-rate",
					Rate:               sdk.MustNewDecFromStr("1.01"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: dAddr2.String(),
					StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
				},
			},
			types.ErrInvalidBudgetRate,
		},
		{
			[]types.Budget{
				{
					Name:               "not-positive-rate",
					Rate:               sdk.MustNewDecFromStr("-0.01"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: dAddr2.String(),
					StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
				},
			},
			types.ErrInvalidBudgetRate,
		},
		{
			[]types.Budget{
				{
					Name:               "invalid budget name",
					Rate:               sdk.MustNewDecFromStr("0.5"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: dAddr2.String(),
					StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
				},
			},
			types.ErrInvalidBudgetName,
		},
		{
			[]types.Budget{
				{
					Name:               "invalid-destination-addr",
					Rate:               sdk.MustNewDecFromStr("0.5"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: "invalidaddr",
					StartTime:          types.MustParseRFC3339("2021-08-19T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-25T00:00:00Z"),
				},
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			[]types.Budget{
				{
					Name:               "invalid-start-time",
					Rate:               sdk.MustNewDecFromStr("0.5"),
					SourceAddress:      sAddr2.String(),
					DestinationAddress: dAddr2.String(),
					StartTime:          types.MustParseRFC3339("2021-08-20T00:00:00Z"),
					EndTime:            types.MustParseRFC3339("2021-08-19T00:00:00Z"),
				},
			},
			types.ErrInvalidStartEndTime,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := types.ValidateBudgets(tc.budgets)
			if tc.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.expectedError)
			}
		})
	}
}

func TestCollectibleBudgets(t *testing.T) {
	testCases := []struct {
		budgets     []types.Budget
		blockTime   time.Time
		expectedLen int
	}{
		{
			[]types.Budget{budgets[0], budgets[1]},
			types.MustParseRFC3339("2021-07-05T00:00:00Z"),
			1,
		},
		{
			[]types.Budget{budgets[0], budgets[1], budgets[2]},
			types.MustParseRFC3339("2021-07-05T00:00:00Z"),
			2,
		},
		{
			[]types.Budget{budgets[4], budgets[5]},
			types.MustParseRFC3339("2021-08-18T00:00:00Z"),
			1,
		},
		{
			[]types.Budget{budgets[4], budgets[5]},
			types.MustParseRFC3339("2021-08-19T00:00:00Z"),
			2,
		},
		{
			[]types.Budget{budgets[4], budgets[5]},
			types.MustParseRFC3339("2021-08-20T00:00:00Z"),
			1,
		},
		{
			[]types.Budget{budgets[5], budgets[6]},
			types.MustParseRFC3339("2021-08-25T00:00:00Z"),
			1,
		},
		{
			[]types.Budget{},
			types.MustParseRFC3339("2021-08-20T00:00:00Z"),
			0,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			require.Len(t, types.CollectibleBudgets(tc.budgets, tc.blockTime), tc.expectedLen)
		})
	}
}
