package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func TestParams_Validate(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

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
			"invalid rewards auction duration",
			func(params *types.Params) {
				params.RewardsAuctionDuration = 0
			},
			"rewards auction duration must be positive: 0",
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
