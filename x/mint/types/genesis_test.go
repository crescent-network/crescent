package types_test

import (
	"testing"
	"time"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/mint/types"
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
				tmpTime := squadtypes.ParseTime("0001-01-01T00:00:00Z")
				genState.LastBlockTime = &tmpTime
			},
			"",
		},
		{
			"valid last block time2",
			func(genState *types.GenesisState) {
				tmpTime := squadtypes.ParseTime("9999-12-31T00:00:00Z")
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
			"invalid mint denom",
			func(genState *types.GenesisState) {
				genState.Params.MintDenom = ""
			},
			"mint denom cannot be blank",
		},
		{
			"invalid mint denom2",
			func(genState *types.GenesisState) {
				genState.Params.MintDenom = "a"
			},
			"invalid denom: a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
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
