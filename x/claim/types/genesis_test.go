package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/claim/types"
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

						Id:            1,
						SourceAddress: sdk.AccAddress(crypto.AddressHash([]byte("sourceAddress"))).String(),
						Conditions: []types.ConditionType{
							types.ConditionTypeDeposit,
							types.ConditionTypeSwap,
							types.ConditionTypeLiquidStake,
							types.ConditionTypeVote,
						},
						StartTime: time.Now(),
						EndTime:   time.Now().AddDate(0, 1, 0),
					},
				},
				ClaimRecords: []types.ClaimRecord{
					{
						AirdropId:             1,
						Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient1"))).String(),
						InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimedConditions:     []types.ConditionType{},
					},
					{
						AirdropId:             1,
						Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient2"))).String(),
						InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(500_000_000_000))),
						ClaimedConditions:     []types.ConditionType{},
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
