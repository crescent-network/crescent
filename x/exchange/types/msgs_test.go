package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestMsgCreateMarket(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCreateMarket)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCreateMarket) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCreateMarket) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid base denom",
			func(msg *types.MsgCreateMarket) {
				msg.BaseDenom = "invaliddenom!"
			},
			"invalid base denom: invalid denom: invaliddenom!: invalid request",
		},
		{
			"invalid quote denom",
			func(msg *types.MsgCreateMarket) {
				msg.QuoteDenom = "invaliddenom!"
			},
			"invalid quote denom: invalid denom: invaliddenom!: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCreateMarket(utils.TestAddress(1), "ucre", "uusd")
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgCreateMarket, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Quantity = sdk.NewInt(0)
			},
			"quantity must be positive: 0: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Quantity = sdk.NewInt(-1000000)
			},
			"quantity must be positive: -1000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgPlaceLimitOrder(
				utils.TestAddress(1), 1, true, utils.ParseDec("12.345"), sdk.NewInt(1000000), false, time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgPlaceLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceMarketOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceMarketOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceMarketOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Quantity = sdk.NewInt(0)
			},
			"quantity must be positive: 0: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Quantity = sdk.NewInt(-1000000)
			},
			"quantity must be positive: -1000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgPlaceMarketOrder(
				utils.TestAddress(1), 1, true, sdk.NewInt(1000000))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgPlaceMarketOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceMMLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceMMLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceMMLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Quantity = sdk.NewInt(0)
			},
			"quantity must be positive: 0: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Quantity = sdk.NewInt(-1000000)
			},
			"quantity must be positive: -1000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgPlaceMMLimitOrder(
				utils.TestAddress(1), 1, true, utils.ParseDec("12.345"), sdk.NewInt(1000000), false, time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgPlaceMMLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
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
			"valid",
			func(msg *types.MsgCancelOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCancelOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid order id",
			func(msg *types.MsgCancelOrder) {
				msg.OrderId = 0
			},
			"order id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgCancelOrder(utils.TestAddress(1), 1)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgCancelOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgSwapExactAmountIn(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgSwapExactAmountIn)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgSwapExactAmountIn) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"empty routes",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Routes = []uint64{}
			},
			"routes must not be empty: invalid request",
		},
		{
			"invalid market id",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Routes = []uint64{0, 1}
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero input",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Input = utils.ParseCoin("0ucre")
			},
			"input must be positive: 0ucre: invalid coins",
		},
		{
			"negative input",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Input = sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-1000000)}
			},
			"invalid input: negative coin amount: -1000000: invalid coins",
		},
		{
			"zero min output",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.MinOutput = utils.ParseCoin("0uusd")
			},
			"",
		},
		{
			"negative min output",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.MinOutput = sdk.Coin{Denom: "uusd", Amount: sdk.NewInt(-5000000)}
			},
			"invalid min output: negative coin amount: -5000000: invalid coins",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgSwapExactAmountIn(
				utils.TestAddress(1), []uint64{1, 2, 3},
				utils.ParseCoin("1000000ucre"), utils.ParseCoin("5000000uusd"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.TypeMsgSwapExactAmountIn, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
