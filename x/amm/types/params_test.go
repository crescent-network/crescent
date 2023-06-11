package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
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
