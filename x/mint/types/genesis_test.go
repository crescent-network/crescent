package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(genState *types.GenesisState)
		expectedErr string
	}{
		{
			"default is valid",
			func(genState *types.GenesisState) {},
			"",
		},
		{
			"valid last block time",
			func(genState *types.GenesisState) {
				tmpTime := utils.ParseTime("0001-01-01T00:00:00Z")
				genState.LastBlockTime = &tmpTime
			},
			"",
		},
		{
			"valid last block time2",
			func(genState *types.GenesisState) {
				tmpTime := utils.ParseTime("9999-12-31T00:00:00Z")
				genState.LastBlockTime = &tmpTime
			},
			"",
		},
		{
			"invalid last block time",
			func(genState *types.GenesisState) {
				tmpTime := time.Unix(-62136697901, 0)
				genState.LastBlockTime = &tmpTime
			},
			"invalid last block time",
		},
		{
			"empty inflation",
			func(genState *types.GenesisState) {
				genState.Params.InflationSchedules = nil
			},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.NewGenesisState(types.DefaultParams(), nil)
			tc.malleate(genState)
			err := types.ValidateGenesis(*genState)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
