package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

var testAddr = sdk.AccAddress(crypto.AddressHash([]byte("test")))

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
			"invalid pool id",
			func(msg *types.MsgMintShare) {
				msg.PoolId = 0
			},
			"invalid pool id: invalid request",
		},
		{
			"invalid farmer",
			func(msg *types.MsgMintShare) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid farming coin",
			func(msg *types.MsgMintShare) {
				msg.PoolId = 1
				msg.FarmingCoin = sdk.NewInt64Coin("pool1", 0)
			},
			"farming coin must be positive: invalid request",
		},
		{
			"invalid farming coin denom",
			func(msg *types.MsgMintShare) {
				msg.PoolId = 1
				msg.FarmingCoin = sdk.NewInt64Coin("denom1", 100_000)
			},
			"expected denom pool1, but got denom1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgMintShare(1, testAddr.String(), utils.ParseCoin("1000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgMintShare, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmer(), signers[0])
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
			"invalid pool id",
			func(msg *types.MsgBurnShare) {
				msg.PoolId = 0
			},
			"invalid pool id: invalid request",
		},
		{
			"invalid farmer",
			func(msg *types.MsgBurnShare) {
				msg.Farmer = "invalidaddr"
			},
			"invalid farmer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid lf coin",
			func(msg *types.MsgBurnShare) {
				msg.UnfarmingCoin = sdk.NewInt64Coin("lf1", 0)
			},
			"unfarming coin must be positive: invalid request",
		},
		{
			"invalid lf coin denom",
			func(msg *types.MsgBurnShare) {
				msg.UnfarmingCoin = sdk.NewInt64Coin("pool1", 100_000)
			},
			"expected denom lf1, but got pool1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnShare(1, testAddr.String(), utils.ParseCoin("1000000lf1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgBurnShare, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetFarmer(), signers[0])
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
			"invalid auction id",
			func(msg *types.MsgPlaceBid) {
				msg.AuctionId = 0
			},
			"invalid auction id: invalid request",
		},
		{
			"invalid pool id",
			func(msg *types.MsgPlaceBid) {
				msg.PoolId = 0
			},
			"invalid pool id: invalid request",
		},
		{
			"invalid bidder",
			func(msg *types.MsgPlaceBid) {
				msg.Bidder = "invalidaddr"
			},
			"invalid bidder address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid bidding coin",
			func(msg *types.MsgPlaceBid) {
				msg.BiddingCoin = sdk.NewInt64Coin("pool1", 0)
			},
			"bidding amount must be positive: invalid request",
		},
		{
			"invalid bidding coin denom",
			func(msg *types.MsgPlaceBid) {
				msg.PoolId = 1
				msg.BiddingCoin = sdk.NewInt64Coin("denom1", 100_000)
			},
			"expected denom pool1, but got denom1: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgPlaceBid(1, 1, testAddr.String(), utils.ParseCoin("1000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgPlaceBid, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetBidder(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgRefundBid(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgRefundBid)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgRefundBid) {},
			"",
		},
		{
			"invalid auction id",
			func(msg *types.MsgRefundBid) {
				msg.AuctionId = 0
			},
			"invalid auction id: invalid request",
		},
		{
			"invalid pool id",
			func(msg *types.MsgRefundBid) {
				msg.PoolId = 0
			},
			"invalid pool id: invalid request",
		},
		{
			"invalid bidder",
			func(msg *types.MsgRefundBid) {
				msg.Bidder = "invalidaddr"
			},
			"invalid bidder address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgRefundBid(1, 1, testAddr.String())
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgRefundBid, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(msg)), msg.GetSignBytes())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetBidder(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
