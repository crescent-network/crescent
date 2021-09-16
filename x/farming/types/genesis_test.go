package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/farming/x/farming/types"
)

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		name        string
		configure   func(*types.GenesisState)
		expectedErr string
	}{
		{
			"default case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				genState.Params = params
			},
			"",
		},
		{
			"invalid NextEpochDays case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.NextEpochDays = 0
				genState.Params = params
			},
			"next epoch days must be positive: 0",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			tc.configure(genState)

			err := types.ValidateGenesis(*genState)
			if tc.expectedErr == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
