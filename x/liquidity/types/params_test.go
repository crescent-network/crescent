package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestParams_Validate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		malleate func(*types.Params)
		errStr   string
	}{
		{
			"default params",
			func(params *types.Params) {},
			"",
		},
		{
			"negative InitialPoolCoinSupply",
			func(params *types.Params) {
				params.InitialPoolCoinSupply = sdk.NewInt(-1)
			},
			"initial pool coin supply must be positive: -1",
		},
		{
			"zero InitialPoolCoinSupply",
			func(params *types.Params) {
				params.InitialPoolCoinSupply = sdk.ZeroInt()
			},
			"initial pool coin supply must be positive: 0",
		},
		{
			"zero BatchSize",
			func(params *types.Params) {
				params.BatchSize = 0
			},
			"batch size must be positive: 0",
		},
		{
			"negative MinInitialDepositAmount",
			func(params *types.Params) {
				params.MinInitialDepositAmount = sdk.NewInt(-1)
			},
			"minimum initial deposit amount must not be negative: -1",
		},
		{
			"invalid PoolCreationFee",
			func(params *types.Params) {
				params.PoolCreationFee = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.ZeroInt()}}
			},
			"invalid pool creation fee: coin 0stake amount is not positive",
		},
		{
			"invalid FeeCollectorAddress",
			func(params *types.Params) {
				params.FeeCollectorAddress = "invalidaddr"
			},
			"invalid fee collector address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"negative MaxPriceLimitRatio",
			func(params *types.Params) {
				params.MaxPriceLimitRatio = sdk.NewDec(-1)
			},
			"max price limit ratio must not be negative: -1.000000000000000000",
		},
		{
			"negative SwapFeeRate",
			func(params *types.Params) {
				params.SwapFeeRate = sdk.NewDec(-1)
			},
			"swap fee rate must not be negative: -1.000000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.malleate(&params)
			err := params.Validate()
			if tc.errStr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errStr)
			}
		})
	}
}
