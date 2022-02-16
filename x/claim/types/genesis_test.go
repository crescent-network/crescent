package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Airdrops: []types.Airdrop{
					{

						Id:                 1,
						SourceAddress:      types.SourceAddress(1).String(),
						SourceCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1_000_000_000_000))),
						TerminationAddress: "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta", // auth fee collector
						StartTime:          farmingtypes.ParseTime("2022-02-01T00:00:00Z"),
						EndTime:            farmingtypes.ParseTime("2022-06-01T00:00:00Z"),
					},
				},
				ClaimRecords: []types.ClaimRecord{
					{
						AirdropId:             1,
						Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient1"))).String(),
						InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						Actions: []types.Action{
							{
								ActionType: types.ActionTypeDeposit,
								Claimed:    false,
							},
							{
								ActionType: types.ActionTypeSwap,
								Claimed:    false,
							},
							{
								ActionType: types.ActionTypeFarming,
								Claimed:    false,
							},
						},
					},
					{
						AirdropId:             1,
						Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient2"))).String(),
						InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						Actions: []types.Action{
							{
								ActionType: types.ActionTypeDeposit,
								Claimed:    false,
							},
							{
								ActionType: types.ActionTypeSwap,
								Claimed:    false,
							},
							{
								ActionType: types.ActionTypeFarming,
								Claimed:    false,
							},
						},
					},
				},
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
