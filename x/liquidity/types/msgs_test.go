package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func TestMsgCreatePair(t *testing.T) {
	for _, tc := range []struct {
		malleate    func(msg *types.MsgCreatePair)
		expectedErr string
	}{
		{
			func(msg *types.MsgCreatePair) {},
			"",
		},
		{
			func(msg *types.MsgCreatePair) {
				msg.Creator = "invalidaddr"
			},
			"invalid creator address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			func(msg *types.MsgCreatePair) {
				msg.BaseCoinDenom = "invaliddenom!"
			},
			"invalid denom: invaliddenom!: invalid request",
		},
		{
			func(msg *types.MsgCreatePair) {
				msg.QuoteCoinDenom = "invaliddenom!"
			},
			"invalid denom: invaliddenom!: invalid request",
		},
	} {
		t.Run("", func(t *testing.T) {
			msg := types.NewMsgCreatePair(sdk.AccAddress(crypto.AddressHash([]byte("creator"))), "denom1", "denom2")
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCreatePair, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetCreator(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCreatePool(t *testing.T) {
	for _, tc := range []struct {
		malleate    func(msg *types.MsgCreatePool)
		expectedErr string
	}{
		{
			func(msg *types.MsgCreatePool) {},
			"", // empty means no error expected
		},
		{
			func(msg *types.MsgCreatePool) {
				msg.PairId = 0
			},
			"pair id must not be 0: invalid request",
		},
		{
			func(msg *types.MsgCreatePool) {
				msg.Creator = "invalidaddr"
			},
			"invalid creator address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			func(msg *types.MsgCreatePool) {
				msg.DepositCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 0), sdk.NewInt64Coin("denom2", 1000000)}
			},
			"coin 0denom1 amount is not positive",
		},
		{
			func(msg *types.MsgCreatePool) {
				msg.DepositCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 1000000), sdk.NewInt64Coin("denom2", 0)}
			},
			"coin denom2 amount is not positive",
		},
	} {
		t.Run("", func(t *testing.T) {
			msg := types.NewMsgCreatePool(testAddr, 1, parseCoins("1000000denom1,1000000denom2"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCreatePool, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetCreator(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgDepositBatch(t *testing.T) {
	testCases := []struct {
		malleate    func(msg *types.MsgDepositBatch)
		expectedErr string
	}{
		{
			func(msg *types.MsgDepositBatch) {},
			"", // empty means no error expected
		},
		{
			func(msg *types.MsgDepositBatch) {
				msg.Depositor = "invalidaddr"
			},
			"invalid depositor address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			func(msg *types.MsgDepositBatch) {
				msg.PoolId = 0
			},
			"pool id must not be 0: invalid request",
		},
		{
			func(msg *types.MsgDepositBatch) {
				msg.DepositCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 0), sdk.NewInt64Coin("denom2", 1000000)}
			},
			"coin 0denom1 amount is not positive",
		},
		{
			func(msg *types.MsgDepositBatch) {
				msg.DepositCoins = sdk.Coins{sdk.NewInt64Coin("denom1", 1000000), sdk.NewInt64Coin("denom2", 0)}
			},
			"coin denom2 amount is not positive",
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			msg := types.NewMsgDepositBatch(testAddr, 1, parseCoins("1000000denom1,1000000denom2"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgDepositBatch, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())

			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetDepositor(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgWithdrawBatch(t *testing.T) {
	testCases := []struct {
		expErr string
		msg    *types.MsgWithdrawBatch
	}{
		{
			"", // empty means no error expected
			types.NewMsgWithdrawBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Withdrawer"))),
				1,
				sdk.NewInt64Coin("PoolCoinDenom", 500_000),
			),
		},
		{
			"invalid withdrawer address: empty address string is not allowed: invalid address",
			types.NewMsgWithdrawBatch(
				sdk.AccAddress{},
				1,
				sdk.NewInt64Coin("PoolCoinDenom", 500_000),
			),
		},
		{
			"pool coin must be positive: invalid request",
			types.NewMsgWithdrawBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Withdrawer"))),
				1,
				sdk.NewInt64Coin("PoolCoinDenom", 0),
			),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, types.TypeMsgWithdrawBatch, tc.msg.Type())
			require.Equal(t, types.RouterKey, tc.msg.Route())

			err := tc.msg.ValidateBasic()
			if tc.expErr == "" {
				require.NoError(t, err)
				signers := tc.msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.msg.GetWithdrawer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expErr)
			}
		})
	}
}

func TestMsgMarketOrderBatch(t *testing.T) {
	orderLifespan := 20 * time.Second

	testCases := []struct {
		expErr string
		msg    *types.MsgMarketOrderBatch
	}{
		{
			"", // empty means no error expected
			types.NewMsgMarketOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom2",
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"invalid orderer address: empty address string is not allowed: invalid address",
			types.NewMsgMarketOrderBatch(
				sdk.AccAddress{},
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom2",
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"offer coin must be positive: invalid request",
			types.NewMsgMarketOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 0),
				"denom2",
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"offer coin denom and demand coin denom must not be same: invalid request",
			types.NewMsgMarketOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom1",
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, types.TypeMsgMarketOrderBatch, tc.msg.Type())
			require.Equal(t, types.RouterKey, tc.msg.Route())

			err := tc.msg.ValidateBasic()
			if tc.expErr == "" {
				require.NoError(t, err)
				signers := tc.msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expErr)
			}
		})
	}
}

func TestMsgLimitOrderBatch(t *testing.T) {
	orderLifespan := 20 * time.Second

	testCases := []struct {
		expErr string
		msg    *types.MsgLimitOrderBatch
	}{
		{
			"", // empty means no error expected
			types.NewMsgLimitOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom2",
				parseDec("1.0"),
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"invalid orderer address: empty address string is not allowed: invalid address",
			types.NewMsgLimitOrderBatch(
				sdk.AccAddress{},
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom2",
				parseDec("1.0"),
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"offer coin must be positive: invalid request",
			types.NewMsgLimitOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 0),
				"denom2",
				parseDec("1.0"),
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
		{
			"offer coin denom and demand coin denom must not be same: invalid request",
			types.NewMsgLimitOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				types.SwapDirectionBuy,
				sdk.NewInt64Coin("denom1", 100_000_000),
				"denom1",
				parseDec("1.0"),
				sdk.NewInt(100_000_000),
				orderLifespan,
			),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, types.TypeMsgLimitOrderBatch, tc.msg.Type())
			require.Equal(t, types.RouterKey, tc.msg.Route())

			err := tc.msg.ValidateBasic()
			if tc.expErr == "" {
				require.NoError(t, err)
				signers := tc.msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expErr)
			}
		})
	}
}

func TestMsgCancelOrderBatch(t *testing.T) {
	testCases := []struct {
		expErr string
		msg    *types.MsgCancelOrderBatch
	}{
		{
			"", // empty means no error expected
			types.NewMsgCancelOrderBatch(
				sdk.AccAddress(crypto.AddressHash([]byte("Orderer"))),
				1,
				1,
			),
		},
		{
			"invalid orderer address: empty address string is not allowed: invalid address",
			types.NewMsgCancelOrderBatch(
				sdk.AccAddress{},
				1,
				1,
			),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, types.TypeMsgCancelOrderBatch, tc.msg.Type())
			require.Equal(t, types.RouterKey, tc.msg.Route())

			err := tc.msg.ValidateBasic()
			if tc.expErr == "" {
				require.NoError(t, err)
				signers := tc.msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expErr)
			}
		})
	}
}
