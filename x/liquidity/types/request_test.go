package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
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
				req.DepositCoins = utils.ParseCoins("1000000denom1,1000000denom2,1000000denom3")
			},
			"wrong number of deposit coins: 3",
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
				req.AcceptedCoins = utils.ParseCoins("1000000denom1,1000000denom2,1000000denom3")
			},
			"wrong number of accepted coins: 3",
		},
		{
			"wrong denom pair",
			func(req *types.DepositRequest) {
				req.DepositCoins = utils.ParseCoins("1000000denom1,1000000denom2")
				req.AcceptedCoins = utils.ParseCoins("1000000denom2,1000000denom3")
			},
			"mismatching denom pair between deposit coins and accepted coins",
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
				req.MintedPoolCoin = utils.ParseCoin("0pool1")
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
			pool := types.NewBasicPool(1, 1, testAddr)
			depositor := sdk.AccAddress(crypto.AddressHash([]byte("depositor")))
			msg := types.NewMsgDeposit(depositor, 1, utils.ParseCoins("1000000denom1,1000000denom2"))
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
				req.PoolCoin = utils.ParseCoin("0pool1")
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
				req.WithdrawnCoins = utils.ParseCoins("1000000denom1")
			},
			"",
		},
		{
			"wrong number of withdrawn coins",
			func(req *types.WithdrawRequest) {
				req.WithdrawnCoins = utils.ParseCoins("100000denom1,1000000denom2,1000000denom3")
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
			msg := types.NewMsgWithdraw(withdrawer, 1, utils.ParseCoin("1000pool1"))
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

func TestOrder_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(order *types.Order)
		expectedErr string
	}{
		{
			"happy case",
			func(order *types.Order) {},
			"",
		},
		{
			"zero id",
			func(order *types.Order) {
				order.Id = 0
			},
			"id must not be 0",
		},
		{
			"zero pair id",
			func(order *types.Order) {
				order.PairId = 0
			},
			"pair id must not be 0",
		},
		{
			"zero message height",
			func(order *types.Order) {
				order.MsgHeight = 0
			},
			"message height must not be 0",
		},
		{
			"invalid orderer addr",
			func(order *types.Order) {
				order.Orderer = "invalidaddr"
			},
			"invalid orderer address invalidaddr: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid direction",
			func(order *types.Order) {
				order.Direction = 10
			},
			"invalid direction: 10",
		},
		{
			"invalid offer coin",
			func(order *types.Order) {
				order.OfferCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid offer coin -1denom1: negative coin amount: -1",
		},
		{
			"zero offer coin",
			func(order *types.Order) {
				order.OfferCoin = utils.ParseCoin("0denom1")
			},
			"offer coin must not be 0",
		},
		{
			"invalid remaining offer coin",
			func(order *types.Order) {
				order.RemainingOfferCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid remaining offer coin -1denom1: negative coin amount: -1",
		},
		{
			"zero remaining offer coin",
			func(order *types.Order) {
				order.RemainingOfferCoin = utils.ParseCoin("0denom2")
			},
			"",
		},
		{
			"mismatching denom pair",
			func(order *types.Order) {
				order.OfferCoin = utils.ParseCoin("1000000denom1")
				order.RemainingOfferCoin = utils.ParseCoin("1000000denom2")
			},
			"offer coin denom denom1 != remaining offer coin denom denom2",
		},
		{
			"invalid received coin",
			func(order *types.Order) {
				order.ReceivedCoin = sdk.Coin{Denom: "denom1", Amount: sdk.NewInt(-1)}
			},
			"invalid received coin -1denom1: negative coin amount: -1",
		},
		{
			"zero received coin",
			func(order *types.Order) {
				order.ReceivedCoin = utils.ParseCoin("0denom1")
			},
			"",
		},
		{
			"zero price",
			func(order *types.Order) {
				order.Price = sdk.ZeroDec()
			},
			"price must be positive: 0.000000000000000000",
		},
		{
			"zero amount",
			func(order *types.Order) {
				order.Amount = sdk.ZeroInt()
			},
			"amount must be positive: 0",
		},
		{
			"negative open amount",
			func(order *types.Order) {
				order.OpenAmount = sdk.NewInt(-1)
			},
			"open amount must not be negative: -1",
		},
		{
			"zero batch id",
			func(order *types.Order) {
				order.BatchId = 0
			},
			"batch id must not be 0",
		},
		{
			"no expiration info",
			func(order *types.Order) {
				order.ExpireAt = time.Time{}
			},
			"no expiration info",
		},
		{
			"invalid status",
			func(order *types.Order) {
				order.Status = 10
			},
			"invalid status: 10",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pair := types.NewPair(1, "denom1", "denom2")
			orderer := sdk.AccAddress(crypto.AddressHash([]byte("orderer")))
			msg := types.NewMsgLimitOrder(
				orderer, pair.Id, types.OrderDirectionBuy, utils.ParseCoin("1000000denom2"),
				"denom1", utils.ParseDec("1.0"), newInt(1000000), types.DefaultMaxOrderLifespan)
			expireAt := utils.ParseTime("2022-01-01T00:00:00Z")
			order := types.NewOrderForLimitOrder(msg, 1, pair, utils.ParseCoin("1000000denom2"), msg.Price, expireAt, 1)
			tc.malleate(&order)
			err := order.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
