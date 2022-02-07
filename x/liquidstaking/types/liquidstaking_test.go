package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func TestActiveCondition(t *testing.T) {
	testCases := []struct {
		validator stakingtypes.Validator
		//whitelist      []types.WhitelistedValidator
		whitelisted    bool
		expectedOutput bool
	}{
		// active case 1
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			expectedOutput: true,
		},
		// active case 2
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          true,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			expectedOutput: true,
		},
		// inactive case 1
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    false,
			expectedOutput: false,
		},
		// inactive case 1
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.Int{},
				DelegatorShares: sdk.Dec{},
			},
			whitelisted:    true,
			expectedOutput: false,
		},
		// inactive case 2
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(0),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			expectedOutput: false,
		},
		// inactive case 3
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Unspecified,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		require.IsType(t, stakingtypes.Validator{}, tc.validator)
		output := types.ActiveCondition(tc.validator, tc.whitelisted)
		require.EqualValues(t, tc.expectedOutput, output)
	}
}
