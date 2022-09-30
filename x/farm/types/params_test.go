package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func TestParams_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(params *types.Params)
		expectedErr string // empty means no error
	}{
		{
			"valid params",
			func(params *types.Params) {},
			"",
		},
		{
			"invalid private plan creation fee",
			func(params *types.Params) {
				params.PrivatePlanCreationFee = sdk.Coins{utils.ParseCoin("0stake")}
			},
			"invalid private plan creation fee: coin 0stake amount is not positive",
		},
		{
			"invalid fee collector",
			func(params *types.Params) {
				params.FeeCollector = "invalidaddr"
			},
			"invalid fee collector address: invalidaddr",
		},
		{
			"zero max num private plans",
			func(params *types.Params) {
				params.MaxNumPrivatePlans = 0
			},
			"",
		},
		{
			"zero max block duration",
			func(params *types.Params) {
				params.MaxBlockDuration = 0
			},
			"max block duration must be positive",
		},
		{
			"negative max block duration",
			func(params *types.Params) {
				params.MaxBlockDuration = -1
			},
			"max block duration must be positive",
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
