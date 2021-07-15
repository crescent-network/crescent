package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
			"happy case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.PrivatePlanCreationFee = sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(100)))
				genState.Params = params
			},
			"",
		},
		// {
		// 	"invalid case",
		// 	func(genState *types.GenesisState) {
		// 		genState.PoolRecords = []types.PoolRecord{{}}
		// 	},
		// 	"bad msg index of the batch",
		// },
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
