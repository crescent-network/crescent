package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

func TestMsgClaim(t *testing.T) {
	var testAddr = sdk.AccAddress(crypto.AddressHash([]byte("test")))

	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgClaim)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgClaim) {},
			"",
		},
		{
			"invalid receipient",
			func(msg *types.MsgClaim) {
				msg.Recipient = "invalidaddr"
			},
			"invalid recipient address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid condition type",
			func(msg *types.MsgClaim) {
				msg.ConditionType = types.ConditionTypeUnspecified
			},
			"invalid condition type: CONDITION_TYPE_UNSPECIFIED: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgClaim(1, testAddr, types.ConditionTypeDeposit)
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgClaim, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetRecipient(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
