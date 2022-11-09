package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v3/x/liquidstaking/types"
)

func TestMsgLiquidStake(t *testing.T) {
	delegatorAddr := sdk.AccAddress(crypto.AddressHash([]byte("delegatorAddr")))
	stakingCoin := sdk.NewCoin("token", sdk.NewInt(1))

	testCases := []struct {
		expectedErr string
		msg         *types.MsgLiquidStake
	}{
		{
			"", // empty means no error expected
			types.NewMsgLiquidStake(delegatorAddr, stakingCoin),
		},
		{
			"invalid delegator address \"\": empty address string is not allowed: invalid address",
			types.NewMsgLiquidStake(sdk.AccAddress{}, stakingCoin),
		},
		{
			"staking amount must not be zero: invalid request",
			types.NewMsgLiquidStake(delegatorAddr, sdk.NewCoin("token", sdk.NewInt(0))),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgLiquidStake{}, tc.msg)
		require.Equal(t, types.TypeMsgLiquidStake, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetDelegator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgLiquidUnstake(t *testing.T) {
	delegatorAddr := sdk.AccAddress(crypto.AddressHash([]byte("delegatorAddr")))
	stakingCoin := sdk.NewCoin("btoken", sdk.NewInt(1))

	testCases := []struct {
		expectedErr string
		msg         *types.MsgLiquidUnstake
	}{
		{
			"", // empty means no error expected
			types.NewMsgLiquidUnstake(delegatorAddr, stakingCoin),
		},
		{
			"invalid delegator address \"\": empty address string is not allowed: invalid address",
			types.NewMsgLiquidUnstake(sdk.AccAddress{}, stakingCoin),
		},
		{
			"unstaking amount must not be zero: invalid request",
			types.NewMsgLiquidUnstake(delegatorAddr, sdk.NewCoin("btoken", sdk.NewInt(0))),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgLiquidUnstake{}, tc.msg)
		require.Equal(t, types.TypeMsgLiquidUnstake, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetDelegator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}
