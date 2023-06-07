package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestValidateFeeRates(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		makerFeeRate, takerFeeRate sdk.Dec
		expectedErr                string
	}{
		{
			"happy case",
			utils.ParseDec("0.001"),
			utils.ParseDec("0.003"),
			"",
		},
		{
			"happy case 2",
			utils.ParseDec("-0.001"),
			utils.ParseDec("0.003"),
			"",
		},
		{
			"too high maker fee rate",
			utils.ParseDec("1.01"),
			utils.ParseDec("0.003"),
			"maker fee rate must be in range [-0.003000000000000000, 1]: 1.010000000000000000",
		},
		{
			"too low maker fee rate",
			utils.ParseDec("-1.01"),
			utils.ParseDec("1"),
			"maker fee rate must be in range [-1.000000000000000000, 1]: -1.010000000000000000",
		},
		{
			"too low maker fee rate 2",
			utils.ParseDec("-0.004"),
			utils.ParseDec("0.003"),
			"maker fee rate must be in range [-0.003000000000000000, 1]: -0.004000000000000000",
		},
		{
			"too high taker fee rate",
			utils.ParseDec("0.001"),
			utils.ParseDec("1.01"),
			"taker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low taker fee rate",
			utils.ParseDec("0.001"),
			utils.ParseDec("-0.001"),
			"taker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMakerTakerFeeRates(tc.makerFeeRate, tc.takerFeeRate)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
