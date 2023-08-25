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
			"too high default maker fee rate",
			func(params *types.Params) {
				params.Fees.DefaultMakerFeeRate = utils.ParseDec("1.01")
			},
			"maker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low default maker fee rate",
			func(params *types.Params) {
				params.Fees.DefaultMakerFeeRate = utils.ParseDec("-1.01")
			},
			"maker fee rate must be in range [0, 1]: -1.010000000000000000",
		},
		{
			"too high default taker fee rate",
			func(params *types.Params) {
				params.Fees.DefaultTakerFeeRate = utils.ParseDec("1.01")
			},
			"taker fee rate must be in range [0, 1]: 1.010000000000000000",
		},
		{
			"too low default taker fee rate",
			func(params *types.Params) {
				params.Fees.DefaultTakerFeeRate = utils.ParseDec("-0.001")
			},
			"taker fee rate must be in range [0, 1]: -0.001000000000000000",
		},
		{
			"too high default order source fee ratio",
			func(params *types.Params) {
				params.Fees.DefaultOrderSourceFeeRatio = utils.ParseDec("1.1")
			},
			"order source fee ratio must be in range [0, 1]: 1.100000000000000000",
		},
		{
			"too low default order source fee ratio",
			func(params *types.Params) {
				params.Fees.DefaultOrderSourceFeeRatio = utils.ParseDec("-0.001")
			},
			"order source fee ratio must be in range [0, 1]: -0.001000000000000000",
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
