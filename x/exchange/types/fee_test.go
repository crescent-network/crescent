package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestValidateFees(t *testing.T) {
	for _, tc := range []struct {
		name                                            string
		makerFeeRate, takerFeeRate, orderSourceFeeRatio sdk.Dec
		expectedErr                                     string
	}{
		{
			"happy case",
			utils.ParseDec("0.001"),
			utils.ParseDec("0.003"),
			utils.ParseDec("0.5"),
			"",
		},
		{
			"too high maker fee rate",
			utils.ParseDec("1.01"),
			utils.ParseDec("0.003"),
			utils.ParseDec("0.5"),
			"maker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low maker fee rate",
			utils.ParseDec("-0.001"),
			utils.ParseDec("1"),
			utils.ParseDec("0.5"),
			"maker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
		{
			"too high taker fee rate",
			utils.ParseDec("0.001"),
			utils.ParseDec("1.01"),
			utils.ParseDec("0.5"),
			"taker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low taker fee rate",
			utils.ParseDec("0.001"),
			utils.ParseDec("-0.001"),
			utils.ParseDec("0.5"),
			"taker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
		{
			"too high order source fee ratio",
			utils.ParseDec("0.001"),
			utils.ParseDec("0.002"),
			utils.ParseDec("1.01"),
			"order source fee ratio must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low order source fee ratio",
			utils.ParseDec("0.001"),
			utils.ParseDec("0.002"),
			utils.ParseDec("-0.01"),
			"order source fee ratio must be in range [0, 1]: -0.010000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateFees(tc.makerFeeRate, tc.takerFeeRate, tc.orderSourceFeeRatio)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
