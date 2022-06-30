package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

var (
	whitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			TargetWeight:     sdk.NewInt(10),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(1),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(-1),
		},
		{
			ValidatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			TargetWeight:     sdk.NewInt(0),
		},
	}
)

func TestBTokenToNativeTokenWithFee(t *testing.T) {
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

		output := types.BTokenToNativeToken(tc.bTokenAmount, tc.bTokenTotalSupplyAmount, tc.netAmount)
		if tc.feeRate.IsPositive() {
			output = types.DeductFeeRate(output, tc.feeRate)
		}
		require.EqualValues(t, tc.expectedOutput, output)
	}
}

func TestActiveCondition(t *testing.T) {
	testCases := []struct {
		validator      stakingtypes.Validator
		whitelisted    bool
		tombstoned     bool
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
			tombstoned:     false,
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
			tombstoned:     false,
			expectedOutput: true,
		},
		// inactive case 1 (not whitelisted)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    false,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 2 (invalid tokens, delShares)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.Int{},
				DelegatorShares: sdk.Dec{},
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 3 (zero tokens)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(0),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 4 (invalid status)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Unspecified,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 5 (tombstoned)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Unbonding,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     true,
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		require.IsType(t, stakingtypes.Validator{}, tc.validator)
		output := types.ActiveCondition(tc.validator, tc.whitelisted, tc.tombstoned)
		require.EqualValues(t, tc.expectedOutput, output)
	}
}
