package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestValidateTickSpacing(t *testing.T) {
	for i, tc := range []struct {
		prevTickSpacing, tickSpacing uint32
		expectedErr                  string
	}{
		{50, 10, ""},
		{50, 5, ""},
		{50, 1, ""},
		{10, 50, "tick spacing must be a divisor of previous tick spacing 10"},
		{10, 30, "tick spacing 30 is not allowed"},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			err := types.ValidateTickSpacing(tc.prevTickSpacing, tc.tickSpacing)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

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
			"invalid pool creation fee",
			func(params *types.Params) {
				params.PoolCreationFee = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid pool creation fee: coin 0ucre amount is not positive",
		},
		{
			"not allowed default tick spacing",
			func(params *types.Params) {
				params.DefaultTickSpacing = 7
			},
			"tick spacing 7 is not allowed",
		},
		{
			"invalid private farming plan creation fee",
			func(params *types.Params) {
				params.PrivateFarmingPlanCreationFee = sdk.Coins{sdk.NewInt64Coin("ucre", 0)}
			},
			"invalid private farming plan creation fee: coin 0ucre amount is not positive",
		},
		{
			"invalid max farming block time",
			func(params *types.Params) {
				params.MaxFarmingBlockTime = 0
			},
			"max farming block time must be positive: 0s",
		},
		{
			"invalid max farming block time 2",
			func(params *types.Params) {
				params.MaxFarmingBlockTime = -time.Second
			},
			"max farming block time must be positive: -1s",
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

func TestValidatePriceRange(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		lowerPrice, upperPrice sdk.Dec
		expectedErr            string
	}{
		{
			"happy case",
			utils.ParseDec("0.5"), utils.ParseDec("2"),
			"",
		},
		{
			"non-positive lower price",
			utils.ParseDec("-0.1"), utils.ParseDec("2"),
			"lower price must be positive: -0.100000000000000000: invalid request",
		},
		{
			"lower price < min price",
			sdk.NewDecWithPrec(1, sdk.Precision), utils.ParseDec("2"),
			"lower price must not be lower than the minimum: 0.000000000000000001 < 0.000000000000010000: invalid request",
		},
		{
			"non-positive upper price",
			utils.ParseDec("0.5"), utils.ParseDec("-0.1"),
			"upper price must be positive: -0.100000000000000000: invalid request",
		},
		{
			"upper price > max price",
			utils.ParseDec("0.5"), sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 40)),
			"upper price must not be higher than the maximum: " +
				"10000000000000000000000000000000000000000.000000000000000000 > " +
				"1000000000000000000000000.000000000000000000: invalid request",
		},
		{
			"invalid upper price",
			utils.ParseDec("0.5"), utils.ParseDec("1.99999"),
			"invalid upper tick price: 1.999990000000000000: invalid request",
		},
		{
			"lower price >= upper price",
			utils.ParseDec("0.5"), utils.ParseDec("0.5"),
			"lower price must be lower than upper price: " +
				"0.500000000000000000 >= 0.500000000000000000: invalid request",
		},
		{
			"invalid lower price",
			utils.ParseDec("0.499999"), utils.ParseDec("2"),
			"invalid lower tick price: 0.499999000000000000: invalid request",
		},
		{
			"invalid upper price",
			utils.ParseDec("0.5"), utils.ParseDec("1.99999"),
			"invalid upper tick price: 1.999990000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidatePriceRange(tc.lowerPrice, tc.upperPrice)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
