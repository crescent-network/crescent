package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestTickInfo_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(tickInfo *types.TickInfo)
		expectedErr string
	}{
		{
			"valid",
			func(tickInfo *types.TickInfo) {},
			"",
		},
		{
			"negative gross liquidity",
			func(tickInfo *types.TickInfo) {
				tickInfo.GrossLiquidity = sdk.NewInt(-1000_000000)
			},
			"gross liquidity must not be negative: -1000000000",
		},
		{
			"zero net liquidity",
			func(tickInfo *types.TickInfo) {
				tickInfo.NetLiquidity = sdk.NewInt(0)
			},
			"net liquidity must not be 0",
		},
		{
			"invalid fee growth outside",
			func(tickInfo *types.TickInfo) {
				tickInfo.FeeGrowthOutside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid fee growth outside: coin 0.000000000000000000ucre amount is not positive",
		},
		{
			"wrong fee growth outside coins number",
			func(tickInfo *types.TickInfo) {
				tickInfo.FeeGrowthOutside = utils.ParseDecCoins("0.0001ucre,0.0001uusd,0.0001uatom")
			},
			"number of coins in fee growth outside must not be higher than 2: 3",
		},
		{
			"invalid farming rewards growth outside",
			func(tickInfo *types.TickInfo) {
				tickInfo.FarmingRewardsGrowthOutside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid farming rewards growth outside: coin 0.000000000000000000ucre amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tickInfo := types.TickInfo{
				GrossLiquidity:              sdk.NewInt(1000_000000),
				NetLiquidity:                sdk.NewInt(-1000_000000),
				FeeGrowthOutside:            utils.ParseDecCoins("0.0001ucre,0.0001uusd"),
				FarmingRewardsGrowthOutside: utils.ParseDecCoins("0.0001stake"),
			}
			tc.malleate(&tickInfo)
			err := tickInfo.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
