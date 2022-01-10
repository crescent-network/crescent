package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

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
			"negative initial pool coin supply",
			func(params *types.Params) {
				params.InitialPoolCoinSupply = sdk.NewInt(-1)
			},
			"initial pool coin supply must be positive: -1",
		},
		{
			"zero initial pool coin supply",
			func(params *types.Params) {
				params.InitialPoolCoinSupply = sdk.ZeroInt()
			},
			"initial pool coin supply must be positive: 0",
		},
		{
			"zero batch size",
			func(params *types.Params) {
				params.BatchSize = 0
			},
			"batch size must be positive: 0",
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
