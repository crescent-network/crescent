package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func TestDepositRequest_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(req *types.DepositRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.DepositRequest) {},
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
			"invalid deposit coins",
			func(req *types.DepositRequest) {
				req.DepositCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 0), sdk.NewInt64Coin("denom2", 1000000)}
			},
			"invalid deposit coins: coin 0denom1 amount is not positive",
		},
		{
			"wrong number of deposit coins",
			func(req *types.DepositRequest) {
				req.DepositCoins = sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000000))
			},
			"wrong number of deposit coins: 1",
		},
		{
			"invalid accepted coins",
			func(req *types.DepositRequest) {
				req.AcceptedCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 0), sdk.NewInt64Coin("denom2", 1000000)}
			},
			"invalid accepted coins: coin 0denom1 amount is not positive",
		},
		{
			"wrong number of accepted coins",
			func(req *types.DepositRequest) {
				req.AcceptedCoins = sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000000))
			},
			"wrong number of accepted coins: 1",
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
			pool := types.NewPool(1, 1)
			depositor := sdk.AccAddress(crypto.AddressHash([]byte("depositor")))
			msg := types.NewMsgDeposit(depositor, 1, parseCoins("1000000denom1,1000000denom2"))
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
	for _, tc := range []struct {
		name        string
		malleate    func(req *types.WithdrawRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.WithdrawRequest) {},
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
			"invalid withdrawn coins",
			func(req *types.WithdrawRequest) {
				req.WithdrawnCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 0)}
			},
			"invalid withdrawn coins: coin 0denom1 amount is not positive",
		},
		{
			"wrong number of withdrawn coins",
			func(req *types.WithdrawRequest) {
				req.WithdrawnCoins = sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000000))
			},
			"wrong number of withdrawn coins: 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withdrawer := sdk.AccAddress(crypto.AddressHash([]byte("withdrawer")))
			msg := types.NewMsgWithdraw(withdrawer, 1, parseCoin("1000pool1"))
			req := types.NewWithdrawRequest(msg, 1, 1)
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
