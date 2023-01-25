package types_test

//import (
//	"testing"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/stretchr/testify/require"
//	"github.com/tendermint/tendermint/crypto"
//
//	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
//)
//
//func TestMsgApplyBootstrap(t *testing.T) {
//	mmAddr := sdk.AccAddress(crypto.AddressHash([]byte("mmAddr")))
//	pairId := []uint64{1}
//	pairIds := []uint64{1, 2}
//	pairIdsDuplicated := []uint64{1, 1, 2}
//	emptyPairIds := []uint64{}
//	hugePairIds := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
//
//	testCases := []struct {
//		expectedErr string
//		msg         *types.MsgLimitOrder
//	}{
//		{
//			"", // empty means no error expected
//			types.NewMsgLimitOrder(mmAddr, pairId),
//		},
//		{
//			"", // empty means no error expected
//			types.NewMsgApplyBootstrap(mmAddr, pairIds),
//		},
//		{
//			"", // empty means no error expected
//			types.NewMsgApplyBootstrap(mmAddr, hugePairIds),
//		},
//		{
//			"pair ids must not be empty: invalid request",
//			types.NewMsgApplyBootstrap(mmAddr, emptyPairIds),
//		},
//		{
//			"duplicated pair id 1: invalid pair id",
//			types.NewMsgApplyBootstrap(mmAddr, pairIdsDuplicated),
//		},
//		{
//			"invalid address \"invalidaddr\": decoding bech32 failed: invalid separator index -1: invalid address",
//			&types.MsgApplyBootstrap{
//				Address: "invalidaddr",
//				PairIds: pairIds,
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		require.IsType(t, &types.MsgApplyBootstrap{}, tc.msg)
//		require.Equal(t, types.TypeMsgApplyBootstrap, tc.msg.Type())
//		require.Equal(t, types.RouterKey, tc.msg.Route())
//		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())
//
//		err := tc.msg.ValidateBasic()
//		if tc.expectedErr == "" {
//			require.Nil(t, err)
//			signers := tc.msg.GetSigners()
//			require.Len(t, signers, 1)
//			require.Equal(t, tc.msg.GetAddress(), signers[0])
//		} else {
//			require.EqualError(t, err, tc.expectedErr)
//		}
//	}
//}
//
//func TestMsgClaimIncentives(t *testing.T) {
//	mmAddr := sdk.AccAddress(crypto.AddressHash([]byte("mmAddr")))
//
//	testCases := []struct {
//		expectedErr string
//		msg         *types.MsgClaimIncentives
//	}{
//		{
//			"", // empty means no error expected
//			types.NewMsgClaimIncentives(mmAddr),
//		},
//		{
//			"invalid address \"invalidaddr\": decoding bech32 failed: invalid separator index -1: invalid address",
//			&types.MsgClaimIncentives{
//				Address: "invalidaddr",
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		require.IsType(t, &types.MsgClaimIncentives{}, tc.msg)
//		require.Equal(t, types.TypeMsgClaimIncentives, tc.msg.Type())
//		require.Equal(t, types.RouterKey, tc.msg.Route())
//		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())
//
//		err := tc.msg.ValidateBasic()
//		if tc.expectedErr == "" {
//			require.Nil(t, err)
//			signers := tc.msg.GetSigners()
//			require.Len(t, signers, 1)
//			require.Equal(t, tc.msg.GetAddress(), signers[0])
//		} else {
//			require.EqualError(t, err, tc.expectedErr)
//		}
//	}
//}
