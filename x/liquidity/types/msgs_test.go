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
		name        string
		malleate    func(msg *types.MsgCreatePair)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgCreatePair) {},
			"",
		},
		{
			"invalid creator",
			func(msg *types.MsgCreatePair) {
				msg.Creator = "invalidaddr"
			},
			"invalid creator address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid base coin denom",
			func(msg *types.MsgCreatePair) {
				msg.BaseCoinDenom = "invaliddenom!"
			},
			"invalid denom: invaliddenom!: invalid request",
		},
		{
			"invalid quote coin denom",
			func(msg *types.MsgCreatePair) {
				msg.QuoteCoinDenom = "invaliddenom!"
			},
			"invalid denom: invaliddenom!: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCreatePair(testAddr, "denom1", "denom2")
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
		name        string
		malleate    func(msg *types.MsgCreatePool)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgCreatePool) {},
			"", // empty means no error expected
		},
		{
			"invalid pair id",
			func(msg *types.MsgCreatePool) {
				msg.PairId = 0
			},
			"pair id must not be 0: invalid request",
		},
		{
			"invalid creator",
			func(msg *types.MsgCreatePool) {
				msg.Creator = "invalidaddr"
			},
			"invalid creator address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgCreatePool) {
				msg.DepositCoins = sdk.Coins{parseCoin("0denom1"), parseCoin("1000000denom2")}
			},
			"coin 0denom1 amount is not positive",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgCreatePool) {
				msg.DepositCoins = sdk.Coins{parseCoin("1000000denom1"), parseCoin("0denom2")}
			},
			"coin denom2 amount is not positive",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgCreatePool) {
				msg.DepositCoins = parseCoins("1000000denom1,1000000denom2,1000000denom3")
			},
			"wrong number of deposit coins: 3: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
		name        string
		malleate    func(msg *types.MsgDepositBatch)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgDepositBatch) {},
			"", // empty means no error expected
		},
		{
			"invalid depositor",
			func(msg *types.MsgDepositBatch) {
				msg.Depositor = "invalidaddr"
			},
			"invalid depositor address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pool id",
			func(msg *types.MsgDepositBatch) {
				msg.PoolId = 0
			},
			"pool id must not be 0: invalid request",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgDepositBatch) {
				msg.DepositCoins = sdk.Coins{parseCoin("0denom1"), parseCoin("1000000denom2")}
			},
			"coin 0denom1 amount is not positive",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgDepositBatch) {
				msg.DepositCoins = sdk.Coins{parseCoin("1000000denom1"), parseCoin("0denom2")}
			},
			"coin denom2 amount is not positive",
		},
		{
			"invalid deposit coins",
			func(msg *types.MsgDepositBatch) {
				msg.DepositCoins = parseCoins("1000000denom1,1000000denom2,1000000denom3")
			},
			"wrong number of deposit coins: 3: invalid request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgWithdrawBatch)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgWithdrawBatch) {},
			"", // empty means no error expected
		},
		{
			"invalid withdrawer",
			func(msg *types.MsgWithdrawBatch) {
				msg.Withdrawer = "invalidaddr"
			},
			"invalid withdrawer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pool id",
			func(msg *types.MsgWithdrawBatch) {
				msg.PoolId = 0
			},
			"pool id must not be 0: invalid request",
		},
		{
			"invalid pool coin",
			func(msg *types.MsgWithdrawBatch) {
				msg.PoolCoin = parseCoin("0pool1")
			},
			"pool coin must be positive: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgWithdrawBatch(testAddr, 1, parseCoin("1000000pool1"))
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgWithdrawBatch, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetWithdrawer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgLimitOrderBatch(t *testing.T) {
	orderLifespan := 20 * time.Second
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgLimitOrderBatch)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgLimitOrderBatch) {},
			"", // empty means no error expected
		},
		{
			"invalid orderer",
			func(msg *types.MsgLimitOrderBatch) {
				msg.Orderer = "invalidaddr"
			},
			"invalid orderer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pair id",
			func(msg *types.MsgLimitOrderBatch) {
				msg.PairId = 0
			},
			"pair id must not be 0: invalid request",
		},
		{
			"invalid direction",
			func(msg *types.MsgLimitOrderBatch) {
				msg.Direction = 0
			},
			"invalid swap direction: SWAP_DIRECTION_UNSPECIFIED: invalid request",
		},
		{
			"invalid offer coin",
			func(msg *types.MsgLimitOrderBatch) {
				msg.OfferCoin = parseCoin("0denom1")
			},
			"offer coin must be positive: invalid request",
		},
		{
			"insufficient offer coin amount",
			func(msg *types.MsgLimitOrderBatch) {
				msg.OfferCoin = parseCoin("10denom1")
			},
			"offer coin is less than minimum coin amount: invalid request",
		},
		{
			"invalid demand coin denom",
			func(msg *types.MsgLimitOrderBatch) {
				msg.DemandCoinDenom = "invaliddenom!"
			},
			"invalid demand coin denom: invalid denom: invaliddenom!",
		},
		{
			"same offer coin denom and demand coin denom",
			func(msg *types.MsgLimitOrderBatch) {
				msg.OfferCoin = parseCoin("1000000denom1")
				msg.DemandCoinDenom = "denom1"
			},
			"offer coin denom and demand coin denom must not be same: invalid request",
		},
		{
			"invalid price",
			func(msg *types.MsgLimitOrderBatch) {
				msg.Price = parseDec("0")
			},
			"price must be positive: invalid request",
		},
		{
			"invalid amount",
			func(msg *types.MsgLimitOrderBatch) {
				msg.Amount = sdk.ZeroInt()
			},
			"amount must be positive: 0: invalid request",
		},
		{
			"insufficient amount",
			func(msg *types.MsgLimitOrderBatch) {
				msg.Amount = newInt(10)
			},
			"base coin is less than minimum coin amount: invalid request",
		},
		{
			"invalid order lifespan",
			func(msg *types.MsgLimitOrderBatch) {
				msg.OrderLifespan = -1
			},
			"order lifespan must not be negative: -1ns: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgLimitOrderBatch(
				testAddr, 1, types.SwapDirectionSell, parseCoin("1000000denom2"),
				"denom1", parseDec("1.0"), newInt(1000000), orderLifespan)
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgLimitOrderBatch, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgMarketOrderBatch(t *testing.T) {
	orderLifespan := 20 * time.Second
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgMarketOrderBatch)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgMarketOrderBatch) {},
			"", // empty means no error expected
		},
		{
			"invalid orderer",
			func(msg *types.MsgMarketOrderBatch) {
				msg.Orderer = "invalidaddr"
			},
			"invalid orderer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pair id",
			func(msg *types.MsgMarketOrderBatch) {
				msg.PairId = 0
			},
			"pair id must not be 0: invalid request",
		},
		{
			"invalid direction",
			func(msg *types.MsgMarketOrderBatch) {
				msg.Direction = 0
			},
			"invalid swap direction: SWAP_DIRECTION_UNSPECIFIED: invalid request",
		},
		{
			"invalid offer coin",
			func(msg *types.MsgMarketOrderBatch) {
				msg.OfferCoin = parseCoin("0denom1")
			},
			"offer coin must be positive: invalid request",
		},
		{
			"insufficient offer coin amount",
			func(msg *types.MsgMarketOrderBatch) {
				msg.OfferCoin = parseCoin("10denom1")
			},
			"offer coin is less than minimum coin amount: invalid request",
		},
		{
			"invalid demand coin denom",
			func(msg *types.MsgMarketOrderBatch) {
				msg.DemandCoinDenom = "invaliddenom!"
			},
			"invalid demand coin denom: invalid denom: invaliddenom!",
		},
		{
			"same offer coin denom and demand coin denom",
			func(msg *types.MsgMarketOrderBatch) {
				msg.OfferCoin = parseCoin("1000000denom1")
				msg.DemandCoinDenom = "denom1"
			},
			"offer coin denom and demand coin denom must not be same: invalid request",
		},
		{
			"invalid amount",
			func(msg *types.MsgMarketOrderBatch) {
				msg.Amount = sdk.ZeroInt()
			},
			"amount must be positive: 0: invalid request",
		},
		{
			"insufficient amount",
			func(msg *types.MsgMarketOrderBatch) {
				msg.Amount = newInt(10)
			},
			"base coin is less than minimum coin amount: invalid request",
		},
		{
			"invalid order lifespan",
			func(msg *types.MsgMarketOrderBatch) {
				msg.OrderLifespan = -1
			},
			"order lifespan must not be negative: -1ns: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgMarketOrderBatch(
				testAddr, 1, types.SwapDirectionBuy, parseCoin("1000000denom1"),
				"denom2", newInt(1000000), orderLifespan)
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgMarketOrderBatch, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCancelOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCancelOrder)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgCancelOrder) {},
			"", // empty means no error expected
		},
		{
			"invalid orderer",
			func(msg *types.MsgCancelOrder) {
				msg.Orderer = "invalidaddr"
			},
			"invalid orderer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pair id",
			func(msg *types.MsgCancelOrder) {
				msg.PairId = 0
			},
			"pair id must not be 0: invalid request",
		},
		{
			"invalid swap request id",
			func(msg *types.MsgCancelOrder) {
				msg.SwapRequestId = 0
			},
			"swap request id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCancelOrder(testAddr, 1, 1)
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCancelOrder, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCancelAllOrders(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCancelAllOrders)
		expectedErr string
	}{
		{
			"happy case",
			func(msg *types.MsgCancelAllOrders) {},
			"",
		},
		{
			"invalid orderer",
			func(msg *types.MsgCancelAllOrders) {
				msg.Orderer = "invalidaddr"
			},
			"invalid orderer address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid pair ids",
			func(msg *types.MsgCancelAllOrders) {
				msg.PairIds = []uint64{0}
			},
			"pair id must not be 0: invalid request",
		},
		{
			"invalid pair ids",
			func(msg *types.MsgCancelAllOrders) {
				msg.PairIds = []uint64{1, 1}
			},
			"duplicate pair id presents in the pair id list",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCancelAllOrders(sdk.AccAddress(crypto.AddressHash([]byte("orderer"))), []uint64{1, 2, 3})
			tc.malleate(msg)
			require.Equal(t, types.TypeMsgCancelAllOrders, msg.Type())
			require.Equal(t, types.RouterKey, msg.Route())
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, msg.GetOrderer(), signers[0])
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
