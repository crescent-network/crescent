package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestDepositRequest_Validate(t *testing.T) {
	for _, tc := range []struct{
		name string
		malleate func(req *types.DepositRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.DepositRequest) { },
			"",
		},
		{
			"zero id",
			func(req *types.DepositRequest) {
				req.Id = 0
			},
			"id must not be 0",
		},
		{
			"zero pool id",
			func(req *types.DepositRequest) {
				req.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"zero message height",
			func(req *types.DepositRequest) {
				req.MsgHeight = 0
			},
			"message height must not be 0",
		},
		{
			"invalid depositor addr",
			func(req *types.DepositRequest) {
				req.Depositor = "invalidaddr"
			},
			"invalid depositor address invalidaddr: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid x coin",
			func(req *types.DepositRequest) {
				req.XCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid x coin -1denom1: negative coin amount: -1",
		},
		{
			"zero x coin",
			func(req *types.DepositRequest) {
				req.XCoin = parseCoin("0denom1")
			},
			"x coin must not be 0",
		},
		{
			"invalid accepted x coin",
			func(req *types.DepositRequest) {
				req.AcceptedXCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid accepted x coin -1denom1: negative coin amount: -1",
		},
		{
			"zero accepted x coin",
			func(req *types.DepositRequest) {
				req.AcceptedXCoin = parseCoin("0denom1")
			},
			"",
		},
		{
			"invalid y coin",
			func(req *types.DepositRequest) {
				req.YCoin = sdk.Coin{Denom: "denom2", Amount: sdk.NewInt(-1)}
			},
			"invalid y coin -1denom2: negative coin amount: -1",
		},
		{
			"zero y coin",
			func(req *types.DepositRequest) {
				req.YCoin = parseCoin("0denom2")
			},
			"y coin must not be 0",
		},
		{
			"invalid accepted y coin",
			func(req *types.DepositRequest) {
				req.AcceptedYCoin = sdk.Coin{Denom: "denom2", Amount: sdk.NewInt(-1)}
			},
			"invalid accepted y coin -1denom2: negative coin amount: -1",
		},
		{
			"zero accepted y coin",
			func(req *types.DepositRequest) {
				req.AcceptedYCoin = parseCoin("0denom2")
			},
			"",
		},
		{
			"invalid minted pool coin",
			func(req *types.DepositRequest) {
				req.MintedPoolCoin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid minted pool coin -1pool1: negative coin amount: -1",
		},
		{
			"zero minted pool coin",
			func(req *types.DepositRequest) {
				req.MintedPoolCoin = parseCoin("0pool1")
			},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := types.NewPool(1, 1, "denom1", "denom2")
			depositor := sdk.AccAddress(crypto.AddressHash([]byte("depositor")))
			msg := types.NewMsgDepositBatch(depositor, 1, parseCoin("1000000denom1"), parseCoin("1000000denom2"))
			req := types.NewDepositRequest(msg, pool, 1, 1)
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestWithdrawRequest_Validate(t *testing.T) {
	for _, tc := range []struct{
		name string
		malleate func(req *types.WithdrawRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.WithdrawRequest) { },
			"",
		},
		{
			"zero id",
			func(req *types.WithdrawRequest) {
				req.Id = 0
			},
			"id must not be 0",
		},
		{
			"zero pool id",
			func(req *types.WithdrawRequest) {
				req.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"zero message height",
			func(req *types.WithdrawRequest) {
				req.MsgHeight = 0
			},
			"message height must not be 0",
		},
		{
			"invalid depositor addr",
			func(req *types.WithdrawRequest) {
				req.Withdrawer = "invalidaddr"
			},
			"invalid withdrawer address invalidaddr: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid pool coin",
			func(req *types.WithdrawRequest) {
				req.PoolCoin = sdk.Coin{Denom: "pool1", Amount: sdk.NewInt(-1)}
			},
			"invalid pool coin -1pool1: negative coin amount: -1",
		},
		{
			"zero pool coin",
			func(req *types.WithdrawRequest) {
				req.PoolCoin = parseCoin("0pool1")
			},
			"pool coin must not be 0",
		},
		{
			"invalid withdrawn x coin",
			func(req *types.WithdrawRequest) {
				req.WithdrawnXCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid withdrawn x coin -1denom1: negative coin amount: -1",
		},
		{
			"zero withdrawn x coin",
			func(req *types.WithdrawRequest) {
				req.WithdrawnXCoin = parseCoin("0denom1")
			},
			"",
		},
		{
			"invalid withdrawn y coin",
			func(req *types.WithdrawRequest) {
				req.WithdrawnYCoin = sdk.Coin{Denom: "denom2", Amount: sdk.NewInt(-1)}
			},
			"invalid withdrawn y coin -1denom2: negative coin amount: -1",
		},
		{
			"zero withdrawn y coin",
			func(req *types.WithdrawRequest) {
				req.WithdrawnYCoin = parseCoin("0denom2")
			},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := types.NewPool(1, 1, "denom1", "denom2")
			withdrawer := sdk.AccAddress(crypto.AddressHash([]byte("withdrawer")))
			msg := types.NewMsgWithdrawBatch(withdrawer, 1, parseCoin("1000pool1"))
			req := types.NewWithdrawRequest(msg, pool, 1, 1)
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestSwapRequest_Validate(t *testing.T) {
	// TODO: not implemented
}

func TestCancelSwapRequest_Validate(t *testing.T) {
	for _, tc := range []struct{
		name string
		malleate func(req *types.CancelSwapRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.CancelSwapRequest) { },
			"",
		},
		{
			"zero id",
			func(req *types.CancelSwapRequest) {
				req.Id = 0
			},
			"id must not be 0",
		},
		{
			"zero pair id",
			func(req *types.CancelSwapRequest) {
				req.PairId = 0
			},
			"pair id must not be 0",
		},
		{
			"zero message height",
			func(req *types.CancelSwapRequest) {
				req.MsgHeight = 0
			},
			"message height must not be 0",
		},
		{
			"invalid orderer addr",
			func(req *types.CancelSwapRequest) {
				req.Orderer = "invalidaddr"
			},
			"invalid orderer address invalidaddr: decoding bech32 failed: invalid separator index -1",
		},
		{
			"zero swap request id",
			func(req *types.CancelSwapRequest) {
				req.SwapRequestId = 0
			},
			"swap request id must not be 0",
		},
		{
			"zero batch id",
			func(req *types.CancelSwapRequest) {
				req.BatchId = 0
			},
			"batch id must not be 0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pair := types.NewPair(1,  "denom1", "denom2")
			orderer := sdk.AccAddress(crypto.AddressHash([]byte("orderer")))
			msg := types.NewMsgCancelSwapBatch(orderer, pair.Id, 1)
			req := types.NewCancelSwapRequest(msg, 1, pair, 1)
			tc.malleate(&req)
			err := req.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
