package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	squad "github.com/cosmosquad-labs/squad/types"
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
				req.DepositCoins = squad.ParseCoins("1000000denom1")
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
				req.AcceptedCoins = squad.ParseCoins("1000000denom1")
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
				req.MintedPoolCoin = squad.ParseCoin("0pool1")
			},
			"",
		},
		{
			"invalid status",
			func(req *types.DepositRequest) {
				req.Status = 10
			},
			"invalid status: 10",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := types.NewPool(1, 1)
			depositor := sdk.AccAddress(crypto.AddressHash([]byte("depositor")))
			msg := types.NewMsgDeposit(depositor, 1, squad.ParseCoins("1000000denom1,1000000denom2"))
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
				req.PoolCoin = squad.ParseCoin("0pool1")
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
			"valid withdrawn coins",
			func(req *types.WithdrawRequest) {
				req.WithdrawnCoins = squad.ParseCoins("1000000denom1")
			},
			"",
		},
		{
			"wrong number of withdrawn coins",
			func(req *types.WithdrawRequest) {
				req.WithdrawnCoins = squad.ParseCoins("100000denom1,1000000denom2,1000000denom3")
			},
			"wrong number of withdrawn coins: 3",
		},
		{
			"invalid status",
			func(req *types.WithdrawRequest) {
				req.Status = 10
			},
			"invalid status: 10",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withdrawer := sdk.AccAddress(crypto.AddressHash([]byte("withdrawer")))
			msg := types.NewMsgWithdraw(withdrawer, 1, squad.ParseCoin("1000pool1"))
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
	for _, tc := range []struct {
		name        string
		malleate    func(req *types.SwapRequest)
		expectedErr string
	}{
		{
			"happy case",
			func(req *types.SwapRequest) {},
			"",
		},
		{
			"zero id",
			func(req *types.SwapRequest) {
				req.Id = 0
			},
			"id must not be 0",
		},
		{
			"zero pair id",
			func(req *types.SwapRequest) {
				req.PairId = 0
			},
			"pair id must not be 0",
		},
		{
			"zero message height",
			func(req *types.SwapRequest) {
				req.MsgHeight = 0
			},
			"message height must not be 0",
		},
		{
			"invalid orderer addr",
			func(req *types.SwapRequest) {
				req.Orderer = "invalidaddr"
			},
			"invalid orderer address invalidaddr: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid direction",
			func(req *types.SwapRequest) {
				req.Direction = 10
			},
			"invalid direction: 10",
		},
		{
			"invalid offer coin",
			func(req *types.SwapRequest) {
				req.OfferCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid offer coin -1denom1: negative coin amount: -1",
		},
		{
			"zero offer coin",
			func(req *types.SwapRequest) {
				req.OfferCoin = squad.ParseCoin("0denom1")
			},
			"offer coin must not be 0",
		},
		{
			"invalid remaining offer coin",
			func(req *types.SwapRequest) {
				req.RemainingOfferCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid remaining offer coin -1denom1: negative coin amount: -1",
		},
		{
			"zero remaining offer coin",
			func(req *types.SwapRequest) {
				req.RemainingOfferCoin = squad.ParseCoin("0denom1")
			},
			"",
		},
		{
			"invalid received coin",
			func(req *types.SwapRequest) {
				req.ReceivedCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid received coin -1denom1: negative coin amount: -1",
		},
		{
			"zero received coin",
			func(req *types.SwapRequest) {
				req.ReceivedCoin = squad.ParseCoin("0denom1")
			},
			"",
		},
		{
			"zero price",
			func(req *types.SwapRequest) {
				req.Price = sdk.ZeroDec()
			},
			"price must be positive: 0.000000000000000000",
		},
		{
			"zero amount",
			func(req *types.SwapRequest) {
				req.Amount = sdk.ZeroInt()
			},
			"amount must be positive: 0",
		},
		{
			"negative open amount",
			func(req *types.SwapRequest) {
				req.OpenAmount = sdk.NewInt(-1)
			},
			"open amount must not be negative: -1",
		},
		{
			"zero batch id",
			func(req *types.SwapRequest) {
				req.BatchId = 0
			},
			"batch id must not be 0",
		},
		{
			"no expiration info",
			func(req *types.SwapRequest) {
				req.ExpireAt = time.Time{}
			},
			"no expiration info",
		},
		{
			"invalid status",
			func(req *types.SwapRequest) {
				req.Status = 10
			},
			"invalid status: 10",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pair := types.NewPair(1, "denom1", "denom2")
			orderer := sdk.AccAddress(crypto.AddressHash([]byte("orderer")))
			msg := types.NewMsgLimitOrder(
				orderer, pair.Id, types.SwapDirectionBuy, squad.ParseCoin("1000000denom2"),
				"denom1", squad.ParseDec("1.0"), newInt(1000000), types.DefaultMaxOrderLifespan)
			expireAt := squad.ParseTime("2022-01-01T00:00:00Z")
			req := types.NewSwapRequestForLimitOrder(msg, 1, pair, squad.ParseCoin("1000000denom2"), msg.Price, expireAt, 1)
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
