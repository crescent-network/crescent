package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

func TestClaimableCoinsForCondition(t *testing.T) {
	for _, tc := range []struct {
		name              string
		airdropConditions []types.ConditionType
		record            types.ClaimRecord
	}{
		{
			"case #1",
			[]types.ConditionType{
				types.ConditionTypeDeposit,
				types.ConditionTypeSwap,
				types.ConditionTypeLiquidStake,
				types.ConditionTypeVote,
			},
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100_000_000))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100_000_000))),
				ClaimedConditions:     []types.ConditionType{},
			},
		},
		{
			"case #2",
			[]types.ConditionType{
				types.ConditionTypeDeposit,
				types.ConditionTypeSwap,
				types.ConditionTypeLiquidStake,
				types.ConditionTypeVote,
			},
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(999_999_999))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(999_999_999))),
				ClaimedConditions:     []types.ConditionType{},
			},
		},
		{
			"case #3",
			[]types.ConditionType{
				types.ConditionTypeDeposit,
				types.ConditionTypeSwap,
				types.ConditionTypeLiquidStake,
				types.ConditionTypeVote,
			},
			types.ClaimRecord{
				AirdropId:             1,
				Recipient:             sdk.AccAddress(crypto.AddressHash([]byte("recipient"))).String(),
				InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(666_777))),
				ClaimableCoins:        sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(666_777))),
				ClaimedConditions:     []types.ConditionType{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < len(tc.airdropConditions); i++ {
				claimableCoins := tc.record.GetClaimableCoinsForCondition(tc.airdropConditions)

				tc.record.ClaimableCoins = tc.record.ClaimableCoins.Sub(claimableCoins)
				tc.record.ClaimedConditions = append(tc.record.ClaimedConditions, tc.airdropConditions[i])
			}
			require.True(t, tc.record.ClaimableCoins.IsZero())
			require.Equal(t, tc.airdropConditions, tc.record.ClaimedConditions)
		})
	}
}
