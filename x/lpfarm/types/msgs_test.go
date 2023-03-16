package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

func TestMsgFarm(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgFarm)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgFarm) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgFarm) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"negative coin",
			func(msg *types.MsgFarm) {
				msg.Coin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid coin: negative coin amount: -1: invalid request",
		},
		{
			"zero coin",
			func(msg *types.MsgFarm) {
				msg.Coin = utils.ParseCoin("0pool1")
			},
			"non-positive coin: 0pool1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgFarm(utils.TestAddress(0), utils.ParseCoin("1000_000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgFarm, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgUnfarm(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgUnfarm)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgUnfarm) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgUnfarm) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"negative bond",
			func(msg *types.MsgUnfarm) {
				msg.Coin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid coin: negative coin amount: -1: invalid request",
		},
		{
			"zero bond",
			func(msg *types.MsgUnfarm) {
				msg.Coin = utils.ParseCoin("0pool1")
			},
			"non-positive coin: 0pool1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgUnfarm(utils.TestAddress(0), utils.ParseCoin("1000_000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgUnfarm, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgHarvest(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgHarvest)
		expectedErr string // empty means no error
	}{
		{
			"happy case",
			func(msg *types.MsgHarvest) {},
			"",
		},
		{
			"invalid farmer",
			func(msg *types.MsgHarvest) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid denom",
			func(msg *types.MsgHarvest) {
				msg.Denom = "invalid!"
			},
			"invalid denom: invalid denom: invalid!: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgHarvest(utils.TestAddress(0), "pool1")
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgHarvest, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmerAddress(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
