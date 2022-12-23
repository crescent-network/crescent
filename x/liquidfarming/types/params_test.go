package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
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
			"invalid fee collector",
			func(params *types.Params) {
				params.FeeCollector = "invalidaddr"
			},
			"invalid fee collector address: invalidaddr",
		},
		{
			"invalid rewards auction duration",
			func(params *types.Params) {
				params.RewardsAuctionDuration = 0
			},
			"rewards auction duration must be positive: 0",
		},
		{
			"invalid pool id in liquid farm",
			func(params *types.Params) {
				params.LiquidFarms = []types.LiquidFarm{
					types.NewLiquidFarm(0, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec()),
				}
			},
			"invalid liquid farm: pool id must not be 0",
		},
		{
			"invalid minimum farm amount in liquid farm",
			func(params *types.Params) {
				params.LiquidFarms = []types.LiquidFarm{
					types.NewLiquidFarm(1, sdk.NewInt(-1), sdk.ZeroInt(), sdk.ZeroDec()),
				}
			},
			"invalid liquid farm: minimum farm amount must be 0 or positive value: -1",
		},
		{
			"invalid minimum bid amount in liquid farm",
			func(params *types.Params) {
				params.LiquidFarms = []types.LiquidFarm{
					types.NewLiquidFarm(1, sdk.ZeroInt(), sdk.NewInt(-1), sdk.ZeroDec()),
				}
			},
			"invalid liquid farm: minimum bid amount must be 0 or positive value: -1",
		},
		{
			"invalid fee rate in liquid farm",
			func(params *types.Params) {
				params.LiquidFarms = []types.LiquidFarm{
					types.NewLiquidFarm(1, sdk.ZeroInt(), sdk.ZeroInt(), sdk.NewDec(-1)),
				}
			},
			"invalid liquid farm: fee rate must be 0 or positive value: -1.000000000000000000",
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
