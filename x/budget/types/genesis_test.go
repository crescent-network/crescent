package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

func TestValidateGenesis(t *testing.T) {
	startTime, _ := time.Parse(time.RFC3339, "0000-01-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "9999-12-31T00:00:00Z")
	testCases := []struct {
		name        string
		configure   func(*types.GenesisState)
		expectedErr string
	}{
		{
			"default case",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
			},
			"",
		},
		{
			"normal budget case",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
				genState.Params.Budgets = []types.Budget{
					{
						Name:               "budget1",
						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
						StartTime:          startTime,
						EndTime:            endTime,
					},
				}
			},
			"",
		},
		{
			"invalid budget case",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
				genState.Params.Budgets = []types.Budget{
					{
						Name:               "budget1",
						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
						SourceAddress:      "cosmos1invalidaddress",
						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
						StartTime:          startTime,
						EndTime:            endTime,
					},
				}
			},
			"invalid source address cosmos1invalidaddress: decoding bech32 failed: invalid character not part of charset: 105: invalid address",
		},
		{
			"duplicate budget name",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
				genState.Params.Budgets = []types.Budget{
					{
						Name:               "budget1",
						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
						StartTime:          startTime,
						EndTime:            endTime,
					},
					{
						Name:               "budget1",
						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
						StartTime:          startTime,
						EndTime:            endTime,
					},
				}
			},
			"budget1: duplicate budget name",
		},
		{
			"invalid budget name case",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
				genState.BudgetRecords = []types.BudgetRecord{
					{
						Name:                "invalid name",
						TotalCollectedCoins: nil,
					},
				}
			},
			"invalid name: budget name only allows letters, digits, and dash(-) without spaces and the maximum length is 50",
		},
		{
			"invalid total_collected_coin case",
			func(genState *types.GenesisState) {
				genState.Params = types.DefaultParams()
				genState.BudgetRecords = []types.BudgetRecord{
					{
						Name:                "budget1",
						TotalCollectedCoins: sdk.Coins{sdk.NewCoin("stake", sdk.ZeroInt())},
					},
				}
			},
			"invalid total collected coins 0stake: coin 0stake amount is not positive: invalid coins",
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
