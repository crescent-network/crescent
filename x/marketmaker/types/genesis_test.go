package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v5/x/marketmaker/types"
)

func TestValidateGenesis(t *testing.T) {
	mmAddr := sdk.AccAddress(crypto.AddressHash([]byte("mmAddr")))
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
			"empty deposit amount case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.DepositAmount = sdk.Coins{}
				genState.Params = params
			},
			"",
		},
		{
			"invalid param case",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.IncentiveBudgetAddress = "invalidaddr"
				genState.Params = params
			},
			"invalid account address: invalidaddr",
		},
		{
			"invalid market maker case",
			func(genState *types.GenesisState) {
				genState.MarketMakers = []types.MarketMaker{
					{
						Address: "invalidaddr",
						PairId:  1,
					},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid incentives",
			func(genState *types.GenesisState) {
				genState.Incentives = []types.Incentive{
					{
						Address:   "invalidaddr",
						Claimable: sdk.Coins{},
					},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid deposit records",
			func(genState *types.GenesisState) {
				genState.DepositRecords = []types.DepositRecord{
					{
						Address: "invalidaddr",
						PairId:  1,
						Amount:  sdk.Coins{},
					},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"empty deposit is valid",
			func(genState *types.GenesisState) {
				genState.MarketMakers = []types.MarketMaker{
					{
						Address:  mmAddr.String(),
						PairId:   1,
						Eligible: false,
					},
				}
				genState.DepositRecords = []types.DepositRecord{
					{
						Address: mmAddr.String(),
						PairId:  1,
						Amount:  sdk.Coins{},
					},
				}
			},
			"",
		},
		{
			"deposit record invariant fail 1",
			func(genState *types.GenesisState) {
				genState.MarketMakers = []types.MarketMaker{
					{
						Address:  mmAddr.String(),
						PairId:   1,
						Eligible: false,
					},
				}
			},
			"deposit invariant failed, not eligible market maker must have deposit record",
		},
		{
			"deposit record invariant fail 2",
			func(genState *types.GenesisState) {
				genState.DepositRecords = []types.DepositRecord{
					{
						Address: mmAddr.String(),
						PairId:  1,
						Amount:  sdk.Coins{},
					},
				}
			},
			"deposit invariant failed, deposit record's market maker must not be eligible",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			tc.configure(genState)

			err := types.ValidateGenesis(*genState)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
