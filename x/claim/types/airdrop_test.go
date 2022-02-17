package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func TestClaimableCoinsForCondition(t *testing.T) {
	for _, tc := range []struct {
		name   string
		record types.ClaimRecord
	}{
		{
			"case #1",
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100_000_000))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100_000_000))),
				ClaimedConditions:     []bool{false, false, false},
			},
		},
		{
			"case #2",
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(999_999_999))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(999_999_999))),
				ClaimedConditions:     []bool{false, false, false},
			},
		},
		{
			"case #3",
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(666_777))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(666_777))),
				ClaimedConditions:     []bool{false, false, false},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claimableCoins := sdk.Coins{}

			for i := len(tc.record.ClaimedConditions); i > 0; i-- {
				amt := tc.record.GetClaimableCoinsForCondition(int64(i))
				tc.record.ClaimableCoins = tc.record.ClaimableCoins.Sub(amt)

				claimableCoins = claimableCoins.Add(amt...)
			}
			require.Equal(t, tc.record.InitialClaimableCoins, claimableCoins)
		})
	}
}
