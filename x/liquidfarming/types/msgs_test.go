package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestMsgMintShare(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgMintShare)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgMintShare) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgMintShare) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid liquid farm id",
			func(msg *types.MsgMintShare) {
				msg.LiquidFarmId = 0
			},
			"liquid farm id must not be 0: invalid request",
		},
		{
			"invalid desired amount",
			func(msg *types.MsgMintShare) {
				msg.DesiredAmount = sdk.Coins{utils.ParseCoin("0ucre")}
			},
			"invalid desired amount: coin 0ucre amount is not positive: invalid coins",
		},
		{
			"single asset",
			func(msg *types.MsgMintShare) {
				msg.DesiredAmount = utils.ParseCoins("100_000000ucre")
			},
			"",
		},
		{
			"invalid desired amount 2",
			func(msg *types.MsgMintShare) {
				msg.DesiredAmount = utils.ParseCoins("100_000000ucre,500_000000uusd,100_000000uatom")
			},
			"invalid desired amount length: 3: invalid request",
		},
		{
			"invalid desired amount 3",
			func(msg *types.MsgMintShare) {
				msg.DesiredAmount = sdk.Coins{}
			},
			"invalid desired amount length: 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgMintShare(utils.TestAddress(0), 1, utils.ParseCoins("100_000000ucre,500_000000uusd"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgMintShare, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.Sender, signers[0].String())
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgBurnShare(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgBurnShare)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgBurnShare) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgBurnShare) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid liquid farm id",
			func(msg *types.MsgBurnShare) {
				msg.LiquidFarmId = 0
			},
			"liquid farm id must not be 0: invalid request",
		},
		{
			"invalid share",
			func(msg *types.MsgBurnShare) {
				msg.Share = utils.ParseCoin("0lfshare1")
			},
			"share amount must be positive: 0lfshare1: invalid request",
		},
		{
			"invalid share denom",
			func(msg *types.MsgBurnShare) {
				msg.Share = utils.ParseCoin("10000lfshare2")
			},
			"share denom must be lfshare1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnShare(utils.TestAddress(0), 1, utils.ParseCoin("1000000lfshare1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgBurnShare, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.Sender, signers[0].String())
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceBid(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceBid)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgPlaceBid) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceBid) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid liquid farm id",
			func(msg *types.MsgPlaceBid) {
				msg.LiquidFarmId = 0
			},
			"liquid farm id must not be 0: invalid request",
		},
		{
			"invalid auction id",
			func(msg *types.MsgPlaceBid) {
				msg.RewardsAuctionId = 0
			},
			"rewards auction id must not be 0: invalid request",
		},
		{
			"invalid share",
			func(msg *types.MsgPlaceBid) {
				msg.Share = utils.ParseCoin("0lfshare1")
			},
			"share amount must be positive: 0lfshare1: invalid request",
		},
		{
			"invalid share denom",
			func(msg *types.MsgPlaceBid) {
				msg.Share = utils.ParseCoin("10000lfshare2")
			},
			"share denom must be lfshare1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgPlaceBid(utils.TestAddress(0), 1, 1, utils.ParseCoin("1000000lfshare1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgPlaceBid, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.Sender, signers[0].String())
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCancelBid(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCancelBid)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgCancelBid) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCancelBid) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid liquid farm id",
			func(msg *types.MsgCancelBid) {
				msg.LiquidFarmId = 0
			},
			"liquid farm id must not be 0: invalid request",
		},
		{
			"invalid auction id",
			func(msg *types.MsgCancelBid) {
				msg.RewardsAuctionId = 0
			},
			"rewards auction id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCancelBid(utils.TestAddress(0), 1, 1)
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCancelBid, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.Sender, signers[0].String())
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
