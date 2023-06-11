package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(genState *types.GenesisState)
		expectedErr string
	}{
		{
			"valid",
			func(genState *types.GenesisState) {},
			"",
		},
		{
			"invalid tick info record pool id",
			func(genState *types.GenesisState) {
				genState.TickInfoRecords = []types.TickInfoRecord{
					{
						PoolId: 0,
						Tick:   100,
						TickInfo: types.TickInfo{
							GrossLiquidity:              sdk.NewInt(1000_000000),
							NetLiquidity:                sdk.NewInt(1000_000000),
							FeeGrowthOutside:            nil,
							FarmingRewardsGrowthOutside: nil,
						},
					},
				}
			},
			"invalid tick info record: pool id must not be 0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesis()
			tc.malleate(genState)
			err := genState.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
