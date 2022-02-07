package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/stretchr/testify/require"
)

func TestBTokenToNativeToken(t *testing.T) {
	testCases := []struct {
		bTokenAmount            sdk.Int
		bTokenTotalSupplyAmount sdk.Int
		netAmount               sdk.Dec
		feeRate                 sdk.Dec
		expectedOutput          sdk.Dec
	}{
		// reward added case
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(5100000000),
			feeRate:                 sdk.MustNewDecFromStr("0.0"),
			expectedOutput:          sdk.MustNewDecFromStr("102000000.0"),
		},
		// reward added case with fee
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(5100000000),
			feeRate:                 sdk.MustNewDecFromStr("0.005"),
			expectedOutput:          sdk.MustNewDecFromStr("101490000.0"),
		},
		// slashed case
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(4000000000),
			feeRate:                 sdk.MustNewDecFromStr("0.0"),
			expectedOutput:          sdk.MustNewDecFromStr("80000000.0"),
		},
		// slashed case with fee
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(4000000000),
			feeRate:                 sdk.MustNewDecFromStr("0.001"),
			expectedOutput:          sdk.MustNewDecFromStr("79920000.0"),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, sdk.Int{}, tc.bTokenAmount)
		require.IsType(t, sdk.Int{}, tc.bTokenTotalSupplyAmount)
		require.IsType(t, sdk.Dec{}, tc.netAmount)
		require.IsType(t, sdk.Dec{}, tc.feeRate)
		require.IsType(t, sdk.Dec{}, tc.expectedOutput)

		output := types.BTokenToNativeToken(tc.bTokenAmount, tc.bTokenTotalSupplyAmount, tc.netAmount, tc.feeRate)
		require.EqualValues(t, tc.expectedOutput, output)
	}
}
