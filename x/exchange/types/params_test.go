package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestParams_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(params *types.Params)
		expectedErr string
	}{
		{
			"happy case",
			func(params *types.Params) {},
			"",
		},
		{
			"invalid market creation fee",
			func(params *types.Params) {
				params.MarketCreationFee = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid market creation fee: coin 0ucre amount is not positive",
		},
		{
			"invalid default fees",
			func(params *types.Params) {
				params.DefaultFees.MakerFeeRate = utils.ParseDec("1.01")
			},
			"invalid default fees: maker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"negative max order lifespan",
			func(params *types.Params) {
				params.MaxOrderLifespan = -time.Hour
			},
			"max order lifespan must not be negative: -1h0m0s",
		},
		{
			"too low max order price ratio",
			func(params *types.Params) {
				params.MaxOrderPriceRatio = sdk.ZeroDec()
			},
			"max order price ratio must be in range (0.0, 1.0): 0.000000000000000000",
		},
		{
			"too high max order price ratio",
			func(params *types.Params) {
				params.MaxOrderPriceRatio = sdk.OneDec()
			},
			"max order price ratio must be in range (0.0, 1.0): 1.000000000000000000",
		},
		{
			"invalid default order quantity limits",
			func(params *types.Params) {
				params.DefaultOrderQuantityLimits.Min = sdk.NewInt(1000000)
				params.DefaultOrderQuantityLimits.Max = sdk.NewInt(10000)
			},
			"invalid default order quantity limits: the minimum value is greater than the maximum value: 1000000 > 10000",
		},
		{
			"invalid default order quote limits",
			func(params *types.Params) {
				params.DefaultOrderQuoteLimits.Min = sdk.NewInt(10000000)
				params.DefaultOrderQuoteLimits.Max = sdk.NewInt(10000)
			},
			"invalid default order quote limits: the minimum value is greater than the maximum value: 10000000 > 10000",
		},
		{
			"zero max swap routes len",
			func(params *types.Params) {
				params.MaxSwapRoutesLen = 0
			},
			"max swap routes len must not be 0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.malleate(&params)
			err := params.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestFees_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		fees        types.Fees
		expectedErr string
	}{
		{
			"happy case",
			types.NewFees(
				utils.ParseDec("0.001"), utils.ParseDec("0.003"), utils.ParseDec("0.5")),
			"",
		},
		{
			"too high maker fee rate",
			types.NewFees(
				utils.ParseDec("1.01"), utils.ParseDec("0.003"), utils.ParseDec("0.5")),
			"maker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low maker fee rate",
			types.NewFees(
				utils.ParseDec("-0.001"), utils.ParseDec("1"), utils.ParseDec("0.5")),
			"maker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
		{
			"too high taker fee rate",
			types.NewFees(
				utils.ParseDec("0.001"), utils.ParseDec("1.01"), utils.ParseDec("0.5")),
			"taker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low taker fee rate",
			types.NewFees(
				utils.ParseDec("0.001"), utils.ParseDec("-0.001"), utils.ParseDec("0.5")),
			"taker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
		{
			"too high order source fee ratio",
			types.NewFees(
				utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("1.01")),
			"order source fee ratio must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low order source fee ratio",
			types.NewFees(
				utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("-0.01")),
			"order source fee ratio must be in range [0, 1]: -0.010000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fees.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestAmountLimits_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		limits      types.AmountLimits
		expectedErr string
	}{
		{
			"happy case",
			types.NewAmountLimits(sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30)),
			"",
		},
		{
			"negative min",
			types.NewAmountLimits(sdk.NewInt(-10000), sdk.NewIntWithDecimal(1, 30)),
			"the minimum value must be positive: -10000",
		},
		{
			"zero min",
			types.NewAmountLimits(sdk.NewInt(0), sdk.NewIntWithDecimal(1, 30)),
			"the minimum value must be positive: 0",
		},
		{
			"min > max",
			types.NewAmountLimits(sdk.NewInt(10001), sdk.NewInt(10000)),
			"the minimum value is greater than the maximum value: 10001 > 10000",
		},
		{
			"min >= max",
			types.NewAmountLimits(sdk.NewInt(10000), sdk.NewInt(10000)),
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.limits.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
